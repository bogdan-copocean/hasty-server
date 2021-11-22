package interfaces

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bogdan-copocean/hasty-server/services/api-server/app"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/publishers"
	"github.com/go-chi/chi/v5"
)

type HandlerInterface interface {
	PostHandler(w http.ResponseWriter, r *http.Request)
	PutHandler(w http.ResponseWriter, r *http.Request)
	GetHandler(w http.ResponseWriter, r *http.Request)
}

type handler struct {
	apiService app.ApiService
	publisher  publishers.JobEventPublisher
}

func NewHandler(apiService app.ApiService, publisher publishers.JobEventPublisher) HandlerInterface {
	return &handler{apiService: apiService, publisher: publisher}
}

func (handler *handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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

	if err := handler.publisher.PublishData(&eventJob); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("created"))

}

func (handler *handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Put"))
}

func (handler *handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	objectId := chi.URLParam(r, "objectId")

	job, err := handler.apiService.GetJob(objectId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	res, _ := json.Marshal(job)
	w.Write(res)
}
