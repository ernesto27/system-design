package models

import "database/sql"

type Role struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
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
