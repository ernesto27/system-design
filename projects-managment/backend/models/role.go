package models

import (
	"database/sql"
	"net/http"
)

type Role struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Percentage int    `json:"percentage"`
}

type RoleService struct {
	DB *sql.DB
}

func (rs *RoleService) GetAll() ([]Role, error) {
	rows, err := rs.DB.Query("SELECT id, name FROM roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (rs *RoleService) GetRole(id int) (Role, error) {
	var role Role

	err := rs.DB.QueryRow("SELECT id, name FROM roles WHERE id = $1", id).Scan(&role.ID, &role.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return role, ErrRoleNotFound
		}
		return role, err
	}

	return role, nil
}

func (rs *RoleService) GetProjectRoles(projectID int) ([]Role, error) {
	rows, err := rs.DB.Query(`
		SELECT 
			r.id, 
			r.name, 
			pr.percentage
		FROM projects_roles pr
		JOIN roles r ON pr.role_id = r.id
		WHERE pr.project_id = $1
	`, projectID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var role Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Percentage); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// Error for when a role isn't found
var ErrRoleNotFound = &RoleError{Code: http.StatusNotFound, Message: "role not found"}

type RoleError struct {
	Code    int
	Message string
}

func (e *RoleError) Error() string {
	return e.Message
}
