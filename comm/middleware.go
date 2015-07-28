package comm

import (
	"fmt"
	"github.com/codegangsta/negroni"
	"net/http"
	"runtime"
)

type RecoverHandle func(http.ResponseWriter, *http.Request, interface{}, []byte)

func Recover(handle RecoverHandle) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		defer func() {
			if err := recover(); err != nil {
				rw.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, 1024*8)
				stack = stack[:runtime.Stack(stack, false)]
				handle(rw, r, err, stack)
			}
		}()
		next(rw, r)
	}
}

func ParseForm() negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if r.Method != "GET" && r.Method != "HEAD" {
			if r.Header.Get("Content-Type") == "multipart/form-data" {
				err := r.ParseMultipartForm(32 << 20)
				if err != nil {
					http.Error(rw, fmt.Sprintf("parse multipart form error: %s", err), 403)
					return
				}
			} else {
				err := r.ParseForm()
				if err != nil {
					http.Error(rw, fmt.Sprintf("parse form error: %s", err), 403)
					return
				}
			}
		}
		next(rw, r)
	}
}

func SetDefaultContentType(contentType string) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Set("Content-type", contentType)
		next(rw, r)
	}
}
