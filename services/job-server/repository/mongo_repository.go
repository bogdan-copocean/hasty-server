package repository

import (
	"context"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
)

type MongoRepository interface {
	SetJob(*events.JobEvent, context.Context) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

func (repo *mongoRepository) SetJob(jobEvent *events.JobEvent, ctx context.Context) error {

	jobIdUuid, err := uuid.New()
	if err != nil {
		return err
	}

	_, err = repo.collection.InsertOne(ctx, bson.M{"jobId": jobIdUuid, "objectId": jobEvent.Job.ObjectId, "sleepTimeUsed": jobEvent.SleepTimeUsed, "status": "Finished"})
	if err != nil {
		return err
	}

	return nil
}
