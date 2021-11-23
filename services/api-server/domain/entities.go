package domain

type Job struct {
	Id            string `json:"id,omitempty" bson:"_id"`
	JobId         string `json:"job_id"`
	ObjectId      string `json:"object_id"`
	Status        string `json:"status"`
	Timestamp     int64  `json:"timestamp" bson:"timestamp"`
	SleepTimeUsed int    `json:"sleep_time_used"`
}

type ResponseJob struct {
	JobId string `json:"job_id"`
}
