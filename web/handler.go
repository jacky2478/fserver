package web

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jery1024/mlog"
)

/* ======THandler====== */
type THandler struct {
	Name             string
	Path             string
	Method           string
	DoHandleFunc     func(rw http.ResponseWriter, r *http.Request, params url.Values) error
	BeforeHandleFunc func(rw http.ResponseWriter, r *http.Request, params url.Values) error
	AfterHandleFunc  func(rw http.ResponseWriter, r *http.Request, params url.Values) error
}

func toUrlValues(r *http.Request) url.Values {
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.ParseMultipartForm(32 << 20)
	} else {
		r.ParseForm()
	}

	ret := make(url.Values, 0)
	for k, _ := range r.URL.Query() {
		ret.Add(k, r.URL.Query().Get(k))
	}
	for k, _ := range r.Header {
		ret.Add(k, r.Header.Get(k))
	}
	for k, _ := range r.Form {
		if ret.Get(k) == "" {
			ret.Add(k, r.Form.Get(k))
		}
	}
	for k, _ := range r.PostForm {
		if ret.Get(k) == "" {
			ret.Add(k, r.PostForm.Get(k))
		}
	}
	mlog.Infof("[%v] %v: %v, request: %+v", r.Header.Get(c_task_ID), r.Method, r.URL.Path, ret)
	return ret
}

func getRandomString(n int) string {
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" + fmt.Sprintf("%v", time.Now().UnixNano())

	buffer := make([]byte, n)
	max := big.NewInt(int64(len(alphanum)))

	for i := 0; i < n; i++ {
		index, err := randomInt(max)
		if err != nil {
			return ""
		}

		buffer[i] = alphanum[index]
	}

	return string(buffer)
}

func randomInt(max *big.Int) (int, error) {
	rand, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return int(rand.Int64()), nil
}

/* ======THandler====== */
