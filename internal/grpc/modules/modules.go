package modules

import (
	"context"
	"fmt"

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
			query = `INSERT INTO modules (name, ip_port) VALUES ($1,$2) RETURNING module_id`
			err = db.DB.Get(&uid, query, req.Name, fmt.Sprintf("%s:%d", req.Ip, req.Port))
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
	var message string = "Module setup failed!"

	query := `INSERT INTO images (module_id, image, fileformat) VALUES ($1, $2, $3)
		ON CONFLICT (module_id) DO UPDATE SET image = EXCLUDED.image, fileformat = EXCLUDED.fileformat;`
	res, err := db.DB.Exec(query, req.ModuleId, req.Image, req.Fileformat)
	if err == nil {
		if rows, _ := res.RowsAffected(); rows > 0 {
			message = "Module setup completed successfully!"
			return &SetupResponse{Success: true, Message: message}, nil
		}
	}
	return &SetupResponse{Success: false, Message: message}, nil
}

// Delete implements ModulesServiceServer
func (s *Server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	var message string = "Module deletion failed!"

	query := `DELETE FROM modules WHERE module_id = $1;`
	res, err := db.DB.Exec(query, req.ModuleId)
	if err == nil {
		if rows, _ := res.RowsAffected(); rows > 0 {
			message = "Module deletion successful!"
			return &DeleteResponse{Success: true, Message: message}, nil
		}
	}
	return &DeleteResponse{Success: false, Message: message}, nil
}
