package gocom

import (
	"encoding/json"
	"github.com/diffguo/gocom/log"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// 有reqform时，可能需要设置header：request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
func FormatReqUrl(reqUrl string, reqForm map[string]string) string {
	if reqForm == nil || len(reqForm) == 0 {
		return reqUrl
	}

	form := url.Values{}
	return reqUrl + "?" + form.Encode()
}

// method POST GET
// headers Ext Header
func DoHttpRequest(method string, reqUrl string, headers map[string]string, body io.Reader) (int, []byte, error) {
	return _doHttpRequest(method, reqUrl, headers, body)
}

func DoHttpRequestWithBody(method string, reqUrl string, headers map[string]string, reqBody interface{}) (int, []byte, error) {
	if reqBody == nil {
		return _doHttpRequest(method, reqUrl, headers, nil)
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Content-Type"] = "application/json"
	b, _ := json.Marshal(reqBody)
	reader := strings.NewReader(string(b))
	return _doHttpRequest(method, reqUrl, headers, reader)
}

func _doHttpRequest(method string, reqUrl string, headers map[string]string, body io.Reader) (int, []byte, error) {
	req, _ := http.NewRequest(method, reqUrl, body)

	if headers != nil && len(headers) > 0 {
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := http.Client{}
	response, err := client.Do(req)
	if nil != err {
		log.Errorf("send request err: %v", err)
		return http.StatusNotFound, nil, err
	}

	defer response.Body.Close()

	rspBody, err := ioutil.ReadAll(response.Body)
	if nil != err {
		rspBody = make([]byte, 0)
	}

	return response.StatusCode, rspBody, err
}
