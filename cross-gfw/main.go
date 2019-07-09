package main

import (
	"flag"
	"os"

	"github.com/mholt/certmagic"
	proxy "remote"
)

var (
	localAddr  string
	remoteAddr string
	useTLS     bool
)

func init() {
	os.Setenv("GODEBUG", os.Getenv("GODEBUG")+",tls13=1") // tls v1.3

	flag.StringVar(&localAddr, "local", "", "local proxy listen address")
	flag.StringVar(&remoteAddr, "remote", "", "remote proxy host")
	flag.BoolVar(&useTLS, "tls", true, "enable tls for tunnel")
}

func main() {
	flag.Parse()

	if localAddr == "" && remoteAddr == "" {
		flag.Usage()
		os.Exit(1)
	}

	// 远程代理
	var err error
	handler := proxy.NewRemoteProxy()

	if useTLS {
		certmagic.Default.Email = "ilib0x00000001@gmail.com"
		certmagic.Default.CA = certmagic.LetsEncryptProductionCA

		ln, err := certmagic.Listen([]string{remoteAddr})
		if err != nil {
			panic(err)
		}

		err = http.Serve(ln.handler)
	} else {
		remoteAddr = remoteAddr[strings.LastIndex(remoteAddr, ":"):]
		err = http.ListenAndServe(remoteAddr, handler)
	}

	if err != nil {
		panic(err)
	}
}
