package modules

import (
	"context"

	"github.com/The-OpenPlatform/backend/internal/db"
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
	var message string = "Module creation failed!"
	var exists bool
	var uid string

	query := `SELECT EXISTS (SELECT 1 FROM modules WHERE name = $1)`
	err := db.DB.Get(&exists, query, req.Name)
	if err == nil {
		if !exists {
			query = `INSERT INTO modules (name, ip_port) VALUES ($1) RETURNING module_id`
			err = db.DB.Get(&uid, query, req.Name)
			if err == nil {
				message = "Module created successfully!"
				return &RegisterResponse{Success: true, ModuleId: uid, Message: message}, nil
			}
		} else {
			message = "Module with the same name already exists."
		}
	}
	return &RegisterResponse{Success: false, ModuleId: "", Message: message}, err
}

// Setup implements ModulesServiceServer
func (s *Server) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	return &SetupResponse{Success: true, Message: "Setup complete"}, nil
}

// Delete implements ModulesServiceServer
func (s *Server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	return &DeleteResponse{Success: true, Message: "User deleted"}, nil
}
