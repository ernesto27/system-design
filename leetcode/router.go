package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.Default()
	
	handlers := NewHandlers(db)
	
	r.GET("/problems", handlers.GetProblems)
	r.GET("/problems/:problem_id", handlers.GetProblem)
	r.POST("/problems/:problem_id/submission", handlers.SubmitSolution)
	
	return r
}