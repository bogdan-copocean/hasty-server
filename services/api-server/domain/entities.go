package domain

type Job struct {
	Id        string `json:"id,omitempty" bson:"_id"`
	JobId     string `json:"job_id"`
	ObjectId  string `json:"object_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at" bson:"createdAt"`
}