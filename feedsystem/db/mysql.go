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

type Tweet struct {
	ID        int    `json:"id"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

func (m *Mysql) CreateTweet(text string, userID int) (int64, error) {
	r, err := m.Db.Exec("INSERT INTO tweets (text, user_id) VALUES (?, ?)", text, userID)

	lastInsertID, err := r.LastInsertId()
	if err != nil {
		return 0, err
	}

	return lastInsertID, nil
}

func (m *Mysql) GetTweetsFollowing(userID int) ([]Tweet, error) {
	rows, err := m.Db.Query("select id, text, created_at from tweets t where user_id IN (SELECT f.follower_user_id FROM followers f WHERE f.following_user_id = ?)", userID)
	if err != nil {
		return nil, err
	}

	tweets := []Tweet{}
	for rows.Next() {
		var tweet Tweet
		err = rows.Scan(&tweet.ID, &tweet.Text, &tweet.CreatedAt)
		if err != nil {
			return nil, err
		}

		tweets = append(tweets, tweet)
	}

	return tweets, nil
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
