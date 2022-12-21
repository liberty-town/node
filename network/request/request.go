package request

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"liberty-town/node/network/network_config"
	"net/http"
)

func RequestPostJson(address string, data any, result any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return RequestPost(address, bytes, result, "application/json")
}

func RequestPost(address string, data []byte, result any, contentType string) error {

	client := http.Client{
		Timeout: network_config.REQUEST_TIMEOUT,
	}

	resp, err := client.Post(address, contentType, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusBadRequest {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bytes))
	}

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

func RequestGet(address string, result any) error {

	client := http.Client{
		Timeout: network_config.REQUEST_TIMEOUT,
	}

	resp, err := client.Get(address)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusBadRequest {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(bytes))
	}

	if err = json.NewDecoder(resp.Body).Decode(result); err != nil {
		return err
	}

	return nil
}

func RequestGetData(address string) ([]byte, error) {

	client := http.Client{
		Timeout: network_config.REQUEST_TIMEOUT,
	}

	resp, err := client.Get(address)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(bytes))
	}

	return ioutil.ReadAll(resp.Body)
}
