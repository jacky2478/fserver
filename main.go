package main

import (
	"flag"
	"net/http"

	"github.com/jery1024/fserver/web"
	"github.com/jery1024/mlog"
)

var port = flag.String("port", "9999", "--port=8080")
var mode = flag.String("mode", "web", "--mode=web")

func main() {
	flag.Parse()
	initLog()

	if err := runServer(); err != nil {
		mlog.Error(err.Error())
	}
	wait()
}

func initLog() {
	mlog.SetDepth(4)
	mlog.SetFlags(mlog.LstdFlags)
	mlog.SetHighlighting(false)
	mlog.SetLevel(mlog.LOG_LEVEL_ALL)
}

func runServer() error {
	switch *mode {
	case "web":
		go runWebServer()
	}
	return nil
}

func runWebServer() {
	server := web.NewTServer("fserver", *port, "./", "/api")
	server.RegistHandler(web.NewTHandler("/status", http.MethodGet, nil, server.DefaultStatus, nil))
	server.RegistMiddware(web.NewStatusMiddware())
	if err := server.Run(); err != nil {
		mlog.Fatal(err.Error())
	}
}

func wait() {
	select {}
}
