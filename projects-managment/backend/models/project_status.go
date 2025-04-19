package models

import (
	"database/sql"
	"time"
)

type ProjectStatus struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type ProjectStatusService struct {
	DB *sql.DB
}

func (s *ProjectStatusService) GetAll() ([]ProjectStatus, error) {
	rows, err := s.DB.Query("SELECT id, name FROM project_statuses ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statuses []ProjectStatus
	for rows.Next() {
		var status ProjectStatus
		if err := rows.Scan(&status.ID, &status.Name); err != nil {
			return nil, err
		}
		statuses = append(statuses, status)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return statuses, nil
}
