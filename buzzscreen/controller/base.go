package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	"github.com/Buzzvil/buzzlib-go/core"
	"github.com/Buzzvil/buzzscreen-api/internal/app/api/common"
	"github.com/asaskevich/govalidator"
)

func bindValue(c core.Context, model interface{}) error {
	if err := c.Bind(model); err != nil {
		return common.NewBindError(err)
	}

	if err := bindHeaders(c, model); err != nil {
		return common.NewBindError(err)
	}

	if err := c.Validate(model); err != nil {
		return common.NewBindError(err)
	}

	// TODO - we shouldn't try to bind 'request' for every bind request
	{
		modelRef := reflect.ValueOf(model)
		requestField := modelRef.Elem().FieldByName("Request")
		if requestField.IsValid() {
			req := c.Request()
			reqValue := reflect.ValueOf(&req).Elem() // contextRef.Elem().FieldByName("Request")
			requestField.Set(reqValue)
		}
	}
	return nil
}

// bindHeaders binds header values to the model. It assumes that all the values
// are string for now. The feature is supposed to be supported by echo
// framework.
func bindHeaders(c core.Context, model interface{}) error {
	typ := reflect.TypeOf(model).Elem()
	val := reflect.ValueOf(model).Elem()

	for i := 0; i < typ.NumField(); i++ {
		typeField := typ.Field(i)
		structField := val.Field(i)
		fieldName := typeField.Tag.Get("header")
		if fieldName != "" {
			inputValue := c.Request().Header.Get(fieldName)
			structField.SetString(inputValue)
		}
	}

	return nil
}

func isV2Request(c core.Context) bool {
	return strings.Contains(c.Request().URL.Path, "v2/")
}

func bindRequestSupport(c core.Context, v3Req interface{}, v2Req interface{}) error {
	if isV2Request(c) {
		if err := bindValue(c, v2Req); err != nil {
			return common.NewBindError(err)
		}
		if err := copyValue(v2Req, v3Req); err != nil {
			return common.NewBindError(err)
		}
		return nil
	}
	return bindValue(c, v3Req)
}

func getResponseSupport(c core.Context, res interface{}) interface{} {
	if isV2Request(c) {
		snakeCased, err := json.Marshal(res)
		if err != nil {
			c.Error(err)
			return nil
		}
		core.Logger.Debugf("getResponseSupport() - snakeCased: %s", snakeCased)
		re := regexp.MustCompile(`"(\w+)":`)
		camelCased := re.ReplaceAllStringFunc(string(snakeCased), func(m string) string {
			a := []rune(govalidator.UnderscoreToCamelCase(m))
			a[1] = unicode.ToLower(a[1])
			c := string(a)
			return c
		})
		core.Logger.Debugf("getResponseSupport() - camelCased: %s", camelCased)
		var results interface{}
		//results := make(map[string]interface{})
		err = json.Unmarshal([]byte(camelCased), &results)
		if err != nil {
			c.Error(err)
			return nil
		}
		//changeMapFromSnakeToCamel(&results)
		return results
	}
	return res
}

func copyValue(a interface{}, b interface{}) error {
	v2Ref := reflect.ValueOf(a)
	v3Ref := reflect.ValueOf(b)

	if v2Ref.Kind() == reflect.Ptr {
		v2Ref = v2Ref.Elem()
	}
	if v3Ref.Kind() == reflect.Ptr {
		v3Ref = v3Ref.Elem()
	}

	for i := 0; i < v2Ref.NumField(); i++ {
		v2Field := v2Ref.Field(i)
		v2Value := v2Field

		v3Field := v3Ref.Field(i)
		typeField := v2Ref.Type().Field(i)
		tag := typeField.Tag

		switch {
		case v2Value.Kind() == reflect.Struct && typeField.Anonymous == true:
			err := copyValue(v2Field.Addr().Interface(), v3Field.Addr().Interface())
			if err != nil {
				return err
			}
			continue
		case (tag.Get("form") != "" && tag.Get("form") != "-" || tag.Get("query") != "" && tag.Get("query") != "-") && v2Value.IsValid() && v2Value.Interface() != "":
			if v2Ref.Type().Field(i).Name != v3Ref.Type().Field(i).Name {
				return fmt.Errorf("%v(v2Field) and %v(v3Field) is different", v2Ref.Type().Field(i).Name, v3Ref.Type().Field(i).Name)
			}
			v3Field.Set(v2Value)
		}
	}
	return nil
}

type (
	// BaseResponse type definition
	BaseResponse struct {
		Code    int     `json:"code"`
		Message *string `json:"message,omitempty"`
		Result  struct {
		} `json:"result"`
	}
)

// Empty func definition
func Empty(c core.Context) error {
	return c.JSON(http.StatusOK, BaseResponse{})
}
