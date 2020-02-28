package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"strconv"
	"io/ioutil"
)

func MakeRequest(client *http.Client, method string, headers map[string]string, url string, body interface{}, result interface{}, file *os.File, code int) error {
	var request *http.Request
	var err error
	if body != nil {
		requestBytes, err := json.Marshal(body)
		if err != nil {
			return err
		}
		request, err = http.NewRequest(method, url, bytes.NewReader(requestBytes))
		if err != nil {
			return err
		}
	} else {
		request, err = http.NewRequest(method, url, nil)
		if err != nil {
			return err
		}
	}
	for k, v := range headers {
		request.Header.Set(k, v)
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode == code {
		if result != nil {
			return json.NewDecoder(response.Body).Decode(result)
		} else if file != nil {
			_, err = io.Copy(file, response.Body)
			if err != nil {
				return err
			}
			return nil
		}
		return nil
	} else {
		d, _ := ioutil.ReadAll(response.Body)
		return errors.New("Response return status code: " + strconv.Itoa(response.StatusCode) + " which different from " + strconv.Itoa(code) + ", for request: " + url + "; message: " + string(d))
	}
}

