package db

import (
	"context"
	"errors"
	"fmt"
	"song-recognition/models"
	"song-recognition/utils"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	client *mongo.Client
}

func NewMongoClient(uri string) (*MongoClient, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %s", err)
	}
	return &MongoClient{client: client}, nil
}

func (db *MongoClient) Close() error {
	if db.client != nil {
		return db.client.Disconnect(context.Background())
	}
	return nil
}

func (db *MongoClient) StoreFingerprints(fingerprints map[uint32]models.Couple) error {
	collection := db.client.Database("song-recognition").Collection("fingerprints")

	for address, couple := range fingerprints {
		filter := bson.M{"_id": address}
		update := bson.M{
			"$push": bson.M{
				"couples": bson.M{
					"anchorTimeMs": couple.AnchorTimeMs,
					"id":           couple.SongID,
				},
			},
		}
		opts := options.Update().SetUpsert(true)

		_, err := collection.UpdateOne(context.Background(), filter, update, opts)
		if err != nil {
			return fmt.Errorf("error upserting document: %s", err)
		}
	}

	return nil
}

func (db *MongoClient) GetCouples(addresses []uint32) (map[uint32][]models.Couple, error) {
	collection := db.client.Database("song-recognition").Collection("fingerprints")

	couples := make(map[uint32][]models.Couple)

	for _, address := range addresses {
		// Find the document corresponding to the address
		var result bson.M
		err := collection.FindOne(context.Background(), bson.M{"_id": address}).Decode(&result)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				continue
			}
			return nil, fmt.Errorf("error retrieving document for address %d: %s", address, err)
		}

		// Extract couples from the document and append them to the couples map
		var docCouples []models.Couple
		couplesList, ok := result["couples"].(primitive.A)
		if !ok {
			return nil, fmt.Errorf("couples field in document for address %d is not valid", address)
		}

		for _, item := range couplesList {
			itemMap, ok := item.(primitive.M)
			if !ok {
				return nil, fmt.Errorf("invalid couple format in document for address %d", address)
			}

			couple := models.Couple{
				AnchorTimeMs: uint32(itemMap["anchorTimeMs"].(int64)),
				SongID:       uint32(itemMap["id"].(int64)),
			}
			docCouples = append(docCouples, couple)
		}
		couples[address] = docCouples
	}

	return couples, nil
}

func (db *MongoClient) TotalSongs() (int, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")
	total, err := existingSongsCollection.CountDocuments(context.Background(), bson.D{})
	if err != nil {
		return 0, err
	}

	return int(total), nil
}

func (db *MongoClient) SongExistsByID(phiZoneID string) (bool, error) {
	existingCollection := db.client.Database("song-recognition").Collection("songs")
	filter := bson.D{{Key: "PhiZoneID", Value: phiZoneID}}

	err := existingCollection.FindOne(context.Background(), filter).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *MongoClient) FindNonExistentSongs(requestedIDs []string) ([]string, error) {
	existingCollection := db.client.Database("song-recognition").Collection("songs")

	filter := bson.M{"PhiZoneID": bson.M{"$in": requestedIDs}}
	cursor, err := existingCollection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	foundIDs := make(map[string]bool)
	for cursor.Next(context.Background()) {
		var result struct {
			PhiZoneID string `bson:"PhiZoneID"`
		}
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		foundIDs[result.PhiZoneID] = true
	}

	var nonExistentIDs []string
	for _, id := range requestedIDs {
		if !foundIDs[id] {
			nonExistentIDs = append(nonExistentIDs, id)
		}
	}

	return nonExistentIDs, nil
}

func (db *MongoClient) RegisterSong(songTitle, songArtist, songID string) (uint32, error) {
	existingSongsCollection := db.client.Database("song-recognition").Collection("songs")

	// Create a compound unique index on songID and key, if it doesn't already exist
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "songID", Value: 1}, {Key: "key", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := existingSongsCollection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		return 0, fmt.Errorf("failed to create unique index: %v", err)
	}

	// Attempt to insert the song with songID and key
	id := utils.GenerateUniqueID()
	key := utils.GenerateSongKey(songTitle, songArtist)
	_, err = existingSongsCollection.InsertOne(context.Background(), bson.M{"_id": id, "key": key, "songID": songID})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, fmt.Errorf("song with songID or key already exists: %v", err)
		} else {
			return 0, fmt.Errorf("failed to register song: %v", err)
		}
	}

	return id, nil
}

var mongofilterKeys = "_id | songID | key"

func (db *MongoClient) GetSong(filterKey string, value interface{}) (s Song, songExists bool, e error) {
	if !strings.Contains(mongofilterKeys, filterKey) {
		return Song{}, false, errors.New("invalid filter key")
	}

	songsCollection := db.client.Database("song-recognition").Collection("songs")
	var song bson.M

	filter := bson.M{filterKey: value}

	err := songsCollection.FindOne(context.Background(), filter).Decode(&song)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return Song{}, false, nil
		}
		return Song{}, false, fmt.Errorf("failed to retrieve song: %v", err)
	}

	songID := song["songID"].(string)
	title := strings.Split(song["key"].(string), "---")[0]
	artist := strings.Split(song["key"].(string), "---")[1]

	songInstance := Song{title, artist, songID}

	return songInstance, true, nil
}

func (db *MongoClient) GetSongByID(id uint32) (Song, bool, error) {
	return db.GetSong("_id", id)
}

func (db *MongoClient) GetSongByYTID(songID string) (Song, bool, error) {
	return db.GetSong("songID", songID)
}

func (db *MongoClient) GetSongByKey(key string) (Song, bool, error) {
	return db.GetSong("key", key)
}

func (db *MongoClient) DeleteSongByID(id uint32) error {
	songsCollection := db.client.Database("song-recognition").Collection("songs")

	filter := bson.M{"_id": id}

	_, err := songsCollection.DeleteOne(context.Background(), filter)
	if err != nil {
		return fmt.Errorf("failed to delete song: %v", err)
	}

	return nil
}

func (db *MongoClient) DeleteCollection(collectionName string) error {
	collection := db.client.Database("song-recognition").Collection(collectionName)
	err := collection.Drop(context.Background())
	if err != nil {
		return fmt.Errorf("error deleting collection: %v", err)
	}
	return nil
}
