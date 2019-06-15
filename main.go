package main

import (
	"flag"
	"net/http"

	"github.com/jery1024/fserver/file"

	"github.com/jery1024/fserver/web"
	"github.com/jery1024/mlog"
)

var port = flag.String("port", "9999", "--port=8080")
var mode = flag.String("mode", "web", "--mode=web,file")

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
	case "file":
		go runFileServer()
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

func runFileServer() {
	server := web.NewTServer("fserver", *port, "./", "/api")
	server.RegistHandler(web.NewTHandler("/file/upload", http.MethodPost, nil, file.UploadFile, nil))
	server.RegistHandler(web.NewTHandler("/status", http.MethodGet, nil, server.DefaultStatus, nil))
	server.RegistMiddware(web.NewStatusMiddware())
	if err := server.Run(); err != nil {
		mlog.Fatal(err.Error())
	}
}

func wait() {
	select {}
}
