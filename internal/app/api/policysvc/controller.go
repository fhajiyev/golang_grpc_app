package policysvc

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common/code"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/policysvc/dto"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/app"
	"github.com/Buzzvil/buzzscreen-api/internal/pkg/location"
)

// Controller struct definition
type Controller struct {
	*common.ControllerBase
	AppUseCase app.UseCase
	LocUseCase location.UseCase
}

// NewController returns new controller and binds requests to the controller
func NewController(e *core.Engine, appUseCase app.UseCase, locUseCase location.UseCase) Controller {
	con := Controller{AppUseCase: appUseCase, LocUseCase: locUseCase}
	e.GET("/api/policies/privacy", con.GetPrivacyPolicy)
	e.POST("/api/policies/privacy", con.PostPrivacyPolicy)
	e.GET("/policies/:lang/:name/terms", con.GetTerms)
	e.GET("/policies/:lang/:name/privacy", con.GetPrivacy)

	e.GET("/api/v2/policies/privacy", con.GetPrivacyPolicy)
	e.POST("/api/v2/policies/privacy", con.PostPrivacyPolicy)

	// TODO: one-line으로 trailingSlash 방안 찾기(2019/5/9) 하지만 SDK에서도 slash(/)를 제거하였기에, 2년후에 제거
	e.GET("/policies/:lang/:name/terms/", con.GetTerms)
	e.GET("/policies/:lang/:name/privacy/", con.GetPrivacy)

	return con
}

// GetTerms redirects to return terms.
func (con *Controller) GetTerms(c core.Context) error {
	err := con.validateRequest(c)
	if err != nil {
		return &core.HttpError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("https://policy.buzzvil.com/%s/buzzscreen/%s/terms/", c.Param("lang"), c.Param("name")))
}

// GetPrivacy redirects to return privacy.
func (con *Controller) GetPrivacy(c core.Context) error {
	err := con.validateRequest(c)
	if err != nil {
		return &core.HttpError{
			Code:    http.StatusBadRequest,
			Message: err.Error(),
		}
	}
	return c.Redirect(http.StatusFound, fmt.Sprintf("https://policy.buzzvil.com/%s/buzzscreen/%s/privacy-policy/", c.Param("lang"), c.Param("name")))
}

// GetPrivacyPolicy returns Privacy Policy translated in different languages.
func (con *Controller) GetPrivacyPolicy(c core.Context) error {
	ctx := c.Request().Context()
	var ppReq dto.GetPrivacyPolicyRequest
	if err := con.Bind(c, &ppReq); err != nil {
		return err
	}
	ppReq.Request = c.Request()

	unit, err := con.getUnit(ctx, &ppReq, con.AppUseCase)
	if err != nil {
		core.Logger.Warnf("GetPrivacyPolicy() - Failed to get unit. err: %s", err)
	}
	ppRes := dto.GetPrivacyPolicyResponse{}

	if unit != nil && (unit.Country == "" || gdprCountries[unit.Country]) && gdprCountries[con.getCountry(&ppReq, con.LocUseCase)] {
		ppRes.IsRequired = true
		ppRes.PrivacyPolicyTranslations = &(map[string]dto.PrivacyPolicyTranslation{
			"en": {
				Title:   "Privacy Notice",
				Content: `This app personailzes your lockscreen experience. Buzzvil Co.,Ltd. and its partners may collect and process personal data such as online identifiers to provide the service. By confirming, you’ll see the ad or content that are personalized to you. <a href="https://policy.buzzvil.com/privacy-policy">See our Privacy Policy</a>`,
				Consent: `I am confirming I am over the age of 16 and would like a personalized lockscreen experience.`,
			},
			"de": {
				Title:   "Datenschutzerklärung",
				Content: `Diese App personalisiert Ihre Lockscreen-Erfahrung. Buzzvil Co., Ltd. und seine Partner können personenbezogene Daten sammeln und verarbeiten, um den Dienst bereitzustellen. Mit der Bestätigung sehen Sie die Anzeige oder den Inhalt, die für Sie personalisiert sind. <a href="https://policy.buzzvil.com/privacy-policy">Siehe unsere Datenschutzrichtlinie </a>`,
				Consent: `Ich bestätige, dass ich älter als 16 bin und einen personalisierten Lockscreen möchte.`,
			},
			"pt": {
				Title:   "Aviso de privacidade",
				Content: `Este aplicativo personaliza sua experiência de bloqueio de tela. Buzzvil Co., Ltd. e seus parceiros podem coletar e processar dados pessoais, como identificadores on-line para fornecer o serviço. Ao confirmar, você verá o anúncio ou o conteúdo personalizado para você. <a href="http://www.buzzvil.com/privacy-policy-2”>Veja nossa Política de Privacidade</a>`,
				Consent: `Estou confirmando que tenho mais de 16 anos e gostaria de ter uma experiência de bloqueio de tela personalizada.`,
			},
			"it": {
				Title:   "Informativa sulla Privacy",
				Content: `Questa applicazione invia l'esperienza lockscreen. Buzzvil Co., Ltd. e i suoi partner possono raccogliere ed elaborare dati personali come identificativi online per fornire il servizio. Confermando, vedrai l'annuncio o i contenuti che sono personalizzati per te. <a href="http://www.buzzvil.com/privacy-policy-2”>Vedi la nostra politica sulla privacy</a>`,
				Consent: `Sto confermando che ho più di 16 anni e vorrei un'esperienza di lockscreen personalizzata.`,
			},
			"ru": {
				Title:   "Уведомление о конфиденциальности",
				Content: `Это приложение обрабатывает ваши функции lockscreen. Buzzvil Co., Ltd. и его партнеры могут собирать и обрабатывать личные данные, такие как онлайн-идентификаторы, для предоставления услуги. Подтвердив, вы увидите объявление или контент, которые персонализированы для вас. <a href="https://policy.buzzvil.com/privacy-policy”>См. Политику конфиденциальности</a>`,
				Consent: `Я подтверждаю, что мне исполнилось 16 лет, и мне хотелось бы, чтобы у вас был персонализированный экран lockscreen.`,
			},
			"fr": {
				Title:   "Avis de confidentialité",
				Content: `Cette application personnalise votre expérience de lockscreen. Buzzvil Co., Ltd. et ses partenaires peuvent collecter et traiter des données personnelles telles que des identifiants en ligne pour fournir le service. En confirmant, vous verrez l’annonce ou le contenu personnalisé. <a href="https://policy.buzzvil.com/privacy-policy>Voir notre politique de confidentialité</a>`,
				Consent: `Je confirme que j'ai plus de 16 ans et que je souhaite une expérience personnalisée du lockscreen.`,
			},
			"es": {
				Title:   "Aviso de Privacidad",
				Content: `Esta aplicación personaliza tu experiencia de pantalla de bloqueo. Buzzvil Co., Ltd. y sus socios pueden recopilar y procesar datos personales para proporcionar el servicio. Al confirmar, verá el anuncio o el contenido que están personalizados para usted. <a href="https://policy.buzzvil.com/privacy-policy>Consulte nuestra Política de privacidad</a>`,
				Consent: `Confirmo que soy mayor de 16 años y me gustaría tener pantalla de bloqueo personalizada.`,
			},
		})
	} else {
		ppRes.IsRequired = false
	}

	core.Logger.Infof("GetPrivacyPolicy() - Req: %v, IsRequired: %v", ppReq, ppRes.IsRequired)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":   code.CodeOk,
		"result": ppRes,
	})
}

