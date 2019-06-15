package web

import (
	"fmt"
	"net/http"
	"sync"

	sessions "github.com/goincremental/negroni-sessions"
)

// const value
const (
	c_task_ID    = "taskID"
	c_session_ID = "sessionID"
)

/* ======TMiddware====== */
type TMiddware struct {
	Name          string
	BeforeSession bool
	DoHandleFunc  http.HandlerFunc
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
