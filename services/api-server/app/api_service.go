package app

import (
	"fmt"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/domain"
	"github.com/bogdan-copocean/hasty-server/services/api-server/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

type ApiService interface {
	ProcessJob(objectId string) (*domain.Job, error)
	UpdateJob(job *domain.Job) error
	GetJob(objectId string) (*domain.Job, error)
}

type apiService struct {
	mongoRepo repository.MongoRepository
}

func NewApiService(mongoRepo repository.MongoRepository) ApiService {
	return &apiService{mongoRepo: mongoRepo}
}

func (as *apiService) ProcessJob(objectId string) (*domain.Job, error) {
	now := time.Now().Unix()

	foundJob, err := as.mongoRepo.GetJobByObjectId(objectId)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, fmt.Errorf("error while getting document: %v", err.Error())
	}

	if foundJob != nil {
		timePassed := now - foundJob.Timestamp

		if timePassed < 300 {
			return nil, fmt.Errorf("5 minutes not passed yet")
		}

		foundJob.JobId = uuid.New().String()
		foundJob.Status = "processing"
		foundJob.Timestamp = now

		if err = as.mongoRepo.SetJob(foundJob); err != nil {
			return nil, err
		}

		return foundJob, nil
	}

	newJob := domain.Job{}

	newJob.JobId = uuid.New().String()
	newJob.Status = "processing"
	newJob.Timestamp = now
	newJob.ObjectId = objectId

	if err = as.mongoRepo.SetJob(&newJob); err != nil {
		return nil, err
	}

	return &newJob, nil
}

func (as *apiService) UpdateJob(job *domain.Job) error {
	if err := as.mongoRepo.UpdateJobStatus(job); err != nil {
		return err
	}
	return nil
}

func (as *apiService) GetJob(objectId string) (*domain.Job, error) {
	job, err := as.mongoRepo.GetJobByObjectId(objectId)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no job with id: %v", objectId)
		}
		return nil, err
	}
	job.Id = ""

	return job, nil
}
