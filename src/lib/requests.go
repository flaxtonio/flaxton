package lib
import (
	"net/http"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

func PostRequest(url string, data []byte, headers map[string]string) (resp_content []byte, err error) {
	var (
		req *http.Request
		resp *http.Response
	)
	req, err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return
	}
	for k, v :=range headers {
		req.Header.Set(k, v)
	}
	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	resp_content, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body.Close()
	return
}

func PostJson(url string, data []byte, obj interface{}, auth string) (err error) {
	var (
		content []byte
	)

	content, err = PostRequest(url, data, map[string]string{
		"Content-Type": "application/json",
		"Authorization": auth,
	})

	if obj != nil {
		err = json.Unmarshal(content, obj)
		if err != nil {
			return err
		}
	}
	return nil
}