package interfaces

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/bogdan-copocean/hasty-server/services/api-server/app"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events"
	"github.com/bogdan-copocean/hasty-server/services/api-server/events/publishers"
	"github.com/go-chi/chi/v5"
	"github.com/unrolled/render"
)

type ApiHandlerInterface interface {
	PostHandler(w http.ResponseWriter, r *http.Request)
	PutHandler(w http.ResponseWriter, r *http.Request)
	GetHandler(w http.ResponseWriter, r *http.Request)
}

type apiHandler struct {
	apiService        app.ApiService
	jobEventPublisher publishers.JobEventPublisher
}

func NewApiHandler(apiService app.ApiService, jobEventPublisher publishers.JobEventPublisher) ApiHandlerInterface {
	return &apiHandler{apiService: apiService, jobEventPublisher: jobEventPublisher}
}

func (handler *apiHandler) PostHandler(w http.ResponseWriter, r *http.Request) {
	render := render.New()
	w.Header().Set("Content-Type", "application/json")

	objectIdMap := map[string]string{}

	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err := json.Unmarshal(body, &objectIdMap); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
		return
	}
	objectId, ok := objectIdMap["object_id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, http.StatusBadRequest, map[string]string{
			"message": "you must provide an object_id",
		})
		return
	}

	job, err := handler.apiService.ProcessJob(objectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
		return
	}

	eventJob := events.JobEvent{
		Subject: "job:created",
		Job:     job,
	}

	if err := handler.jobEventPublisher.PublishData(&eventJob); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusCreated)
	render.JSON(w, http.StatusCreated, map[string]interface{}{
		"message": job,
	})

}

func (handler *apiHandler) PutHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Put"))
}

func (handler *apiHandler) GetHandler(w http.ResponseWriter, r *http.Request) {
	render := render.New()
	w.Header().Set("Content-Type", "application/json")

	objectId := chi.URLParam(r, "objectId")

	job, err := handler.apiService.GetJob(objectId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render.JSON(w, http.StatusBadRequest, map[string]string{
			"message": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	render.JSON(w, http.StatusOK, map[string]interface{}{
		"message": job,
	})
}
