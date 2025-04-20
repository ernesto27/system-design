package controllers

import (
	"encoding/json"
	"fmt"

	"net/http"
	"server/internal"
	"server/models"
	"server/response"
	"strconv"

	"github.com/go-chi/chi"
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

	if projectReq.Name == "" || projectReq.Description == "" {
		fmt.Println(MessageFieldsEmpty)
		response.NewWithoutData().WithMessage(MessageFieldsEmpty).BadRequest(w)
		return
	}

	userID, ok := internal.GetUserIDFromContext(r.Context())
	if !ok {
		fmt.Println("User ID not found in context")
		response.NewWithoutData().WithMessage("Unauthorized").Unauthorized(w)
		return
	}

	projectReq.CreatedBy = userID

	project, err := p.ProjectService.Create(projectReq)

	if err != nil {
		fmt.Println("Error creating project:", err)
		response.NewWithoutData().InternalServerError(w)
		return
	}

	response.New(project).Success(w)
}

// GetAll handles retrieving projects with pagination
func (p *Project) GetAll(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}

	projects, err := p.ProjectService.GetAllProjects(page, limit)
	if err != nil {
		fmt.Println("Error fetching projects:", err)
		response.NewWithoutData().InternalServerError(w)
		return
	}

	response.New(projects).Success(w)
}

func (p *Project) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		fmt.Println("Missing project ID parameter")
		response.NewWithoutData().WithMessage("Missing project ID").BadRequest(w)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Invalid project ID:", err)
		response.NewWithoutData().WithMessage("Invalid project ID").BadRequest(w)
		return
	}
	project, err := p.ProjectService.GetByID(id)
	if err != nil {
		if err.Error() == "project not found" {
			fmt.Println("Project not found:", id)
			response.NewWithoutData().WithMessage("Project not found").InternalServerError(w)
			return
		}

		fmt.Println("Error fetching project:", err)
		response.NewWithoutData().InternalServerError(w)
		return
	}

	response.New(project).Success(w)
}
