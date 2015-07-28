package proxy

import (
	"encoding/json"
	"io/ioutil"
	"koogroup/lib/log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/**
 * 发出http post请求
 */
func HttpPost(result interface{}, urlStr string, values url.Values, timeout time.Duration) error {
	logger := log.DefaultLogger()

	logger.Infoln(urlStr, values.Encode())

	request, err := http.NewRequest("POST", urlStr, strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}
	request.Header.Set("Expect:", "")

	client := http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	logger.Infoln(string(body))

	err = json.Unmarshal(body, result)
	if err != nil {
		return err
	}

	return nil
}

/**
 * 发出http get请求
 */
func HttpGet(result interface{}, urlStr string, values url.Values, timeout time.Duration) error {
	logger := log.DefaultLogger()

	urlStr = urlStr + "?" + values.Encode()
	logger.Infoln(urlStr)

	request, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: timeout}
	response, err := client.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	logger.Infoln(string(body))

	err = json.Unmarshal(body, result)
	if err != nil {
		return err
	}

	return nil
}
