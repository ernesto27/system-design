package models

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type Project struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Status         ProjectStatus `json:"status"`
	TimeEstimation int           `json:"timeEstimation"`
	CreatedAt      string        `json:"createdAt"`
	UpdatedAt      string        `json:"updatedAt"`
	CreatedBy      int           `json:"createdBy"`
	Roles          []Role        `json:"roles"`
}

type ProjectService struct {
	DB *sql.DB
}

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

func (ps *ProjectService) Create(project Project) (Project, error) {
	now := time.Now()

	// Begin transaction
	tx, err := ps.DB.Begin()
	if err != nil {
		return project, fmt.Errorf("failed to begin transaction: %v", err)
	}

	// Defer a rollback in case anything fails
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var id int
	err = tx.QueryRow(`
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

	projectRole := ProjectRoleService{tx: tx}

	for _, role := range project.Roles {
		err = projectRole.Create(ProjectRole{
			ProjectID:  id,
			RoleID:     role.ID,
			Percentage: role.Percentage,
		})
		if err != nil {
			return project, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return project, fmt.Errorf("failed to commit transaction: %v", err)
	}

	project.ID = id

	return project, nil
}

func (ps *ProjectService) GetByID(id int) (Project, error) {
	rows, err := ps.DB.Query(`
		SELECT 
			p.id, 
			p.name, 
			p.description, 
			ps.id AS status_id,
			ps.name AS status_name, 
			p.time_estimation,
			p.created_at, 
			p.updated_at,
			p.created_user_id,
			r.id AS role_id,
			r.name AS role_name,
			pr.percentage
		FROM projects p
		JOIN project_statuses ps ON p.status_id = ps.id
		LEFT JOIN projects_roles pr ON p.id = pr.project_id
		LEFT JOIN roles r ON pr.role_id = r.id
		WHERE p.id = $1
	`, id)

	if err != nil {
		return Project{}, err
	}
	defer rows.Close()

	var project Project
	var found bool

	for rows.Next() {
		found = true
		var (
			timeEstimation, createdUserID int
			createdAt, updatedAt          time.Time
			roleID                        sql.NullInt64
			roleName                      sql.NullString
			rolePercentage                sql.NullInt64
		)

		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status.ID,
			&project.Status.Name,
			&timeEstimation,
			&createdAt,
			&updatedAt,
			&createdUserID,
			&roleID,
			&roleName,
			&rolePercentage,
		)

		if err != nil {
			return Project{}, err
		}

		// Convert time fields to string format
		project.CreatedAt = createdAt.Format(time.RFC3339)
		project.UpdatedAt = updatedAt.Format(time.RFC3339)
		project.TimeEstimation = timeEstimation
		project.CreatedBy = createdUserID
	}

	if err = rows.Err(); err != nil {
		return Project{}, err
	}

	if !found {
		return Project{}, ErrProjectNotFound
	}

	roleService := RoleService{
		DB: ps.DB,
	}

	roles, err := roleService.GetProjectRoles(id)
	if err != nil {
		return Project{}, fmt.Errorf("error fetching roles for project %d: %w", id, err)
	}

	project.Roles = roles

	return project, nil
}

func (ps *ProjectService) GetAllProjects(page, limit int) ([]Project, error) {
	offset := (page - 1) * limit
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
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var project Project
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status.Name,
			&project.TimeEstimation,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.CreatedBy,
		)

		if err != nil {
			return nil, err
		}

		roleService := RoleService{
			DB: ps.DB,
		}

		roles, err := roleService.GetProjectRoles(project.ID)
		if err != nil {
			return projects, fmt.Errorf("error fetching roles for project %d: %w", project.ID, err)
		}

		project.Roles = roles
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return projects, nil
}

func (ps *ProjectService) CountProjects() (int, error) {
	var count int
	err := ps.DB.QueryRow(`SELECT COUNT(*) FROM projects`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error counting projects: %w", err)
	}
	return count, nil
}

func (ps *ProjectService) Update(id int, project Project) (Project, error) {
	_, err := ps.GetByID(id)
	if err != nil {
		return Project{}, err
	}

	now := time.Now()

	tx, err := ps.DB.Begin()
	if err != nil {
		return project, fmt.Errorf("failed to begin transaction: %v", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Update the project
	_, err = tx.Exec(`
		UPDATE projects 
		SET name = $1, 
		    description = $2, 
		    status_id = $3,
		    time_estimation = $4,
		    updated_at = $5
		WHERE id = $6`,
		project.Name,
		project.Description,
		project.Status.ID,
		project.TimeEstimation,
		now,
		id)

	if err != nil {
		return project, fmt.Errorf("failed to update project: %v", err)
	}

	// Delete existing project roles
	_, err = tx.Exec("DELETE FROM projects_roles WHERE project_id = $1", id)
	if err != nil {
		return project, fmt.Errorf("failed to delete existing project roles: %v", err)
	}

	// Insert new project roles
	projectRole := ProjectRoleService{tx: tx}
	for _, role := range project.Roles {
		err = projectRole.Create(ProjectRole{
			ProjectID:  id,
			RoleID:     role.ID,
			Percentage: role.Percentage,
		})
		if err != nil {
			return project, err
		}
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return project, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return ps.GetByID(id)
}

type ProjectRole struct {
	ID         int `json:"id"`
	ProjectID  int `json:"projectId"`
	RoleID     int `json:"roleId"`
	Percentage int `json:"percentage"`
}

type ProjectRoleService struct {
	tx *sql.Tx
}

func (prs *ProjectRoleService) Create(projectRole ProjectRole) error {
	_, err := prs.tx.Exec(`
		INSERT INTO projects_roles (
			project_id, 
			role_id, 
			percentage
		)
		VALUES ($1, $2, $3)`,
		projectRole.ProjectID,
		projectRole.RoleID,
		projectRole.Percentage)

	if err != nil {
		return fmt.Errorf("error inserting project role: %v", err)
	}

	return nil
}
