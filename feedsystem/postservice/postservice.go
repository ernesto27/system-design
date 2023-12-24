package postservice

import (
	"encoding/json"
	"feedsystem/cache"
	"feedsystem/db"
	"feedsystem/types"
	"fmt"
)

func Create(db *db.Mysql, cache *cache.Redis, rp types.Request) (int64, bool) {
	id, created, err := db.CreateTweet(rp.Text, rp.UserID)
	if err != nil {
		return 0, false
	}

	go func() {
		p := types.Post{
			ID:        int(id),
			UserID:    rp.UserID,
			Text:      rp.Text,
			CreatedAt: created,
		}

		j, err := json.Marshal(p)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = cache.Set("post:"+fmt.Sprint(id), string(j))
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("cache set")
		}
	}()

	return id, true
}
