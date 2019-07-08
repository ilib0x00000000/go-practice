package local

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

// httpConnected 当本地客户端发起一个 HTTP/HTTPS 请求之后，本地的客户端会给远程的代理发起会话
// 远程的代理收到请求后，会发起到目标机器的连接，连接建立后需要返回给本地的客户端，表示隧道已经建立
var httpConnected = []byte("HTTP/1.1 200 Connection established\r\n\r\n")

type Proxy struct {
	Dial func(address string) (net.Conn, error)

	serverName string
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

	if req.Method == http.MethodConnect {
		downConn.Write(httpConnected)
	} else {
		// 因为需要走 TLS 所以把 http的内容 序列化成 字节发送
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			http.Error(w, "cannot dump request", http.StatusInternalServerError)
			log.Println("dump https request error: ", err)
			return
		}

		upConn.Write(dump) // 发送 HTTPS 内容
	}

	go func() {
		io.Copy(upConn, downConn)
	}()

	io.Copy(downConn, upConn)
}

// NewClientProxy 创建本地的客户端代理
func NewClientProxy(remote string, useTLS bool) *Proxy {
	serverName := remote

	i := strings.LastIndex(remote, ":")
	if i == -1 {
		remote = remote + ":443"
	} else {
		serverName = remote[:i]
	}

	p := &Proxy{serverName: serverName}
	p.Dial = func(address string) (conn net.Conn, err error) {
		// host := address[:strings.LastIndex(address, ":")]
		// TODO 判断 host 在不在 GFW 黑名单上
		// 如果 host 在 GFW 的黑名单上，连接远程的代理
		// 如果不在直连
		// proxy: remote    dst:address
		log.Printf("dial %s via %s", address, remote)

		// 直连远程的代理
		conn, err = net.DialTimeout("tcp", remote, 1*time.Second)
		if err != nil {
			log.Println("dial remote proxy error: ", err)
		}

		// 使用 TLS 协议
		if useTLS {
			conn = tls.Client(conn, &tls.Config{
				ServerName:         serverName,
				MinVersion:         tls.VersionTLS13,
				ClientSessionCache: tls.NewLRUClientSessionCache(0),
			})
		}
		defer conn.Close()

		// 发送 CONNECT 头
		req := "CONNECT " + address + " HTTP/1.1\r\n"
		_, err = conn.Write([]byte(req))
		if err != nil {
			log.Println("send \"HTTP CONNECT\" error: ", err)
			return
		}

		// proxy 收到 CONNECT 头之后会连接 目标dst host，连接成功之后会发送 CONNECT 请求，然后dst host返回返回 httpConnected
		buf := make([]byte, len(httpConnected))
		_, err = conn.Read(buf[:])
		if err != nil {
			log.Println("read from remote error: ", err)
			return
		}

		if !bytes.Equal(buf, httpConnected) {
			log.Println("remote CONNECT dst host error: ", err)
			err = errors.New(string(buf))
		}

		return
	}

	return p
}
