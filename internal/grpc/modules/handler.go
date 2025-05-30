package modulespb

import (
	"context"

	pb "github.com/The-OpenPlatform/backend/internal/grpc/modules"
)

// HealthCheck implements pb.ModulesServiceServer
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	return &pb.HealthCheckResponse{Status: "OK"}, nil
}

// Register implements pb.ModulesServiceServer
func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return &pb.RegisterResponse{Success: true, Message: "User registered"}, nil
}

// Setup implements pb.ModulesServiceServer
func (s *Server) Setup(ctx context.Context, req *pb.SetupRequest) (*pb.SetupResponse, error) {
	return &pb.SetupResponse{Success: true, Message: "Setup complete"}, nil
}

// Delete implements pb.ModulesServiceServer
func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return &pb.DeleteResponse{Success: true, Message: "User deleted"}, nil
}

// Server is your implementation of pb.ModulesServiceServer.
// You must define this struct elsewhere or here if not already present.
type Server struct {
	pb.UnimplementedModulesServiceServer
}
