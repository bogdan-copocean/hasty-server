package events

import "github.com/bogdan-copocean/hasty-server/services/api-server/domain"

type JobEvent struct {
	Subject string     `json:"subject"`
	Job     domain.Job `json:"job"`
}
