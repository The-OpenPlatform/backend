package modules

import (
	"context"
)

type Server struct {
	UnimplementedModulesServiceServer
}

// HealthCheck implements ModulesServiceServer
func (s *Server) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	// Also consider maintenance mode or other health checks
	return &HealthCheckResponse{Status: "OK"}, nil
}

// Register implements ModulesServiceServer
func (s *Server) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	return &RegisterResponse{Success: true, Message: "User registered"}, nil
}

// Setup implements ModulesServiceServer
func (s *Server) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	return &SetupResponse{Success: true, Message: "Setup complete"}, nil
}

// Delete implements ModulesServiceServer
func (s *Server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	return &DeleteResponse{Success: true, Message: "User deleted"}, nil
}
