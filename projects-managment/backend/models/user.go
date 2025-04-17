package models

import (
	"database/sql"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserService struct {
	DB *sql.DB
}

type UserNotFoundError struct {
	Code    int
	Message string
}

func (e *UserNotFoundError) Error() string {
	return e.Message
}

var (
	ErrUserNotFound = &UserNotFoundError{Code: http.StatusNotFound, Message: "user not found"}
)

func (userService *UserService) Login(email, password string) (User, error) {
	row := userService.DB.QueryRow(`
	SELECT id, password_hash
	FROM users WHERE email=$1`, email)

	var user User
	err := row.Scan(&user.ID, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, ErrUserNotFound
		}
		return user, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return user, err
	}

	return user, nil
}
