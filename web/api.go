package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/jacky2478/mlog"
)

/*
usage:
server := fserver.NewTServer("demo", "8080", "./", "/api")
server.Run()
*/

type IEngine interface {
	InitEngine() error
	InitMiddware(middwares ...*TMiddware) error
	InitStatic(path string, middwares ...*TMiddware) error
	InitRouter(apiPrefix string, handlers ...*THandler) error

	ServeHTTP(rw http.ResponseWriter, r *http.Request)
}

func NewTServer(name, port, static, apiPrefix string, handlers ...*THandler) *TServer {
	if !filepath.IsAbs(static) {
		if ret, err := filepath.Abs(static); err != nil {
			mlog.Errorf("NewTServer failed with invalid static directory, static: %v, detai: %v", static, err.Error())
			return nil
		} else {
			static = ret
		}
	}
	ts := &TServer{Name: name, Port: port, Static: static, ApiPrefix: apiPrefix, handlers: make([]*THandler, 0), middwares: make([]*TMiddware, 0)}
	if handlers != nil && len(handlers) > 0 {
		ts.handlers = handlers[:]
	}
	return ts
}

func NewTHandler(path, method string, beforeHandleFunc, doHandleFunc, afterHandleFunc func(rw http.ResponseWriter, r *http.Request, params url.Values) error) *THandler {
	return &THandler{Path: path, Method: method, BeforeHandleFunc: beforeHandleFunc, DoHandleFunc: doHandleFunc, AfterHandleFunc: afterHandleFunc}
}

func NewTMiddware(name string, beforeStatic, beforeSession bool, doHandleFunc func(rw http.ResponseWriter, r *http.Request, params url.Values) error) *TMiddware {
	return &TMiddware{Name: name, BeforeStatic: beforeStatic, BeforeSession: beforeSession, DoHandleFunc: doHandleFunc}
}

func NewStatusMiddware(ptr *TServer) *TMiddware {
	doTaskID := func(rw http.ResponseWriter, r *http.Request, params url.Values) error {
		if sessions.GetSession(r).Get(c_session_ID) == nil {
			sessID := getRandomString(32)
			sessions.GetSession(r).Set(c_session_ID, sessID)
			setStatus(C_Status_Server, ptr, r)
		}
		r.Header.Add(c_task_ID, getRandomString(6))
		return nil
	}
	return &TMiddware{Name: "StatusMiddware", BeforeSession: false, DoHandleFunc: doTaskID}
}

func ResponseDirect(w http.ResponseWriter, r *http.Request, data []byte) {
	mlog.Infof("[%v] %v: %v, response: %+v", r.Header.Get(c_task_ID), r.Method, r.URL.Path, string(data))

	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func ResponseError(w http.ResponseWriter, r *http.Request, errorInfo string) {
	ret := struct {
		Error string
	}{}
	ret.Error = errorInfo
	respBuf, err := json.MarshalIndent(ret, "", "")
	if err != nil {
		jsonErr := fmt.Errorf("ResponseError failed while doing json.MarshalIndent, detail: %v", err.Error())
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		mlog.Error(jsonErr.Error())
		return
	}
	mlog.Errorf("[%v] %v: %v, response: %+v", r.Header.Get(c_task_ID), r.Method, r.URL.Path, string(respBuf))

	w.Header().Set("Content-Length", strconv.Itoa(len(respBuf)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBuf)
}

func ResponseOk(w http.ResponseWriter, r *http.Request, data interface{}) {
	respBuf, err := json.MarshalIndent(data, "", "")
	if err != nil {
		jsonErr := fmt.Errorf("ResponseOk failed while doing json.MarshalIndent, detail: %v", err.Error())
		http.Error(w, jsonErr.Error(), http.StatusInternalServerError)
		mlog.Error(jsonErr.Error())
		return
	}
	mlog.Infof("[%v] %v: %v, response: %+v", r.Header.Get(c_task_ID), r.Method, r.URL.Path, string(respBuf))

	w.Header().Set("Content-Length", strconv.Itoa(len(respBuf)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBuf)
}

func SessionID(r *http.Request) string {
	return sessionID(r)
}

const (
	C_Status_Server = "server"
)

func GetStatus(key string, r *http.Request) interface{} {
	return getStatus(key, r)
}

func SetStatus(key string, val interface{}, r *http.Request) {
	setStatus(key, val, r)
}
