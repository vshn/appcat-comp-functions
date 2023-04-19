package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/crossplane/crossplane/apis/apiextensions/fn/proto/v1alpha1"
	vp "github.com/vshn/appcat-comp-functions/functions/vshn-postgres-func"
	"github.com/vshn/appcat-comp-functions/runtime"
	"google.golang.org/grpc"
)

var postgresFunctions = []runtime.Transform{
	{
		Name:          "url-connection-details",
		TransformFunc: vp.AddUrlToConnectionDetails,
	},
	{
		Name:          "user-alerting",
		TransformFunc: vp.AddUserAlerting,
	},
	{
		Name:          "random-default-schedule",
		TransformFunc: vp.TransformSchedule,
	},
}

var (
	Network = "unix"
	//Address = "@crossplane/fn/default.sock"
	// for testing purposes, especially on MacOS it's much easier to create local socket than whole directory structure and permissions
	Address = "default.sock"
)

type server struct {
	pb.UnimplementedContainerizedFunctionRunnerServiceServer
}

func (s *server) RunFunction(ctx context.Context, in *pb.RunFunctionRequest) (*pb.RunFunctionResponse, error) {
	switch in.Image {
	case "postgresql":
		fnio, err := runtime.RunCommand(&ctx, in.Input, postgresFunctions)
		return &pb.RunFunctionResponse{
			Output: fnio,
		}, err
	case "redis":
		return &pb.RunFunctionResponse{
			// return what was sent as it's currently not supported
			Output: in.Input,
		}, nil
	default:
		return &pb.RunFunctionResponse{
			Output: []byte("Bad configuration"),
		}, fmt.Errorf("unrecogised configuration")
	}
}

func main() {
	lis, err := net.Listen(Network, Address)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterContainerizedFunctionRunnerServiceServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
