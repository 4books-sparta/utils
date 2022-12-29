package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"

	"gopkg.in/go-playground/validator.v9"
)

func IsSlug(in string) bool {
	reg := regexp.MustCompile("^([a-z0-9_-]{1,500})$")
	valid := reg.MatchString(in)
	return valid
}

func ValidateRequest(request interface{}) error {
	validate := validator.New()
	_ = validate.RegisterValidation("slug", isSlug)

	err := validate.Struct(request)
	if err != nil {
		return ValidationError{
			Children: err.(validator.ValidationErrors),
		}
	}

	return nil
}

func isSlug(fl validator.FieldLevel) bool {
	return IsSlug(fl.Field().String())
}

func DecodeRequestBody(req *http.Request, request interface{}) error {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	//fmt.Printf("\nDecodingBody... \n %s", b)

	req.Body = io.NopCloser(bytes.NewBuffer(b))

	return json.Unmarshal(b, request)
}
