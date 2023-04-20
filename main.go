package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"

	pb "github.com/crossplane/crossplane/apis/apiextensions/fn/proto/v1alpha1"
	vp "github.com/vshn/appcat-comp-functions/functions/vshn-postgres-func"
	"github.com/vshn/appcat-comp-functions/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	Network  string = "unix"
	LogLevel int    = 1
)

type server struct {
	pb.UnimplementedContainerizedFunctionRunnerServiceServer
	ctx context.Context
}

func (s *server) RunFunction(ctx context.Context, in *pb.RunFunctionRequest) (*pb.RunFunctionResponse, error) {
	switch in.Image {
	case "postgresql":
		fnio, err := runtime.RunCommand(s.ctx, in.Input, postgresFunctions)
		if err != nil {
			return &pb.RunFunctionResponse{
				Output: fnio,
			}, status.Errorf(codes.Aborted, "Can't process request for PostgreSQL")
		}
		return &pb.RunFunctionResponse{
			Output: fnio,
		}, nil
	case "redis":
		return &pb.RunFunctionResponse{
			// return what was sent as it's currently not supported
			Output: in.Input,
		}, status.Error(codes.Unimplemented, "Redis is not yet implemented")
	default:
		return &pb.RunFunctionResponse{
			Output: []byte("Bad configuration"),
		}, status.Error(codes.NotFound, "Unknown request")
	}
}

func main() {
	cntx := context.Background()
	err := runtime.SetupLogging(vp.AI, &cntx, LogLevel)
	if err != nil {
		log.Fatal("logging broke, exiting")
	}
	_ = runtime.LogMetadata(cntx, vp.AI)

	var AddressFlag = flag.String("socket", "@crossplane/fn/default.sock", "optional -> set where socket should be located")
	flag.IntVar(&LogLevel, "loglevel", 1, "optional -> set log level [0,1]")
	flag.Parse()

	if err := cleanStart(*AddressFlag); err != nil {
		log.Fatal(err)
	}

	lis, err := net.Listen(Network, *AddressFlag)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterContainerizedFunctionRunnerServiceServer(s, &server{ctx: cntx})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

// socket isn't removed after server stop listening and blocks another starts
func cleanStart(socketName string) error {
	if _, err := os.Stat(socketName); err == nil {
		err := os.RemoveAll(socketName)
		if err != nil {
			return err
		}
	}

	return nil
}
