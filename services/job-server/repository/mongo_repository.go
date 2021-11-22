package repository

import (
	"context"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository interface {
	SetJob(*events.JobEvent) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

func (repo *mongoRepository) SetJob(jobEvent *events.JobEvent) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := repo.collection.InsertOne(ctx, bson.M{"jobId": jobEvent.Job.JobId, "objectId": jobEvent.Job.ObjectId, "sleepTimeUsed": jobEvent.SleepTimeUsed, "status": jobEvent.Job.Status})
	if err != nil {
		return err
	}

	return nil
}
