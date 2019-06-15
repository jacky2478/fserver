package web

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/cors"
)

/* ======TServer====== */
type TServer struct {
	Name      string
	Port      string
	Static    string `json:"-"`
	ApiPrefix string

	// http Handler
	Engine IEngine `json:"-"`

	middwares []*TMiddware
	handlers  []*THandler

	// http server
	server *http.Server

	// restart callback function
	// input: none
	// output: run server or not
	onShutdown func() bool

	// signal channel for shutdown
	shutdownCh chan bool

	// allow origins
	allowOrigins []string
}

func (p *TServer) RegistHandler(handlers ...*THandler) *TServer {
	p.handlers = append(p.handlers, handlers...)
	return p
}

func (p *TServer) RegistMiddware(middwares ...*TMiddware) *TServer {
	p.middwares = append(p.middwares, middwares...)
	return p
}

func (p *TServer) RegistOnShutdown(f func() bool) {
	p.shutdownCh = make(chan bool, 0)
	p.server.RegisterOnShutdown(func() {
		select {
		case p.shutdownCh <- p.onShutdown():
		default:
			return
		}
	})
	p.onShutdown = f
}

func (p *TServer) RegistAllowOrigins(origins ...string) {
	p.allowOrigins = origins[:]
}

func (p *TServer) Run() error {
	return p.listenAndServe()
}

func (p *TServer) Restart() error {
	p.server.Shutdown(context.Background())
	if runs, ok := <-p.shutdownCh; ok && runs {
		return p.listenAndServe()
	}
	return nil
}

func (p *TServer) DefaultStatus(rw http.ResponseWriter, r *http.Request, params url.Values) error {
	if !strings.HasSuffix(r.URL.Path, "/status") {
		return nil
	}
	ret := struct {
		ID     string
		Server interface{}
	}{
		ID:     SessionID(r),
		Server: getStatus(C_Status_Server, r),
	}
	ResponseOk(rw, r, ret)
	return nil
}

func newCors(allowOrigins ...string) *cors.Cors {
	if len(allowOrigins) == 0 {
		allowOrigins = append(allowOrigins, "*")
	}
	corsMiddleware := cors.New(cors.Options{
		AllowedOrigins: allowOrigins,
	})
	return corsMiddleware
}

func (p *TServer) listenAndServe() error {
	if p.Engine == nil {
		p.Engine = useNegroni()
	}
	if p.Engine != nil {
		if err := p.Engine.InitEngine(); err != nil {
			return fmt.Errorf("%s running failed while doing InitEngine, detail: %v", p.Name, err.Error())
		}
		if err := p.Engine.InitMiddware(p.middwares...); err != nil {
			return fmt.Errorf("%s running failed while doing InitMiddware, detail: %v", p.Name, err.Error())
		}
		if err := p.Engine.InitStatic(p.Static); err != nil {
			return fmt.Errorf("%s running failed while doing InitStatic, detail: %v", p.Name, err.Error())
		}
		if err := p.Engine.InitRouter(p.ApiPrefix, p.handlers...); err != nil {
			return fmt.Errorf("%s running failed while doing InitRouter, detail: %v", p.Name, err.Error())
		}

		p.logServer()
		p.server = &http.Server{Addr: fmt.Sprintf("0.0.0.0:%v", p.Port), Handler: newCors(p.allowOrigins...).Handler(p.Engine)}

		if err := p.server.ListenAndServe(); err != nil {
			return fmt.Errorf("%s running failed, detail: %v", p.Name, err.Error())
		}
		return nil
	}
	return fmt.Errorf("%s running failed without valid engine", p.Name)
}

func (p *TServer) logServer() {
	fmt.Printf("\n\n%s running at 0.0.0.0:%v, powered by fserver:\n{\n", p.Name, p.Port)
	fmt.Printf("	%v: Static\n", p.Static)
	fmt.Printf("\n")
	for _, v := range p.middwares {
		if v.BeforeSession {
			fmt.Printf("	%s: BeforeSession\n", v.Name)
		}
		if !v.BeforeSession {
			fmt.Printf("	%s: AfterSession\n", v.Name)
		}
	}
	fmt.Printf("\n")
	for _, v := range p.handlers {
		fmt.Printf("	%v%v: %v\n", p.ApiPrefix, v.Path, v.Method)
	}

	allowOrigins := p.allowOrigins[:]
	if len(allowOrigins) == 0 {
		allowOrigins = append(allowOrigins, "*")
	}
	fmt.Printf("\n	AllowOrigins: [%v]\n", strings.Join(allowOrigins, ","))
	fmt.Printf("}\n\n")
}

/* ======TServer====== */
