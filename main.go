package main

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

func DefaultClient(ip string) *http.Client {
	p := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("http://" + ip)
	}
	transport := &http.Transport{
		Proxy:           p,
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
	}
	return client
}

func Get(u, ip string, c chan []byte, ctx context.Context) {
	req, _ := http.NewRequest(http.MethodGet, u, nil)
	req = req.WithContext(ctx)
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Baiduspider/2.0; +http://www.baidu.com/search/spider.html)")
	start := time.Now().Unix()
	resp, err := DefaultClient(ip).Do(req)

	if err != nil {
		log.Printf("%s, error: %v", u, err)
		return
	}

	log.Printf("%s, get cost %ds", u, time.Now().Unix()-start)
	defer resp.Body.Close()
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("%s, error: %v", u, err)
		return
	}
	c <- bytes
}

func mainN() {
	redisUtil := NewRedis()
	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		strings := r.URL.Query()["url"]
		if len(strings) <= 0 {
			_, _ = w.Write([]byte(""))
			return
		}
		u := strings[0]
		ipList, err := redisUtil.GetProxies()
		if err != nil {
			_, _ = w.Write([]byte(""))
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		c := make(chan []byte, 5)
		for _, ip := range ipList {
			go Get(u, ip, c, ctx)
		}
		var b []byte
		for {
			select {
			case b = <-c:
				goto end
			case <-time.After(time.Second * 5):
				b = []byte("")
				goto end
			}
		}
	end:
		cancel()
		_, _ = w.Write(b)
	})

	if err := http.ListenAndServe(":8092", nil); err != nil {
		log.Fatalf("run error %v", err)
	}
}
