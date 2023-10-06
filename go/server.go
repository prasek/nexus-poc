// package main

// import (
// 	"context"
// 	"fmt"
// 	"log"
// 	"net"
// 	"net/http"

// 	"github.com/nexus-rpc/sdk-go/nexus"
// )

// type handler struct {
// 	nexusservice.UnimplementedNexusServiceServer
// }

// // StartOperation implements nexusservice.NexusServiceServer.
// func (*handler) StartOperation(ctx context.Context, request *nexusservice.StartOperationRequest) (*nexusservice.StartOperationResponse, error) {
// 	fmt.Println("request", request)
// 	return &nexusservice.StartOperationResponse{}, nil
// }

// func (*handler) CancelOperation(context.Context, *nexusservice.CancelOperationRequest) (*nexusservice.CancelOperationResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method CancelOperation not implemented")
// }
// func (*handler) GetOperationResult(context.Context, *nexusservice.GetOperationResultRequest) (*result.OperationResult, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetOperationResult not implemented")
// }
// func (*handler) GetOperationInfo(context.Context, *nexusservice.GetOperationInfoRequest) (*nexusservice.OperationInfo, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method GetOperationInfo not implemented")
// }
// func (*handler) DeliverResult(context.Context, *nexusservice.DeliverResultRequest) (*nexusservice.DeliverResultResponse, error) {
// 	return nil, status.Errorf(codes.Unimplemented, "method DeliverResult not implemented")
// }

// var port = ":50051"

// func main() {
// 	go runProxy()
// 	runServer()
// }

// func runServer() {
// 	lis, err := net.Listen("tcp", port)
// 	if err != nil {
// 		log.Fatalf("failed to listen: %v", err)
// 	}
// 	s := grpc.NewServer()
// 	fmt.Printf("\nServer listening on port %v \n", port)
// 	nexusservice.RegisterNexusServiceServer(s, &handler{})
// 	if err := s.Serve(lis); err != nil {
// 		log.Fatalf("failed to serve: %v", err)
// 	}
// }

// func runProxy() error {
// 	ctx := context.Background()
// 	ctx, cancel := context.WithCancel(ctx)
// 	defer cancel()

// 	mux := runtime.NewServeMux()
// 	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

// 	err := nexusservice.RegisterNexusServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Print("\nProxy listening on 8081\n")
// 	return http.ListenAndServe(":8081", mux)
// }
