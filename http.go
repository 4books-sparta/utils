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
	TimeOut            time.Duration
	Url                string
	Port               uint
	PermanentHeaders   map[string]string
	PermanentUrlParams url.Values
	GetUrl             func(path string) string
	ErrorParser        func(res *http.Response) error
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

	if msc.PermanentUrlParams != nil && len(msc.PermanentUrlParams) > 0 {
		request.URL.RawQuery = msc.PermanentUrlParams.Encode()
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	msc.fillPermanentHeaders(request)

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
		if msc.ErrorParser != nil {
			return msc.ErrorParser(response)
		}
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

func (msc MicroserviceClient) Put(path string, payload interface{}, ret interface{}) error {
	u := msc.getUrl(path)
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	request, err := http.NewRequest("PUT", u, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	if msc.PermanentUrlParams != nil && len(msc.PermanentUrlParams) > 0 {
		request.URL.RawQuery = msc.PermanentUrlParams.Encode()
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	msc.fillPermanentHeaders(request)

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
		if msc.ErrorParser != nil {
			return msc.ErrorParser(response)
		}
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

	if msc.PermanentUrlParams != nil && len(msc.PermanentUrlParams) > 0 {
		request.URL.RawQuery = msc.PermanentUrlParams.Encode()
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	msc.fillPermanentHeaders(request)

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
		if msc.ErrorParser != nil {
			return msc.ErrorParser(response)
		}
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

	if msc.PermanentUrlParams != nil && len(msc.PermanentUrlParams) > 0 {
		if VV == nil {
			VV = url.Values{}
		}
		for k, vv := range msc.PermanentUrlParams {
			for _, v := range vv {
				VV.Add(k, v)
			}
		}
	}

	if VV != nil && len(VV) > 0 {
		request.URL.RawQuery = VV.Encode()
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	msc.fillPermanentHeaders(request)

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
		if msc.ErrorParser != nil {
			return msc.ErrorParser(response)
		}
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

func (msc MicroserviceClient) fillPermanentHeaders(req *http.Request) {
	if msc.PermanentHeaders == nil {
		return
	}
	for k, v := range msc.PermanentHeaders {
		req.Header.Set(k, v)
	}
}

func (msc MicroserviceClient) Delete(path string, ret interface{}) error {
	u := msc.getUrl(path)
	request, err := http.NewRequest("DELETE", u, bytes.NewBuffer([]byte("")))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
	msc.fillPermanentHeaders(request)

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
		if msc.ErrorParser != nil {
			return msc.ErrorParser(response)
		}
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
	if msc.GetUrl != nil {
		return msc.GetUrl(path)
	}
	return msc.Url + ":" + strconv.Itoa(int(msc.Port)) + path
}

type HttpMicroClient interface {
	Post(url string, payload interface{}, ret interface{}) error
	Put(url string, payload interface{}, ret interface{}) error
	Patch(url string, payload interface{}, ret interface{}) error
	Get(url string, ret interface{}, VV url.Values) error
	Delete(url string, ret interface{}) error
}
