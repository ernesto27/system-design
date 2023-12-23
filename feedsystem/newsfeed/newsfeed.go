package newsfeed

import (
	"feedsystem/cache"
	"feedsystem/db"
	"fmt"
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
		err = n.Cache.ListSet("feed:"+fmt.Sprint(id), value)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil

}

func (n *newsFeed) GetFollowersIDs() ([]int, error) {
	ids, err := n.Db.GetFollwers(n.userID)
	if err != nil {
		return nil, err
	}

	return ids, err
}
