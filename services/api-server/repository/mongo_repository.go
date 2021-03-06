package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoRepository interface {
	GetJobByJobId(jobId string) (*domain.Job, error)
	GetJobByObjectId(objectId string) (*domain.Job, error)
	SetJob(job *domain.Job) error
	UpdateJobStatusAndTimeSlept(job *domain.Job) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

func (repo *mongoRepository) GetJobByObjectId(objectId string) (*domain.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	job := domain.Job{}

	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	if err := repo.collection.FindOne(ctx, bson.M{"objectId": objectId}, opts).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (repo *mongoRepository) GetJobByJobId(jobId string) (*domain.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	job := domain.Job{}

	opts := options.FindOne().SetSort(bson.M{"timestamp": -1})
	if err := repo.collection.FindOne(ctx, bson.M{"jobId": jobId}, opts).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (repo *mongoRepository) SetJob(job *domain.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := repo.collection.InsertOne(ctx, bson.M{
		"jobId":         job.JobId,
		"objectId":      job.ObjectId,
		"status":        job.Status,
		"timestamp":     job.Timestamp,
		"sleepTimeUsed": job.SleepTimeUsed,
	})

	if err != nil {
		return err
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("could not type assert oid")
	}
	job.Id = oid.Hex()

	return nil
}

func (repo *mongoRepository) UpdateJobStatusAndTimeSlept(job *domain.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := repo.collection.FindOneAndUpdate(ctx, bson.M{"jobId": job.JobId}, bson.M{"$set": bson.M{"status": job.Status, "sleepTimeUsed": job.SleepTimeUsed}}).Err(); err != nil {
		return err
	}

	return nil
}
