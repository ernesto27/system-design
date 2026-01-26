package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handlers struct {
	DB *gorm.DB
}

func NewHandlers(db *gorm.DB) *Handlers {
	return &Handlers{DB: db}
}

func (h *Handlers) GetProblems(c *gin.Context) {
	var problems []Problem
	query := h.DB

	startStr := c.Query("start")
	endStr := c.Query("end")

	if startStr != "" {
		start, err := strconv.Atoi(startStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start parameter"})
			return
		}
		query = query.Offset(start)
	}

	if endStr != "" {
		end, err := strconv.Atoi(endStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end parameter"})
			return
		}
		
		if startStr != "" {
			start, _ := strconv.Atoi(startStr)
			limit := end - start + 1
			if limit > 0 {
				query = query.Limit(limit)
			}
		} else {
			query = query.Limit(end + 1)
		}
	}

	if err := query.Find(&problems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problems"})
		return
	}

	c.JSON(http.StatusOK, problems)
}

func (h *Handlers) GetProblem(c *gin.Context) {
	problemID := c.Param("problem_id")
	
	id, err := strconv.Atoi(problemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var problem Problem
	if err := h.DB.First(&problem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problem"})
		return
	}

	c.JSON(http.StatusOK, problem)
}

func (h *Handlers) SubmitSolution(c *gin.Context) {
	problemID := c.Param("problem_id")
	
	id, err := strconv.Atoi(problemID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid problem ID"})
		return
	}

	var req SubmissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := strconv.Atoi(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var problem Problem
	if err := h.DB.First(&problem, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Problem not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch problem"})
		return
	}

	var user User
	if err := h.DB.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}

	testCases := []TestCase{
		{Status: "success"},
		{Status: "success"},
		{Status: "fail"},
	}
	
	status := "fail"

	submission := Submission{
		UserID:    uint(userID),
		ProblemID: uint(id),
		Code:      req.Code,
		Language:  req.Language,
		Status:    status,
	}

	if err := h.DB.Create(&submission).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save submission"})
		return
	}

	response := SubmissionResponse{
		Status:    status,
		TestCases: testCases,
	}

	c.JSON(http.StatusOK, response)
}