package comm

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

/**
 * 解析请求路径上的变量，并能转化成不同类型
 */
type MuxVarsParser struct {
	Vars map[string]string
}

func NewMuxVarsParser(r *http.Request) *MuxVarsParser {
	return &MuxVarsParser{Vars: mux.Vars(r)}
}

func (this *MuxVarsParser) GetInt(key string, d int) int {
	var value int
	var err error
	str, ok := this.Vars[key]
	if !ok {
		value = d
	} else {
		value, err = strconv.Atoi(str)
		if err != nil {
			value = d
		}
	}
	return value
}

func (this *MuxVarsParser) GetString(key string, d string) string {
	value, ok := this.Vars[key]
	if !ok {
		value = d
	}
	return value
}
