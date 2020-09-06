package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/url"
	"runtime/debug"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":7856")
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Printf("Accept error %v", err)
		}
		go handle(client)
	}

}

func handle(client net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("handle recover, erros is %v", err)
			debug.PrintStack()
		}
	}()
	if client == nil {
		return
	}
	log.Printf("client tcp tunnel connection: local: %s -> remote %s", client.LocalAddr().String(), client.RemoteAddr().String())
	defer client.Close()

	var b [1024]byte
	n, err := client.Read(b[:]) //读取应用层的所有数据
	if err != nil || bytes.IndexByte(b[:], '\n') == -1 {
		log.Println(err) //传输层的连接是没有应用层的内容 比如：net.Dial()
		return
	}
	var method, host, address string
	fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
	log.Println(method, host)
	hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" { //https访问
		address = hostPortURL.Scheme + ":443"
	} else {                                            //http访问
		if strings.Index(hostPortURL.Host, ":") == -1 { //host不带端口， 默认80
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

	server, err := Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	//在应用层完成数据转发后，关闭传输层的通道
	defer server.Close()
	log.Println("server tcp tunnel connection:", server.LocalAddr().String(), "->", server.RemoteAddr().String())
	// server.SetDeadline(time.Now().Add(time.Duration(10) * time.Second))

	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
	} else {
		log.Println("server write", method) //其它协议
		server.Write(b[:n])
	}

	//进行转发
	go func() {
		io.Copy(server, client)
	}()
	io.Copy(client, server) // 阻塞转发
}
