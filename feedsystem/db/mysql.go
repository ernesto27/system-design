package db

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	Db *sql.DB
}

func NewMysql(host, user, password, port, database string) (*Mysql, error) {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", user, password, host, port, database))
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, errors.New("error connecting to the database")
	}

	return &Mysql{
		Db: db,
	}, nil
}

func (m *Mysql) CreateTweet(text string, userID int) error {
	_, err := m.Db.Exec("INSERT INTO tweets (text, user_id) VALUES (?, ?)", text, userID)
	return err
}

func (m *Mysql) Close() {
	m.Db.Close()
}
