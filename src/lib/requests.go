package lib
import (
	"net/http"
	"io/ioutil"
	"bytes"
	"encoding/json"
)

func PostJson(url string, data []byte, obj interface{}) error {
	var (
		req *http.Request
		req_err error
	)
	req, req_err = http.NewRequest("POST", url, bytes.NewBuffer(data))
	if req_err != nil {
		return req_err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, resp_error := client.Do(req)
	if resp_error != nil {
		return resp_error
	}
	content, read_err := ioutil.ReadAll(resp.Body)
	if read_err != nil {
		return read_err
	}
	if obj != nil {
		decode_error := json.Unmarshal(content, obj)
		if decode_error != nil {
			return decode_error
		}
	}
	resp.Body.Close()
	return nil
}