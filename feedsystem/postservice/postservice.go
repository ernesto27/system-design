package postservice

import (
	"feedsystem/cache"
	"feedsystem/db"
	"feedsystem/types"
	"fmt"
)

func Create(db *db.Mysql, cache *cache.Redis, rp types.Request) bool {
	id, err := db.CreateTweet(rp.Text, rp.UserID)
	if err != nil {
		return false
	}

	go func() {
		err = cache.Set("posts."+fmt.Sprint(id), rp.Text)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("cache set")
		}
	}()

	return true
}
