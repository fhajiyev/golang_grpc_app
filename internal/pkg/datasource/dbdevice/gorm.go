package dbdevice

import (
	"fmt"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

// GormDB struct definition
type GormDB struct {
	db *gorm.DB
}

// GetByID returns device entity by deviceID.
func (r *GormDB) GetByID(deviceID int64) (*Device, error) {
	d := Device{ID: deviceID}
	err := r.db.Find(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

// GetByAppIDAndIFA returns device entity by appID and IFA.
func (r *GormDB) GetByAppIDAndIFA(appID int64, ifa string) (*Device, error) {
	d := Device{AppID: appID, IFA: ifa}
	err := r.db.Where(&d).Find(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

// GetByAppIDAndPubUserID returns device entity by appID and IFA.
func (r *GormDB) GetByAppIDAndPubUserID(appID int64, pubUserID string) (*Device, error) {
	d := Device{AppID: appID, UnitDeviceToken: pubUserID}
	err := r.db.Where(&d).Find(&d).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &d, nil
}

// UpsertDevice creates or updates device entity.
func (r *GormDB) UpsertDevice(req Device) (*Device, error) {
	// 1. <AppID, UnitDeviceToken> or <AppID, IFA> 로 device 조회
	record, err := r.findDeviceByKey(req.AppID, req.UnitDeviceToken, req.IFA)
	if err != nil {
		return nil, err
	} else if record == nil {
		// record가 없을경우 새로 생성
		return r.getOrCreateDevice(req)
	}

	// 2. IFA, UnitDeviceToken, PackageName에 대해 history record를 만들기 위한 정보 생성
	histories := record.updateWithHistory(req)

	// 3. TRANSACTION: device record 저장 시도
	// TODO migrate to accountsvc
	h, err := r.updateDeviceRecord(req, *record)
	if err != nil { // aborted transaction
		return nil, err
	} else if h != nil {
		histories = append(histories, *h)
	}

	// 4. TRANSACTION to bulk insert: history record 저장
	err = r.insertHistories(histories)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *GormDB) updateDeviceRecord(req Device, record Device) (*DeviceUpdateHistory, error) {
	var history *DeviceUpdateHistory

	err := r.dbTransact(func(tx *gorm.DB) error {
		err := tx.Save(&record).Error
		if err == nil {
			return nil
		}

		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			history, err = r.patchIFAForCollidedDeviceInTransaction(tx, req, record.ID)
			if err != nil {
				return err
			}

			// patch후 기존 레코드 다시 저장 시도
			err = tx.Save(&record).Error
			if err != nil {
				core.Logger.Errorf("updateDeviceRecord - save error %v %v %v %v", record.ID, record.IFA, record.UnitDeviceToken, err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return history, nil
}

/*
	이미 버즈스크린을 사용하던 디바이스에 다른 유저가 로그인 하는경우 이 상황이 발생한다
	예를들어 (device, unit_device_token, ifa) = (1, A, aaa)와 (2, B, bbb)가 있는데,
	(2, B, bbb)유저가 (1, A, aaa)가 가진 기기 aaa로 접속 할경우 (2, B, aaa)로 요청이 들어오게됨
	따라서 (2, B, bbb)를 (2, B, aaa)로 바꿔줘야하는데 (1, A, aaa)가 이미 있어 ifa가 충돌이 남
	이 경우 (1, A, aaa)의 ifa에 device_id suffix를 붙여서 충돌을 해결
	         BEFORE        REQUEST        AFTER
	DID     1       2         2        1        2
	UDT     A       B     ->  B  ->    A        B
	IFA    aaa     bbb       aaa    aaa_1_2    aaa
*/
func (r *GormDB) patchIFAForCollidedDeviceInTransaction(tx *gorm.DB, req Device, opponentDeviceID int64) (*DeviceUpdateHistory, error) {
	var oldDevice Device
	oldDeviceReq := Device{
		AppID: req.AppID,
		IFA:   req.IFA,
	}
	err := tx.Where(oldDeviceReq).Find(&oldDevice).Error
	if gorm.IsRecordNotFoundError(err) {
		// ifa로 레코드가 없을 경우 history를 만들지 않고 nil 리턴
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	ifaForOldDevice := fmt.Sprintf("%s_d_%v_%v", oldDevice.IFA, opponentDeviceID, oldDevice.ID)
	h := newDeviceHistory(oldDevice.ID, historyFieldNameIFA, oldDevice.IFA, ifaForOldDevice)

	oldDevice.IFA = ifaForOldDevice
	err = tx.Save(oldDevice).Error
	if err != nil {
		return nil, err
	}
	core.Logger.Infof("Device owner changed device_id:%d, app_id:%d from <%s, %s> to <%s, %s>", req.ID, req.AppID, oldDevice.IFA, oldDevice.UnitDeviceToken, req.IFA, req.UnitDeviceToken)

	return &h, nil
}

func (r *GormDB) insertHistories(histories []DeviceUpdateHistory) error {
	if len(histories) == 0 {
		return nil
	}

	return r.dbTransact(func(tx *gorm.DB) error {
		for _, h := range histories {
			err := tx.Create(&h).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormDB) getOrCreateDevice(d Device) (*Device, error) {
	d.CreatedAt = gorm.NowFunc()
	d.UpdatedAt = gorm.NowFunc()

	err := r.db.Create(&d).Error
	if err == nil {
		return &d, nil
	}

	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok && mysqlErr.Number == 1062 { // duplicated entry error
		return r.findDeviceByKey(d.AppID, d.UnitDeviceToken, d.IFA)
	}

	return nil, err
}

func (r *GormDB) findDeviceByKey(appID int64, unitDeviceToken string, ifa string) (*Device, error) {
	// unitDeviceToken이 대부분의 앱에서 변경하지 않기 때문에 먼저 조회
	d, _ := r.GetByAppIDAndPubUserID(appID, unitDeviceToken)
	if d != nil {
		return d, nil
	}

	d, err := r.GetByAppIDAndIFA(appID, ifa)
	if d != nil {
		return d, nil
	} else if err != nil {
		return nil, err
	}

	return nil, nil // record를 찾지 못함
}

func (r *GormDB) dbTransact(txFunc func(*gorm.DB) error) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	err := txFunc(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// NewSource returns new GormDB instance.
func NewSource(db *gorm.DB) *GormDB {
	return &GormDB{db}
}

var _ DBSource = &GormDB{}
