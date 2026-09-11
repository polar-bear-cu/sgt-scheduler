package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/polar-bear-cu/sgt-scheduler/controllers"
)

func Register(r *gin.Engine) {
	r.GET("/health", controllers.Health)
}
