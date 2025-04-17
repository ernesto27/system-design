// filepath: /home/ernesto/code/system-design/projects-managment/backend/controllers/project.go
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/models"
	"server/response"
)

type Project struct {
	ProjectService models.ProjectService
}

const MessageFieldsEmpty = "project name or description is empty"

func (p *Project) Create(w http.ResponseWriter, r *http.Request) {
	var projectReq models.Project

	err := json.NewDecoder(r.Body).Decode(&projectReq)
	if err != nil {
		fmt.Println("Error parsing request:", err)
		response.NewWithoutData().BadRequest(w)
		return
	}

	// Validate request
	if projectReq.Name == "" || projectReq.Description == "" {
		fmt.Println(MessageFieldsEmpty)
		response.NewWithoutData().WithMessage(MessageFieldsEmpty).BadRequest(w)
		return
	}

	// Call service to create project
	project, err := p.ProjectService.CreateProject(projectReq)

	if err != nil {
		fmt.Println("Error creating project:", err)
		response.NewWithoutData().InternalServerError(w)
		return
	}

	// Return successful response
	response.New(project).Success(w)
}
