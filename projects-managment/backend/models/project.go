// filepath: /home/ernesto/code/system-design/projects-managment/backend/models/project.go
package models

import (
	"database/sql"
	"net/http"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Status         string    `json:"status"`
	TimeEstimation int       `json:"time_estimation"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	CreatedBy      int       `json:"created_by"`
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
		INSERT INTO projects (
			name, 
			description, 
			status_id, 
			created_user_id,
			time_estimation,
			created_at, 
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
		project.Name,
		project.Description,
		1,
		project.CreatedBy,
		project.TimeEstimation,
		now,
		now).Scan(&id)

	if err != nil {
		return project, err
	}

	// Set the returned ID
	project.ID = id
	project.CreatedAt = now
	project.UpdatedAt = now

	return project, nil
}

// GetProject retrieves a project by ID
func (ps *ProjectService) GetProject(id int) (Project, error) {
	var project Project
	var statusName string

	err := ps.DB.QueryRow(`
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			ps.name, 
			p.time_estimation,
			p.created_at, 
			p.updated_at,
			p.created_user_id
		FROM projects p
		JOIN project_statuses ps ON p.status_id = ps.id
		WHERE p.id = $1`, id).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&statusName,
		&project.TimeEstimation,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.CreatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return project, ErrProjectNotFound
		}
		return project, err
	}

	project.Status = statusName
	return project, nil
}

// GetAllProjects retrieves all projects
func (ps *ProjectService) GetAllProjects() ([]Project, error) {
	rows, err := ps.DB.Query(`
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			ps.name, 
			p.time_estimation,
			p.created_at, 
			p.updated_at,
			p.created_user_id
		FROM projects p
		JOIN project_statuses ps ON p.status_id = ps.id
		ORDER BY p.id DESC
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		var statusName string

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&statusName,
			&project.TimeEstimation,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.CreatedBy,
		)

		if err != nil {
			return nil, err
		}

		project.Status = statusName
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}
