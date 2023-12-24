package db

import (
	"database/sql"
	"errors"
	"feedsystem/types"
	"fmt"
	"time"

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

func (m *Mysql) CreateTweet(text string, userID int) (int64, time.Time, error) {
	r, err := m.Db.Exec("INSERT INTO tweets (text, user_id) VALUES (?, ?)", text, userID)
	if err != nil {
		return 0, time.Time{}, err
	}

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, time.Time{}, err
	}

	var createdAtBytes []byte
	err = m.Db.QueryRow("SELECT created_at FROM tweets WHERE id = ?", lastInsertID).Scan(&createdAtBytes)
	if err != nil {
		fmt.Println(err)
	}

	createdAt, err := time.Parse("2006-01-02 15:04:05", string(createdAtBytes))
	if err != nil {
		return 0, time.Time{}, err
	}

	return lastInsertID, createdAt, nil
}

func (m *Mysql) GetTweetsFollowing(userID int) ([]types.Post, error) {
	rows, err := m.Db.Query("select id, text, created_at from tweets t where user_id IN (SELECT f.follower_user_id FROM followers f WHERE f.following_user_id = ?)", userID)
	if err != nil {
		return nil, err
	}

	posts := []types.Post{}
	for rows.Next() {
		var p types.Post
		err = rows.Scan(&p.ID, &p.Text, &p.CreatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, p)
	}

	return posts, nil
}

func (m *Mysql) GetFollwers(userID int) ([]int, error) {
	rows, err := m.Db.Query("SELECT follower_user_id FROM followers WHERE following_user_id = ?", userID)
	if err != nil {
		return nil, err
	}

	followers := []int{}
	for rows.Next() {
		var follower int
		err = rows.Scan(&follower)
		if err != nil {
			return nil, err
		}

		followers = append(followers, follower)
	}

	return followers, nil
}

func (m *Mysql) Close() {
	m.Db.Close()
}
