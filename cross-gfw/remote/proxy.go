package remote

import (
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

var (
	httpConnected = []byte("HTTP/1.1 200 Connection established\r\n\r\n")
)

// Proxy 远程代理隧道服务
type Proxy struct {
	Dial func(address string) (net.Conn, error)

	serverName string // 目标主机
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var host string

	if req.Method == http.MethodConnect {
		// TODO 鉴权
		host = req.RequestURI
		if strings.LastIndex(host, ":") == -1 {
			host += ":443"
		}
	} else {
		host = req.Host
		if strings.LastIndex(host, ":") == -1 {
			host += ":80"
		}
	}

	// 连接目标主机
	upConn, err := p.Dial(host)
	if err != nil {
		http.Error(w, "cannot connect to upstream", http.StatusBadGateway)
		log.Println("dial to upstream error: ", err)
		return
	}
	defer upConn.Close()

	hj := w.(http.Hijacker)
	downConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "cannot hijack", http.StatusInternalServerError)
		log.Println("remote proxy hijack error: ", err)
		return
	}
	defer downConn.Close()

	// 已经和目标主机建立连接 返回给客户端代理  CONNECT established
	if req.Method == http.MethodConnect {
		downConn.Write(httpConnected)
	} else {
		// 直接将 http 请求序列化成字节，发送给目标主机
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Println("remote proxy dump http request error: ", err)
			http.Error(w, "cannot dump request", http.StatusInternalServerError)
			return
		}

		upConn.Write(dump)
	}

	go func() {
		io.Copy(upConn, downConn)
	}()

	io.Copy(downConn, upConn)
}

// NewRemoteProxy 创建远程代理的实例
func NewRemoteProxy() *Proxy {
	return &Proxy{
		Dial: func(address string) (net.Conn, error) {
			return net.DialTimeout("tcp", address, 500*time.Millisecond) // 远程代理连接目标主机
		},
	}
}
