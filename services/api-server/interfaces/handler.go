package interfaces

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bogdan-copocean/hasty-server/services/api-server/app"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/publishers"
)

type HandlerInterface interface {
	PostHandler(w http.ResponseWriter, r *http.Request)
	PutHandler(w http.ResponseWriter, r *http.Request)
	GetHandler(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	apiService app.ApiService
	natsPub    publishers.NatsPublisherInterface
}

func NewHandler(apiService app.ApiService, natsPub publishers.NatsPublisherInterface) HandlerInterface {
	return &handler{apiService: apiService, natsPub: natsPub}
}

func (handler *handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 800*time.Millisecond)
	defer cancel()

	doneCh := make(chan struct{})
	errCh := make(chan error)

	objectIdMap := map[string]string{}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err := json.Unmarshal(body, &objectIdMap); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		res, _ := json.Marshal(fmt.Errorf("unmarshal error: %v", err.Error()))
		w.Write(res)
		return
	}
	objectId, ok := objectIdMap["object_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("could not get objectId from map"))
		return
	}

	job, err := handler.apiService.ProcessJob(objectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	eventJob := events.JobEvent{
		Subject: "job:created",
		Job:     job,
	}

	go handler.natsPub.PublishData(&eventJob, doneCh, errCh)

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
