package main

import (
	"context"
	"fmt"
	"net/http"
	pb "twirp_demo/rpc/business/v1"
)

/**
	根据业务创建 rpc 的 protobuf 文件

	`
	|--rpc
		|---business
				|---v0 对内接口
					|-------service.proto
				|---v1 对外接口
					|-------service.proto
 */


var data = map[int32]string {

}

type Server struct {}

func (s *Server) Get(ctx context.Context, req *pb.GetReq) (resp *pb.GetResp, err error) {
	name := data[req.Id]

	resp = &pb.GetResp{
		Name: name,
	}

	if name == "" {
		resp.Name = "None"
	}

	return
}

func (s *Server) Set(ctx context.Context, req *pb.SetReq) (resp *pb.SetResp, err error) {
	maxID := int32(0)

	for k := range data {
		if k > maxID {
			maxID = k
		}
	}

	data[maxID+1] = req.Name
	resp = &pb.SetResp{
		Id: maxID,
	}
	return
}


func (s *Server) Del(ctx context.Context, req *pb.DelReq) (resp *pb.DefaultResp, err error) {
	delete(data, req.Id)
	return
}



func main() {
	twirpHandler := pb.NewBusinessServer(&Server{}, nil)

	mux := http.NewServeMux()
	mux.Handle(pb.BusinessPathPrefix, twirpHandler)

	fmt.Println("start server ....")
	http.ListenAndServe(":8080", mux)
}
