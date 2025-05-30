package main

import (
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"

	"github.com/The-OpenPlatform/backend/internal/api"
	"github.com/The-OpenPlatform/backend/internal/db"
	"github.com/The-OpenPlatform/backend/internal/grpc/modules"
)

func main() {
	db.MustConnect()
	defer db.Close()

	r := api.SetupRouter()
	go startGRPCServer()

	log.Println("Server is running on port 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}

func startGRPCServer() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	registerGRPCServices(grpcServer)

	log.Println("gRPC server running on :50051")
	if err := grpcServer.Serve(lis); err != nil {

		log.Fatalf("failed to serve: %v", err)
	}
}

func registerGRPCServices(grpcServer *grpc.Server) {
	modules.RegisterModulesServiceServer(grpcServer, &modules.Server{})
}
