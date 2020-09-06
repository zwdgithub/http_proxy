package test

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestSelect(t *testing.T) {
	a := &ListNode{Val: 1}
	b := &ListNode{Val: 2}
	c := &ListNode{Val: 3}
	d := &ListNode{Val: 5}
	e := &ListNode{Val: 5}
	a.Next = b
	b.Next = c
	c.Next = d
	d.Next = e
	for {
		fmt.Println(a.Val)
		if a.Next == nil {
			break
		}
		a = a.Next
	}

}

type ListNode struct {
	Val  int
	Next *ListNode
}

func TestHttp(t *testing.T) {
	p := func(_ *http.Request) (*url.URL, error) {
		return url.Parse("http://" + "localhost:7856")
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Proxy:           p,
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   time.Second * 3,
	}

	resp, err := client.Get("http://myip.ipip.net")
	if err != nil{
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(b))
}
