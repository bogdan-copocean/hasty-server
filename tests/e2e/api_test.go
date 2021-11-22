package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

func setUpContainers() (*testcontainers.LocalDockerCompose, error) {
	composeFilePaths := []string{"../../infra/docker/docker-compose.yaml"}
	identifier := strings.ToLower(uuid.New().String())

	compose := testcontainers.NewLocalDockerCompose(composeFilePaths, identifier)
	execError := compose.WithCommand([]string{"up", "-d"}).Invoke()

	err := execError.Error
	if err != nil {
		return nil, fmt.Errorf("Could not run compose file: %v - %v", composeFilePaths, err)
	}

	return compose, nil
}

type errorResponse struct {
	Message string `json:"message"`
}

type detailResponse struct {
	Id        string `json:"id,omitempty"`
	JobId     string `json:"job_id"`
	ObjectId  string `json:"object_id"`
	Status    string `json:"status"`
	Timestamp int64  `json:"timestamp"`
}

type successResponse struct {
	Message detailResponse `json:"message"`
}

func TestUserCreateJobFlow(t *testing.T) {
	compose, err := setUpContainers()
	if err != nil {
		t.Fatal(err)
	}
	defer compose.Down()

	t.Run("create job", func(t *testing.T) {
		objectId := "random-object-id"
		status := "processing"
		payload := strings.NewReader(fmt.Sprintf(`{"object_id": "%v"}`, objectId))

		succRes := successResponse{Message: detailResponse{}}
		expected := successResponse{Message: detailResponse{ObjectId: objectId, Status: status}}

		res, err := http.Post("http://localhost", "application/json", payload)
		if err != nil {
			t.Fatal(err.Error())
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}
		defer res.Body.Close()

		if err := json.Unmarshal(data, &succRes); err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}

		if res.StatusCode != http.StatusCreated {
			t.Errorf("got: %v, wanted %v", res.StatusCode, http.StatusCreated)
		}

		if succRes.Message.ObjectId != expected.Message.ObjectId {
			t.Errorf("got: %v, wanted %v", succRes.Message.ObjectId, expected.Message.ObjectId)
		}

		if succRes.Message.Status != expected.Message.Status {
			t.Errorf("got: %v, wanted %v", succRes.Message.Status, expected.Message.Status)
		}
	})

	t.Run("create job with the same object id in less than 5 minutes", func(t *testing.T) {
		objectId := "random-object-id"
		payload := strings.NewReader(fmt.Sprintf(`{"object_id": "%v"}`, objectId))

		errRes := errorResponse{}
		expected := errorResponse{Message: "you need to wait 5 minutes before rerunning the same job"}

		res, err := http.Post("http://localhost", "application/json", payload)
		if err != nil {
			t.Fatal(err.Error())
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}
		defer res.Body.Close()

		if err := json.Unmarshal(data, &errRes); err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}

		if res.StatusCode != http.StatusBadRequest {
			t.Errorf("got: %v, wanted %v", res.StatusCode, http.StatusBadRequest)
		}

		if errRes.Message != expected.Message {
			t.Errorf("got: %v, wanted %v", errRes.Message, expected.Message)
		}
	})

	t.Run("create second job", func(t *testing.T) {
		objectId := "random-object-id_number_two"
		status := "processing"
		payload := strings.NewReader(fmt.Sprintf(`{"object_id": "%v"}`, objectId))

		succRes := successResponse{Message: detailResponse{}}
		expected := successResponse{Message: detailResponse{ObjectId: objectId, Status: status}}

		res, err := http.Post("http://localhost", "application/json", payload)
		if err != nil {
			t.Fatal(err.Error())
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}
		defer res.Body.Close()

		if err := json.Unmarshal(data, &succRes); err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}

		if res.StatusCode != http.StatusCreated {
			t.Errorf("got: %v, wanted %v", res.StatusCode, http.StatusCreated)
		}

		if succRes.Message.ObjectId != expected.Message.ObjectId {
			t.Errorf("got: %v, wanted %v", succRes.Message.ObjectId, expected.Message.ObjectId)
		}

		if succRes.Message.Status != expected.Message.Status {
			t.Errorf("got: %v, wanted %v", succRes.Message.Status, expected.Message.Status)
		}
	})

	t.Run("sleep to finish job processing and verify updated status", func(t *testing.T) {
		fmt.Println("[!] sleeping for 45 seconds to finish the job...")
		time.Sleep(time.Second * 45)

		objectId := "random-object-id"
		status := "finished"

		succRes := successResponse{Message: detailResponse{}}
		expected := successResponse{Message: detailResponse{ObjectId: objectId, Status: status}}

		res, err := http.Get("http://localhost/" + objectId)
		if err != nil {
			t.Fatal(err.Error())
		}

		data, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}
		defer res.Body.Close()

		if err := json.Unmarshal(data, &succRes); err != nil {
			t.Fatalf("error not expected, but got: %v", err.Error())
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("got: %v, wanted %v", res.StatusCode, http.StatusOK)
		}

		if succRes.Message.ObjectId != expected.Message.ObjectId {
			t.Errorf("got: %v, wanted %v", succRes.Message.ObjectId, expected.Message.ObjectId)
		}

		if succRes.Message.Status != expected.Message.Status {
			t.Errorf("got: %v, wanted %v", succRes.Message.Status, expected.Message.Status)
		}
	})
}
