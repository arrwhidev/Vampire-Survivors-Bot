package main

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
)

//Checks wether message author has administrator permissions
func IsAdmin(s *discordgo.Session, m *discordgo.MessageCreate) (bool, error) {
	perm, err := s.UserChannelPermissions(m.Author.ID, m.ChannelID)
	if err != nil {
		return false, err
	}
	return perm&discordgo.PermissionAdministrator != 0, nil
}

//Creating buckets
func CreateBuckets() {
	database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Chans"))
		if err != nil {
			return fmt.Errorf("creation Error: %s", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte("Library"))
		if err != nil {
			return fmt.Errorf("creation Error: %s", err)
		}
		return nil
	})
}
