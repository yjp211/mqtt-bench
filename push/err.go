package main

import "encoding/json"

const (
	SUCCESS = iota

	REMOTE_CONN_ERR
	REMOTE_RESP_ERR

	BROKER_CONN_ERR
	BROKER_PUSH_ERR
)

type Error struct {
	Code int   // error code
	Err  error // origin error
	Msg  string
}

func (e Error) String() string {
	if e.Err != nil {
		return e.Msg + ", " + e.Err.Error()
	}
	return e.Msg
}

func (e Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Msg
}

func (e Error) Ok() bool {
	if e.Code == SUCCESS {
		return true
	}
	return false
}

func (e Error) Json() string {
	dict := Dict{}
	if e.Code == SUCCESS {
		dict["ok"] = true
	} else {
		dict["ok"] = false
		dict["code"] = e.Code
		dict["msg"] = e.Msg
	}
	jstr, _ := json.Marshal(dict)
	return string(jstr)
}

var OK Error = Error{SUCCESS, nil, ""}
