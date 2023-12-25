package newsfeed

import (
	"encoding/json"
	"feedsystem/cache"
	"feedsystem/db"
	"feedsystem/types"
	"fmt"
	"strconv"
	"strings"
)

type newsFeed struct {
	userID int
	Cache  *cache.Redis
	Db     *db.Mysql
}

func New(userID int, cache *cache.Redis, db *db.Mysql) *newsFeed {
	return &newsFeed{
		userID: userID,
		Cache:  cache,
		Db:     db,
	}
}

func (n *newsFeed) SaveCache(data string) error {
	// get followers ids
	ids, err := n.GetFollowersIDs()
	if err != nil {
		return err
	}

	// save post userid postid on cache newsfeed:userid
	for _, id := range ids {
		value := fmt.Sprint(n.userID) + ":" + data
		err = n.Cache.SetList("feed:"+fmt.Sprint(id), value)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (n *newsFeed) GetPostsCache(page int) ([]types.Post, error) {
	const limit = 3
	var start int64 = int64((page - 1) * limit)
	var stop int64 = int64(((page - 1) * limit) + limit - 1)

	resp := []types.Post{}
	userPost, err := n.Cache.GetList("feed:"+fmt.Sprint(n.userID), start, stop)
	if err != nil {
		return resp, err
	}

	for _, p := range userPost {
		d := strings.Split(p, ":")
		if len(d) == 2 {
			userID, err := strconv.Atoi(d[0])
			if err != nil {
				fmt.Println(err)
				continue
			}
			postID, err := strconv.Atoi(d[1])
			if err != nil {
				fmt.Println(err)
				continue
			}

			// Get content post
			post, err := n.Cache.Get("post:" + fmt.Sprint(postID))
			if err != nil {
				fmt.Println(err)
				continue
			}

			var p types.Post
			err = json.Unmarshal([]byte(post), &p)
			if err != nil {
				fmt.Println(err)
				continue
			}

			resp = append(resp, types.Post{
				ID:     postID,
				UserID: userID,
				Text:   p.Text,
			})
		}

	}

	return resp, nil
}

func (n *newsFeed) GetFollowersIDs() ([]int, error) {
	ids, err := n.Db.GetFollwers(n.userID)
	if err != nil {
		return nil, err
	}

	return ids, err
}