// PostPrivacyPolicy posts privacy policy based on the request
func (con *Controller) PostPrivacyPolicy(c core.Context) error {
	var ppReq dto.PostPrivacyRequest
	if err := con.Bind(c, &ppReq); err != nil {
		return err
	}
	core.Logger.Infof("PostPrivacyPolicy() - Req: %+v", ppReq)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"code": code.CodeOk,
	})
}

func (con *Controller) validateRequest(c core.Context) error {
	if c.Param("lang") == "" || c.Param("name") == "" {
		return errors.New("lang and name shouldn't be empty")
	}
	return nil
}

func (con *Controller) getUnit(ctx context.Context, ppReq *dto.GetPrivacyPolicyRequest, useCase app.UseCase) (*app.Unit, error) {
	if ppReq.AppID > 0 {
		return useCase.GetUnitByAppID(ctx, ppReq.AppID)
	} else if ppReq.UnitIDV1 > 0 {
		return useCase.GetUnitByID(ctx, ppReq.UnitIDV1)
	}
	return nil, errors.New("app_id == 0 && unit_id == 0")
}

func (con *Controller) getCountry(ppReq *dto.GetPrivacyPolicyRequest, useCase location.UseCase) string {
	country := ppReq.GetCountry()
	if country == "" {
		if ppReq.CountryReq != "" {
			country = ppReq.CountryReq
		} else {
			if ppReq.Request != nil {
				loc := useCase.GetClientLocation(ppReq.Request, country)
				country = loc.Country
			}
		}
	}
	return country
}

var gdprCountries = map[string]bool{
	"AT": true, //Austria
	"BE": true, //Belgium
	"BG": true, //Bulgaria
	"HR": true, //Croatia
	"CY": true, //Cyprus
	"CZ": true, //Czech Republic
	"DK": true, //Denmark
	"EE": true, //Estonia
	"FI": true, //Finland
	"FR": true, //France
	"DE": true, //Germany
	"GR": true, //Greece
	"HU": true, //Hungary
	"IS": true, //Iceland
	"IE": true, //Ireland
	"IT": true, //Italy
	"LV": true, //Latvia
	"LI": true, //Liechtenstein
	"LT": true, //Lithuania
	"LU": true, //Luxembourg
	"MT": true, //Malta
	"NL": true, //Netherlands
	"NO": true, //Norway
	"PL": true, //Poland
	"PT": true, //Portugal
	"RO": true, //Romania
	"SK": true, //Slovakia
	"SI": true, //Slovenia
	"ES": true, //Spain
	"SE": true, //Sweden
	"GB": true, //United Kingdom
}
