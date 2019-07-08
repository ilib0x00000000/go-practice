package local

import (
	"log"
	"net"
	"net/http"
	"strings"
)

// httpConnected 当本地客户端发起一个 HTTP/HTTPS 请求之后，本地的客户端会给远程的代理发起会话
// 远程的代理收到请求后，会发起到目标机器的连接，连接建立后需要返回给本地的客户端，表示隧道已经建立
var httpConnected = []byte("HTTP/1.1 200 Connection established\r\n\r\n")

type Proxy struct {
	Dial func(address string) (net.Conn, error)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var host string

	if req.Method == http.MethodConnect { // HTTPS
		host := req.RequestURI
		if strings.LastIndex(host, ":") == -1 {
			host += ":443"
		}
	} else {
		// HTTP 请求
		host := req.Host
		if strings.LastIndex(host, ":") == -1 {
			host += ":80"
		}
	}

	upConn, err := p.Dial(host)
	if err != nil {
		http.Error(w, "cannot connect to upstream", http.StatusBadGateway)
		log.Println("dial to upstream err: ", err)
		return
	}
	defer upConn.Close()

	hj := w.(http.Hijacker)
	downConn, _, err := hj.Hijack()
	if err != nil {
		http.Error(w, "cannot hijack", http.StatusInternalServerError)
		log.Println("hijack error: ", err)
		return
	}
	defer downConn.Close()
}
