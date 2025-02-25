package db

import (
	"fmt"
	"song-recognition/models"
	"song-recognition/utils"
)

type DBClient interface {
	Close() error
	StoreFingerprints(fingerprints map[uint32]models.Couple) error
	GetCouples(addresses []uint32) (map[uint32][]models.Couple, error)
	TotalSongs() (int, error)
	SongExistsByID(uuid string) (bool, error)
	FindNonExistentSongs(requestedIDs []string) ([]string, error)
	RegisterSong(songTitle, songArtist, ytID string) (string, error)
	GetSong(filterKey string, value interface{}) (Song, bool, error)
	GetSongByID(songID string) (Song, bool, error)
	GetSongByUUID(uuid string) (Song, bool, error)
	GetSongByYTID(ytID string) (Song, bool, error)
	GetSongByKey(key string) (Song, bool, error)
	DeleteSongByID(songID string) error
	DeleteCollection(collectionName string) error
	DeleteSongByUUID(uuid string) error

	RegisterCopyrightSong(songTitle, songArtist, songID string) (string, error)
	TotalCopyrightSongs() (int, error)
	CopyrightSongExistsByID(uuid string) (bool, error)
	GetCopyrightSong(filterKey string, value interface{}) (Song, bool, error)
	GetCopyrightSongByID(songId string) (Song, bool, error)
	GetCopyrightSongByUUID(uuid string) (Song, bool, error)
	GetCopyrightSongByKey(key string) (Song, bool, error)
	DeleteCopyrightSongByID(songID string) error
	DeleteCopyrightSongByUUID(uuid string) error
	DeleteCopyrightCollection(collectionName string) error
}

type Song struct {
	Title  string
	Artist string
	SongID string
}

var DBtype = utils.GetEnv("DB_TYPE", "mongo")

func NewDBClient() (DBClient, error) {
	switch DBtype {
	case "mongo":
		var (
			dbUsername = utils.GetEnv("DB_USER")
			dbPassword = utils.GetEnv("DB_PASS")
			dbName     = utils.GetEnv("DB_NAME")
			dbHost     = utils.GetEnv("DB_HOST")
			dbPort     = utils.GetEnv("DB_PORT")
			dbUri      = utils.GetEnv("DB_URI")
		)
		if dbUri == "" {
			if dbUsername == "" || dbPassword == "" {
				dbUri = "mongodb://localhost:27017"
			} else {
				dbUri = "mongodb://" + dbUsername + ":" + dbPassword + "@" + dbHost + ":" + dbPort + "/" + dbName
			}
		}
		return NewMongoClient(dbUri)

	default:
		return nil, fmt.Errorf("unsupported database type: %s", DBtype)
	}
}
