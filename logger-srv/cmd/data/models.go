package data

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

var client *mongo.Client

func New(mongo *mongo.Client) Models {
	client = mongo
	return Models{LogEntry: LogEntry{}}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        string    `json:"id,omitempty" bson:"_id,omitempty"`
	Name      string    `json:"name" bson:"name"`
	Data      string    `json:"data" bson:"data"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		log.Printf("error inserting into logs :\n%v", err)
		return err
	}
	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	collection := client.Database("logs").Collection("logs")
	ops := options.Find()
	ops.SetSort(bson.D{{"created_at", -1}})
	cursor, err := collection.Find(context.TODO(), bson.D{}, ops)
	if err != nil {
		log.Printf("finding all logs error :%v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry
		err := cursor.Decode(&logs)
		if err != nil {
			log.Printf("Error decoding logs into slice %v", err)
			return nil, err
		}
		logs = append(logs, &item)
	}
	return logs, err
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	collection := client.Database("logs").Collection("logs")

	docId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	var entry LogEntry
	err = collection.FindOne(ctx, bson.M{"_id": docId}).Decode(&entry)
	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	collection := client.Database("logs").Collection("logs")

	if err := collection.Drop(ctx); err != nil {
		return err
	}
	return nil
}

func (l *LogEntry) Update() (*mongo.UpdateResult, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancelFunc()

	collection := client.Database("logs").Collection("logs")
	docId, err := primitive.ObjectIDFromHex(l.ID)
	if err != nil {
		return nil, err
	}
	updateRes, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": docId},
		bson.D{
			{
				"$set", bson.D{
					{"name", l.Name},
					{"data", l.Data},
					{"updated_at", time.Now()},
				},
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return updateRes, nil
}
