package interfaces

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/domain"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/publishers"
	"github.com/bogdan-copocean/hasty-server/services/api-server/repository"
)

type HandlerInterface interface {
	PostHandler(w http.ResponseWriter, r *http.Request)
	PutHandler(w http.ResponseWriter, r *http.Request)
	GetHandler(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	mongoRepository repository.MongoRepository
	natsPub         publishers.NatsPublisherInterface
}

func NewHandler(mongoRepository repository.MongoRepository, natsPub publishers.NatsPublisherInterface) HandlerInterface {
	return &handler{mongoRepository: mongoRepository, natsPub: natsPub}
}

func (handler *handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 800*time.Millisecond)
	defer cancel()

	doneCh := make(chan struct{})
	errCh := make(chan error)

	job := domain.Job{}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err := json.Unmarshal(body, &job); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(err.Error())
		w.Write(res)
		return
	}

	if err := handler.mongoRepository.SetJob(&job, ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		res, _ := json.Marshal(err.Error())
		w.Write(res)
		return
	}

	eventJob := events.JobEvent{
		Subject: "job:created",
		Job:     job,
	}

	go handler.natsPub.PublishData(eventJob, doneCh, errCh)

	for {
		select {
		case <-doneCh:
			w.WriteHeader(http.StatusCreated)
			res, _ := json.Marshal(job)
			w.Write(res)
			return
		case <-errCh:
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal([]byte("something went wrong. please rerun the job"))
			w.Write(res)
			return
		case <-ctx.Done():
			w.WriteHeader(http.StatusInternalServerError)
			res, _ := json.Marshal([]byte("timeout"))
			w.Write(res)
			return

		}
	}

}

func (handler *handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Put"))
}

func (handler *handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get"))
}
