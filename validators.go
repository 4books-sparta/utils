package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"

	"gopkg.in/go-playground/validator.v9"
)

func IsSlug(in string) bool {
	if in == "" {
		return true
	}
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

func DecodeRequestBody(req *http.Request, ret interface{}) error {
	b, err := DecodeToInterface(req.Body, ret)

	req.Body = io.NopCloser(bytes.NewBuffer(b))

	return err
}

func DecodeResponseBody(res *http.Response, ret interface{}) error {
	b, err := DecodeToInterface(res.Body, ret)

	res.Body = io.NopCloser(bytes.NewBuffer(b))

	return err
}

type Msg struct {
	Message string `json:"message"`
}

func ServiceResponse2Error(res *http.Response) error {
	msg := Msg{}
	err := DecodeResponseBody(res, &msg)
	if err != nil {
		return err
	}

	return errors.New(msg.Message)
}

func DecodeToInterface(r io.Reader, ret interface{}) ([]byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return []byte(""), err
	}

	return b, json.Unmarshal(b, ret)
}
