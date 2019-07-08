package main

import (
	"flag"
	"os"
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

	var err error
	if localAddr != "" {

	}
}
