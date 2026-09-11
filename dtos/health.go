package dtos

import "time"

type HealthResponse struct {
	Status      string    `json:"status"`
	Timestamp   time.Time `json:"timestamp"`
	ServiceName string    `json:"serviceName"`
}
