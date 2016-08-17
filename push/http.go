package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

func HttpPostJson(httpUrl string, params Dict, invoker, key string) (Dict, Error) {
	//fmt.Printf("http post <%v> to %s\n", params, httpUrl)

	jstr, _ := json.Marshal(params)
	buf := fmt.Sprintf("%s%s%s", string(jstr), invoker, key)
	t := md5.New()
	io.WriteString(t, buf)
	sign := fmt.Sprintf("%x", t.Sum(nil))
	body := bytes.NewBuffer(jstr)
	//fmt.Printf("http post <%v> to %s\n", string(jstr), httpUrl)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", httpUrl, body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("BUGLE-PROVIDER-INVOKER", invoker)
	req.Header.Set("BUGLE-PROVIDER-SIGN", sign)

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("post<%s> to <%s>failed, %s\n", params, httpUrl, err)
		return nil, Error{REMOTE_CONN_ERR, err, "Remote serever can't access"}
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("post<%s> failed, %s\n", httpUrl, err)
		return nil, Error{REMOTE_RESP_ERR, err, "Remote serever response error"}
	}
	//fmt.Printf("http response:<%s>\n", string(result))

	dict := Dict{}
	json.Unmarshal([]byte(result), &dict)

	//fmt.Printf("http response:<%v>\n", dict)
	if dict["ok"] == true {
		return dict, OK
	} else {
		return nil, Error{REMOTE_RESP_ERR, nil, "error"}
	}

	return dict, OK
}

func HttpPost(httpUrl string, params Dict) (Dict, Error) {
	fmt.Printf("http post <%v> to %s\n", params, httpUrl)
	data := url.Values{}
	for key, val := range params {
		data.Set(key, val.(string))
	}

	body := ioutil.NopCloser(strings.NewReader(data.Encode()))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", httpUrl, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("post<%s> to <%s>failed, %s\n", params, httpUrl, err)
		return nil, Error{REMOTE_CONN_ERR, err, "Remote serever can't access"}
	}
	defer resp.Body.Close()

	result, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("post<%s> failed, %s\n", httpUrl, err)
		return nil, Error{REMOTE_RESP_ERR, err, "Remote serever response error"}
	}
	fmt.Printf("http response:<%s>\n", string(result))

	dict := Dict{}
	json.Unmarshal([]byte(result), &dict)

	fmt.Printf("http response:<%v>\n", dict)
	if dict["ok"] == true {
		return dict, OK
	} else {
		return nil, Error{REMOTE_RESP_ERR, nil, "error"}
	}

	return dict, OK
}
