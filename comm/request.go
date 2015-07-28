package comm

import (
	"net/http"
	"strconv"
)

func FormGetInt(r *http.Request, key string, d int) int {
	value := FormGetString(r, key, "")
	if value == "" {
		return d
	}
	ret, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return ret
}

func FormGetString(r *http.Request, key string, d string) string {
	value := r.Form.Get(key)
	if value == "" {
		value = r.URL.Query().Get(key)
	}
	if value == "" {
		return d
	}
	return value
}
