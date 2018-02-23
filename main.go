package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	listenAddr := flag.String("listen", ":8081", "proxy listen addr")

	if !flag.Parsed() {
		flag.Parse()
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	l, err := net.Listen("tcp", *listenAddr)
	if err != nil {
		log.Panic(err)
	}

	for {
		client, err := l.Accept()
		if err != nil {
			log.Panic(err)
		}

		go serve(client)
	}
}

func serve(client net.Conn) {
	if client == nil {
		return
	}
	defer client.Close()

	var requestContent [1024]byte
	nr, err := client.Read(requestContent[:])
	if err != nil {
		log.Println(err)
		return
	}

	method, host := parseRequest(requestContent[:])
	hostUrl, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("req<", method, host)

	// 根据不同协议构造目标地址，注意加上端口号
	var address string
	if hostUrl.Opaque == "443" {
		address = hostUrl.Scheme + ":443"
	} else {
		address = hostUrl.Host
		if strings.Index(address, ":") == -1 {
			address = address + ":80"
		}
	}

	// 与请求的目标建立连接
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}
	defer server.Close()

	if method == http.MethodConnect {
		// 隧道代理
		// HTTP 客户端通过 CONNECT 方法请求隧道代理创建一条到达任意目的服务器和端口的 TCP 连接，并对客户端和服务器之间的后继数据进行盲转发。
		// 对于 CONNECT 请求来说，只是用来让代理创建 TCP 连接，所以只需要提供服务器域名及端口即可，并不需要具体的资源路径。
		// 代理收到这样的请求后，并响应给请求方一个 HTTP 报文：HTTP/1.1 200 Connection Established
		client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	} else {
		// 普通代理
		// HTTP 客户端向代理发送请求报文，代理服务器需要正确地处理请求和连接（例如正确处理 Connection: keep-alive），同时向服务器发送请求，并将收到的响应转发给客户端。
		// 把请求数据转发给server
		server.Write(requestContent[:nr])
	}

	// 双向转发，直到报错或者双向全部copy完毕
	if err := transport(client, server); err != nil {
		log.Println(err)
	}
}

func parseRequest(content []byte) (string, string) {
	var method, host string
	if i := bytes.IndexByte(content, '\n'); i >= 0 {
		content = content[0:i]
	}
	fields := bytes.Fields(content)
	if len(fields) >= 1 {
		method = string(fields[0])
	}
	if len(fields) >= 2 {
		host = string(fields[1])
	}
	return method, host
}

func transport(local, remote io.ReadWriteCloser) error {
	errch := make(chan error, 1)

	go copy(local, remote, errch)
	go copy(remote, local, errch)

	for i := 0; i < 2; i++ {
		if err := <-errch; err != nil {
			return err
		}
	}
	return nil
}

func copy(dst, src io.ReadWriter, errch chan error) {
	_, err := io.Copy(dst, src)
	errch <- err
}
