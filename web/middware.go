package web

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	sessions "github.com/goincremental/negroni-sessions"
)

// const value
const (
	c_task_ID         = "taskID"
	c_session_ID      = "session_fserver"
	c_parse_params_ok = "parseParamsOK"
)

/* ======TMiddware====== */
type TMiddware struct {
	Name          string
	BeforeStatic  bool
	BeforeSession bool
	DoHandleFunc  func(rw http.ResponseWriter, r *http.Request, params url.Values) error
}

func sessionID(r *http.Request) string {
	sessID := sessions.GetSession(r).Get(c_session_ID)
	if sessID == nil {
		return ""
	}
	return fmt.Sprintf("%v", sessID)
}

var statusMap sync.Map

func getStatus(key string, r *http.Request) interface{} {
	sessID := sessionID(r)
	ret, _ := statusMap.Load(fmt.Sprintf("%v_%v", sessID, key))
	return ret
}

func setStatus(key string, val interface{}, r *http.Request) {
	sessID := sessionID(r)
	if sessID != "" {
		statusMap.Store(fmt.Sprintf("%v_%v", sessID, key), val)
	}
}

/* ======TMiddware====== */
