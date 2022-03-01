package main

import (
	"encoding/json"
	"fmt"

	"github.com/boltdb/bolt"
)

type Channel struct {
	Id     string `json:"id"`
	Prefix string `json:"prefix"`
}

type DatabaseHandler struct {
	Bot    *VampBot
	Bolt   *bolt.DB
	Chans  map[string]Channel
	Guilds map[string]bool
}

func MakeDatabaseHandler(bot *VampBot, path string) *DatabaseHandler {
	db := &DatabaseHandler{Bot: bot}
	db.Bolt, _ = bolt.Open(path, 0600, nil)
	return db
}

//Creating buckets
func (db *DatabaseHandler) CreateBuckets() {
	db.Bolt.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Chans"))
		if err != nil {
			return fmt.Errorf("creation Error: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Guilds"))
		if err != nil {
			return fmt.Errorf("creation Error: %s", err)
		}
		return nil
	})
}

//Loading channels from database
func (db *DatabaseHandler) LoadChannels() {
	db.Chans = make(map[string]Channel)
	db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Chans"))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var ch Channel
			err := json.Unmarshal(v, &ch)
			if err != nil {
				db.Bot.Logger.Printf("[SETUP] Error getting channel %s", k)
				continue
			}
			db.Chans[string(k)] = ch
		}
		return nil
	})
}

//Loading guilds from database
func (db *DatabaseHandler) LoadGuilds() {
	db.Guilds = make(map[string]bool)
	db.Bolt.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Guilds"))
		c := b.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			db.Guilds[string(k)] = true
		}
		return nil
	})
}

//Adding channel to database
func (db *DatabaseHandler) CreateChan(id, prefix string) (Channel, error) {
	ch := Channel{id, prefix}
	v, _ := json.Marshal(ch)
	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Chans"))
		err := b.Put([]byte(id), v)
		return err
	})
	if err != nil {
		db.Bot.Logger.Printf("[BOT] Failed adding channel %s Error: %v", id, err)
	} else {
		db.Bot.Logger.Printf("[BOT] Added channel %s", id)
	}
	return ch, err
}

//Adding guild to database
func (db *DatabaseHandler) CreateGuild(id string) error {
	v, _ := json.Marshal(true)
	err := db.Bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Guilds"))
		err := b.Put([]byte(id), v)
		return err
	})
	return err
}
