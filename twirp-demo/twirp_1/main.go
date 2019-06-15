package main

import (
	"context"
	"fmt"
	"net/http"
)

/**
# protoc demo.proto  --go_out=. --twirp_out=.
# # 根据 demo.proto 生成 go 文件
# # --go_out 生成 go 的 pb 文件
# # --twirp_out 生成 go 的 twirp 文件   (twirp依赖pb文件)
*/

type HelloWorldServer struct{}

// Hello FIXME 会炸
func (s *HelloWorldServer) Hello(ctx context.Context, req *HelloReq) (resp *HelloResp, err error) {
	resp = &HelloResp{
		Text: "Hello " + req.Subject,
	}
	return
}

// 编译 运行 server
// # go run *.go
func main() {
	twirpHandler := NewHelloWorldServer(&helloWorldServer{}, nil)

	mux := http.NewServeMux()

	mux.Handle(HelloWorldPathPrefix, twirpHandler)

	fmt.Println("start server ....")
	http.ListenAndServe(":8080", mux)
}
