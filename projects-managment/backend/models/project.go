// filepath: /home/ernesto/code/system-design/projects-managment/backend/models/project.go
package models

import (
	"database/sql"
	"net/http"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedBy   int       `json:"created_by"` // User ID who created the project
}

// ProjectService handles database operations for projects
type ProjectService struct {
	DB *sql.DB
}

// ProjectError defines custom error types for project operations
type ProjectError struct {
	Code    int
	Message string
}

func (e *ProjectError) Error() string {
	return e.Message
}

var (
	ErrProjectNotFound = &ProjectError{Code: http.StatusNotFound, Message: "project not found"}
)

func (ps *ProjectService) CreateProject(project Project) (Project, error) {
	now := time.Now()

	var id int
	err := ps.DB.QueryRow(`
		INSERT INTO projects (name, description, status_id, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id
	`, project.Name, project.Description, 1, project.CreatedBy, now, now).Scan(&id)

	if err != nil {
		return project, err
	}

	// Set the returned ID
	project.ID = id
	project.CreatedAt = now
	project.UpdatedAt = now

	return project, nil
}
