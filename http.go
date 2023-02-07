package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type MicroserviceClient struct {
	TimeOut time.Duration
	Url     string
	Port    uint
}

func (msc MicroserviceClient) Post(path string, payload interface{}, ret interface{}) error {
	u := msc.getUrl(path)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("POST", u, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	if msc.TimeOut > 0 {
		client.Timeout = msc.TimeOut
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return ServiceResponse2Error(response)
	}

	if ret == nil {
		//Nothing to parse
		return nil
	}

	err = DecodeResponseBody(response, ret)
	if err != nil {
		return err
	}

	return nil
}

func (msc MicroserviceClient) Patch(path string, payload interface{}, ret interface{}) error {
	u := msc.getUrl(path)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PATCH", u, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	if msc.TimeOut > 0 {
		client.Timeout = msc.TimeOut
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return ServiceResponse2Error(response)
	}

	if ret == nil {
		//Nothing to parse
		return nil
	}

	err = DecodeResponseBody(response, ret)
	if err != nil {
		return err
	}

	return nil
}

func (msc MicroserviceClient) Get(path string, ret interface{}, VV url.Values) error {
	u := msc.getUrl(path)
	request, err := http.NewRequest("GET", u, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}

	if VV != nil && len(VV) > 0 {
		request.URL.RawQuery = VV.Encode()
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	if msc.TimeOut > 0 {
		client.Timeout = msc.TimeOut
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode == http.StatusNotFound {
		return NotFound{}
	}

	if response.StatusCode != http.StatusOK {
		return ServiceResponse2Error(response)
	}

	if ret == nil {
		//Nothing to parse
		return nil
	}

	err = DecodeResponseBody(response, ret)
	if err != nil {
		return err
	}

	return nil
}

func (msc MicroserviceClient) Delete(path string, ret interface{}) error {
	u := msc.getUrl(path)
	request, err := http.NewRequest("DELETE", u, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{}
	if msc.TimeOut > 0 {
		client.Timeout = msc.TimeOut
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		return ServiceResponse2Error(response)
	}

	if ret == nil {
		//Nothing to parse
		return nil
	}

	err = DecodeResponseBody(response, ret)
	if err != nil {
		return err
	}

	return nil
}

func (msc MicroserviceClient) getUrl(path string) string {
	return msc.Url + ":" + strconv.Itoa(int(msc.Port)) + path
}
