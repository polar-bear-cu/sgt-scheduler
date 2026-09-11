package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/polar-bear-cu/sgt-scheduler/dtos"
)

func Health(c *gin.Context) {
	c.JSON(http.StatusOK, dtos.HealthResponse{
		Status:      "ok",
		Timestamp:   time.Now(),
		ServiceName: "scheduler",
	})
}
