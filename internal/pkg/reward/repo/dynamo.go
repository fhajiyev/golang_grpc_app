package repo

import (
	"github.com/guregu/dynamo"
)

const maxRecordsPerHour int64 = 100

func (r *Repository) getImpressionPoints(deviceID int64, limit int64) []DBPoint {
	query := r.pointTable.Get(keyDeviceID, deviceID).Order(dynamo.Descending).Limit(limit).SearchLimit(maxRecordsPerHour)
	itr := query.Iter()
	more := true

	points := make([]DBPoint, 0)
	for i := 0; i < int(limit); i++ {
		p := DBPoint{}
		more = itr.Next(&p)
		if !more {
			break
		}

		if p.Type != pointTypeImpression {
			continue
		}

		points = append(points, p)
	}

	return points
}
