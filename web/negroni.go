package web

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strings"

	sessions "github.com/goincremental/negroni-sessions"
	"github.com/goincremental/negroni-sessions/cookiestore"
	"github.com/jery1024/mlog"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func useNegroni() IEngine {
	return &tNegroni{}
}

type tNegroni struct {
	engine *negroni.Negroni
}

func (p *tNegroni) InitEngine() error {
	p.engine = negroni.New()
	return nil
}

func newSession() negroni.Handler {
	pwd := getRandomString(32)
	data := []byte(pwd)
	has := md5.Sum(data)
	pwd = fmt.Sprintf("%x", has)

	store := cookiestore.New([]byte(pwd))
	store.Options(sessions.Options{
		MaxAge: 86400,
	})
	refitself_session := "session_fserver"
	return negroni.HandlerFunc(sessions.Sessions(refitself_session, store))
}

func (p *tNegroni) InitStatic(path string, middwares ...*TMiddware) error {
	for _, m := range middwares {
		if m.BeforeStatic && m.DoHandleFunc != nil {
			p.engine.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				params := toUrlValues(r)
				if err := m.DoHandleFunc(rw, r, params); err != nil {
					return
				}
				next(rw, r)
			}))
		}
	}
	p.engine.Use(negroni.NewStatic(http.Dir(path)))
	return nil
}

func (p *tNegroni) InitMiddware(middwares ...*TMiddware) error {
	if len(middwares) == 0 {
		p.engine.Use(newSession())
		return nil
	}
	for _, m := range middwares {
		if m.BeforeSession && m.DoHandleFunc != nil {
			p.engine.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				params := toUrlValues(r)
				if err := m.DoHandleFunc(rw, r, params); err != nil {
					return
				}
				next(rw, r)
			}))
		}
	}
	p.engine.Use(newSession())
	for _, m := range middwares {
		if !m.BeforeSession && m.DoHandleFunc != nil {
			p.engine.Use(negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
				params := toUrlValues(r)
				if err := m.DoHandleFunc(rw, r, params); err != nil {
					return
				}
				next(rw, r)
			}))
		}
	}
	return nil
}

func (p *tNegroni) InitRouter(apiPrefix string, handlers ...*THandler) error {
	apiRouterPrefix := apiPrefix
	apiRouter := httprouter.New()
	for _, th := range handlers {
		if th == nil {
			mlog.Errorf("regist router failed with invalid THandler, handlers: %+v", handlers)
			continue
		}
		if !strings.HasPrefix(th.Path, "/") {
			th.Path = "/" + th.Path
		}
		doRegistHandler(apiRouter, apiRouterPrefix+th.Path, th)
	}
	p.engine.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if strings.HasPrefix(r.URL.Path, apiRouterPrefix) {
			apiRouter.ServeHTTP(rw, r)
			return
		}
		next(rw, r)
	})
	return nil
}

func (p *tNegroni) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	p.engine.ServeHTTP(rw, r)
}

func doRegistHandler(router *httprouter.Router, path string, th *THandler) {
	var rfunc func(path string, handle httprouter.Handle)
	switch th.Method {
	case http.MethodGet:
		rfunc = router.GET
	case http.MethodPost:
		rfunc = router.POST
	case http.MethodPut:
		rfunc = router.PUT
	case http.MethodHead:
		rfunc = router.HEAD
	case http.MethodDelete:
		rfunc = router.DELETE
	case http.MethodOptions:
		rfunc = router.OPTIONS
	case http.MethodPatch:
		rfunc = router.PATCH
	default:
		mlog.Errorf("doRegistHandler failed with MethodNotAllowed, method: %v, path: %v", th.Method, path)
		return
	}

	rfunc(path, func(rw http.ResponseWriter, r *http.Request, params httprouter.Params) {
		uvs := toUrlValues(r)
		if th.BeforeHandleFunc != nil {
			if err := th.BeforeHandleFunc(rw, r, uvs); err != nil {
				mlog.Error(err.Error())
				return
			}
		}
		if th.DoHandleFunc != nil {
			if err := th.DoHandleFunc(rw, r, uvs); err != nil {
				mlog.Error(err.Error())
				return
			}
		}
		if th.AfterHandleFunc != nil {
			if err := th.AfterHandleFunc(rw, r, uvs); err != nil {
				mlog.Error(err.Error())
				return
			}
		}
	})
}
