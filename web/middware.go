package web

import (
	"fmt"
	"net/http"

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

/* ======TMiddware====== */
