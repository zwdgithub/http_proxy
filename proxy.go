package main

import (
	xhttp "github.com/zwdgithub/simple_http"
	"strings"
)

type ProxyGetter interface {
	GetProxy() string
}

type ProxyHttpGetter struct {
}

func (g *ProxyHttpGetter) GetProxy() (string, error) {
	content, err := xhttp.NewHttpUtil().Get("http://182.61.186.195:8091/get").Do().RContent()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(content, "https://") && !strings.HasPrefix(content, "http://") {
		content = "https://" + content
	}
	return content, err
}
