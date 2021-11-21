package repository

import (
	"context"

	"github.com/bogdan-copocean/hasty-server/services/job-server/events"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
)

type MongoRepository interface {
	// GetByJobId(jobId string, ctx context.Context) (*events.JobEvent, error)
	SetJob(*events.JobEvent, context.Context) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

// func (repo *mongoRepository) GetByJobId(jobId string, ctx context.Context) (*events.JobEvent, error) {
// 	// job := domain.Job{}

// 	// if err := repo.collection.FindOne(ctx, bson.M{"jobId": jobId}).Decode(&job); err != nil {
// 	// 	return nil, err
// 	// }

// 	// return &job, nil
// }

func (repo *mongoRepository) SetJob(jobEvent *events.JobEvent, ctx context.Context) error {

	jobIdUuid, err := uuid.New()
	if err != nil {
		return err
	}

	_, err = repo.collection.InsertOne(ctx, bson.M{"jobId": jobIdUuid, "objectId": jobEvent.Job.ObjectId, "sleepTimeUsed": jobEvent.SleepTimeUsed, "status": "DONE"})
	if err != nil {
		return err
	}

	return nil
}
