// Package modules provides gRPC service implementation for module management.
// It handles module registration, setup, deletion, and health checking operations
// with comprehensive input validation and error handling.
package modules

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/The-OpenPlatform/backend/internal/db"
)

// Server implements the ModulesServiceServer interface and provides
// methods for managing modules in the system.
type Server struct {
	UnimplementedModulesServiceServer
}

// HealthCheck returns the health status of the modules service.
// It performs basic validation and tests database connectivity.
func (s *Server) HealthCheck(ctx context.Context, req *HealthCheckRequest) (*HealthCheckResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("health check request cannot be nil")
	}

	// Test database connectivity
	if err := db.DB.PingContext(ctx); err != nil {
		return &HealthCheckResponse{Status: "UNHEALTHY"}, fmt.Errorf("database connection failed: %w", err)
	}

	return &HealthCheckResponse{Status: "OK"}, nil
}

// Register creates a new module with the given name and IP:port combination.
// It validates input parameters, checks for name conflicts, and creates the module.
// Returns an error if the module name already exists or if input validation fails.
func (s *Server) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("register request cannot be nil")
	}

	// Validate input parameters
	if err := s.validateRegisterRequest(req); err != nil {
		return &RegisterResponse{
			Success:  false,
			ModuleId: "",
			Message:  fmt.Sprintf("Validation failed: %s", err.Error()),
		}, nil
	}

	// Check if module with same name already exists
	moduleExists, err := s.moduleNameExists(ctx, req.Name)
	if err != nil {
		return &RegisterResponse{
			Success:  false,
			ModuleId: "",
			Message:  "Failed to check module existence",
		}, fmt.Errorf("failed to check module existence: %w", err)
	}

	if moduleExists {
		return &RegisterResponse{
			Success:  false,
			ModuleId: "",
			Message:  "Module with the same name already exists",
		}, nil
	}

	// Create new module
	moduleID, err := s.createModule(ctx, req.Name, req.Ip, req.Port)
	if err != nil {
		return &RegisterResponse{
			Success:  false,
			ModuleId: "",
			Message:  "Module creation failed",
		}, fmt.Errorf("failed to create module: %w", err)
	}

	return &RegisterResponse{
		Success:  true,
		ModuleId: moduleID,
		Message:  "Module created successfully",
	}, nil
}

// Setup configures a module with image data and file format.
// It validates the request, verifies the module exists, and performs an upsert operation
// to handle both insert and update scenarios for module images.
func (s *Server) Setup(ctx context.Context, req *SetupRequest) (*SetupResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("setup request cannot be nil")
	}

	// Validate input parameters
	if err := s.validateSetupRequest(req); err != nil {
		return &SetupResponse{
			Success: false,
			Message: fmt.Sprintf("Validation failed: %s", err.Error()),
		}, nil
	}

	// Verify module exists before setup
	moduleExists, err := s.moduleIDExists(ctx, req.ModuleId)
	if err != nil {
		return &SetupResponse{
			Success: false,
			Message: "Failed to verify module existence",
		}, fmt.Errorf("failed to verify module existence: %w", err)
	}

	if !moduleExists {
		return &SetupResponse{
			Success: false,
			Message: "Module not found",
		}, nil
	}

	// Perform module setup
	if err := s.setupModuleImage(ctx, req.ModuleId, req.Image, req.Fileformat); err != nil {
		return &SetupResponse{
			Success: false,
			Message: "Module setup failed",
		}, fmt.Errorf("failed to setup module: %w", err)
	}

	return &SetupResponse{
		Success: true,
		Message: "Module setup completed successfully",
	}, nil
}

// Delete removes a module and its associated data from the system.
// It validates the request and removes the specified module.
// Returns success even if the module doesn't exist to maintain idempotency.
func (s *Server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("delete request cannot be nil")
	}

	// Validate input parameters
	if err := s.validateDeleteRequest(req); err != nil {
		return &DeleteResponse{
			Success: false,
			Message: fmt.Sprintf("Validation failed: %s", err.Error()),
		}, nil
	}

	// Delete module (cascading deletes should handle related records)
	deleted, err := s.deleteModule(ctx, req.ModuleId)
	if err != nil {
		return &DeleteResponse{
			Success: false,
			Message: "Module deletion failed",
		}, fmt.Errorf("failed to delete module: %w", err)
	}

	if !deleted {
		return &DeleteResponse{
			Success: true,
			Message: "Module not found (already deleted)",
		}, nil
	}

	return &DeleteResponse{
		Success: true,
		Message: "Module deleted successfully",
	}, nil
}

