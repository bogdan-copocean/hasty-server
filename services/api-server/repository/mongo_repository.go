package repository

import (
	"context"
	"errors"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/domain"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepository interface {
	GetByJobId(jobId string) (*domain.Job, error)
	SetJob(job *domain.Job) error
	UpdateJob(job *domain.Job) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

func (repo *mongoRepository) GetByJobId(jobId string) (*domain.Job, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	job := domain.Job{}

	if err := repo.collection.FindOne(ctx, bson.M{"jobId": jobId}).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (repo *mongoRepository) SetJob(job *domain.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	jobIdUuid := uuid.New().String()

	res, err := repo.collection.InsertOne(ctx, bson.M{"jobId": jobIdUuid, "objectId": job.ObjectId, "metadata": "metadata", "status": "pending"})
	if err != nil {
		return err
	}

	oid, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return errors.New("could not type assert oid")
	}
	job.Id = oid.Hex()
	job.JobId = jobIdUuid

	return nil
}

func (repo *mongoRepository) UpdateJob(job *domain.Job) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := repo.collection.FindOneAndUpdate(ctx, bson.M{"jobId": job.JobId}, bson.M{"$set": bson.M{"status": job.Status}}).Err(); err != nil {
		return err
	}

	return nil
}
