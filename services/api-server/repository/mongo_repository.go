package repository

import (
	"context"
	"errors"

	"github.com/bogdan-copocean/hasty-server/services/api-server/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver/uuid"
)

type MongoRepository interface {
	GetByJobId(jobId string, ctx context.Context) (*domain.Job, error)
	SetJob(job *domain.Job, ctx context.Context) error
}

type mongoRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewMongoRepository(client *mongo.Client, collection *mongo.Collection) MongoRepository {
	return &mongoRepository{client: client, collection: collection}
}

func (repo *mongoRepository) GetByJobId(jobId string, ctx context.Context) (*domain.Job, error) {
	job := domain.Job{}

	if err := repo.collection.FindOne(ctx, bson.M{"jobId": jobId}).Decode(&job); err != nil {
		return nil, err
	}

	return &job, nil
}

func (repo *mongoRepository) SetJob(job *domain.Job, ctx context.Context) error {

	jobIdUuid, err := uuid.New()
	if err != nil {
		return err
	}

	res, err := repo.collection.InsertOne(ctx, bson.M{"jobId": jobIdUuid, "objectId": job.ObjectId, "metadata": "metadata", "status": "pending"})
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