// Helper methods for validation and database operations

// validateRegisterRequest validates the register request parameters.
// It checks for empty module names, name length limits, valid IP addresses,
// and port number ranges (1-65535).
func (s *Server) validateRegisterRequest(req *RegisterRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("module name cannot be empty")
	}

	if len(req.Name) > 255 {
		return fmt.Errorf("module name too long (max 255 characters)")
	}

	if net.ParseIP(req.Ip) == nil {
		return fmt.Errorf("invalid IP address: %s", req.Ip)
	}

	if req.Port <= 0 || req.Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be 1-65535)", req.Port)
	}

	return nil
}

// validateSetupRequest validates the setup request parameters.
// It checks for empty module IDs, image data presence, file format validity,
// and ensures the file format is in the allowed list of image formats.
func (s *Server) validateSetupRequest(req *SetupRequest) error {
	if strings.TrimSpace(req.ModuleId) == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	if len(req.Image) == 0 {
		return fmt.Errorf("image data cannot be empty")
	}

	if strings.TrimSpace(req.Fileformat) == "" {
		return fmt.Errorf("file format cannot be empty")
	}

	// Validate common image formats
	allowedFormats := []string{"png", "jpg", "jpeg", "gif", "webp", "svg"}
	format := strings.ToLower(strings.Split(strings.TrimSpace(req.Fileformat), "/")[1])
	for _, allowed := range allowedFormats {
		if format == allowed {
			return nil
		}
	}

	return fmt.Errorf("unsupported file format: %s", req.Fileformat)
}

// validateDeleteRequest validates the delete request parameters.
// It ensures the module ID is not empty or whitespace-only.
func (s *Server) validateDeleteRequest(req *DeleteRequest) error {
	if strings.TrimSpace(req.ModuleId) == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	return nil
}

// moduleNameExists checks if a module with the given name already exists.
// It returns true if a module with the specified name is found in the database.
func (s *Server) moduleNameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM modules WHERE name = $1)`

	if err := db.DB.GetContext(ctx, &exists, query, name); err != nil {
		return false, fmt.Errorf("failed to check module name existence: %w", err)
	}

	return exists, nil
}

// moduleIDExists checks if a module with the given ID exists.
// It returns true if a module with the specified ID is found in the database.
func (s *Server) moduleIDExists(ctx context.Context, moduleID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS (SELECT 1 FROM modules WHERE module_id = $1)`

	if err := db.DB.GetContext(ctx, &exists, query, moduleID); err != nil {
		return false, fmt.Errorf("failed to check module ID existence: %w", err)
	}

	return exists, nil
}

// createModule inserts a new module into the database.
// It creates a module record with the provided name and IP:port combination,
// returning the generated module ID.
func (s *Server) createModule(ctx context.Context, name, ip string, port int32) (string, error) {
	var moduleID string
	query := `INSERT INTO modules (name, ip_port) VALUES ($1, $2) RETURNING module_id`
	ipPort := fmt.Sprintf("%s:%d", ip, port)

	if err := db.DB.GetContext(ctx, &moduleID, query, name, ipPort); err != nil {
		return "", fmt.Errorf("failed to insert module: %w", err)
	}

	return moduleID, nil
}

// setupModuleImage inserts or updates module image data.
// It performs an upsert operation to handle both new image uploads and updates
// to existing module images, including updating the timestamp.
func (s *Server) setupModuleImage(ctx context.Context, moduleID string, image []byte, fileFormat string) error {
	query := `INSERT INTO images (module_id, image, fileformat) VALUES ($1, $2, $3)
		ON CONFLICT (module_id) DO UPDATE SET 
			image = EXCLUDED.image, 
			fileformat = EXCLUDED.fileformat,
			updated_at = CURRENT_TIMESTAMP`

	result, err := db.DB.ExecContext(ctx, query, moduleID, image, fileFormat)
	if err != nil {
		return fmt.Errorf("failed to setup module image: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows affected during image setup")
	}

	return nil
}

// deleteModule removes a module from the database.
// It deletes the module record and returns true if a module was actually deleted,
// false if no module with the given ID was found.
func (s *Server) deleteModule(ctx context.Context, moduleID string) (bool, error) {
	query := `DELETE FROM modules WHERE module_id = $1`

	result, err := db.DB.ExecContext(ctx, query, moduleID)
	if err != nil {
		return false, fmt.Errorf("failed to delete module: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected > 0, nil
}
