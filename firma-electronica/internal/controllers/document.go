package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"firmaelectronica/pkg/auth"
	"firmaelectronica/pkg/db"
	"firmaelectronica/pkg/response"
	"firmaelectronica/pkg/storage"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// DocumentHandler handles document-related requests
type DocumentHandler struct {
	DB      *db.DB
	Storage *storage.S3Storage
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(database *db.DB, s3storage *storage.S3Storage) *DocumentHandler {
	return &DocumentHandler{
		DB:      database,
		Storage: s3storage,
	}
}

// CreateDocumentRequest represents the request to create a document
type CreateDocumentRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Signers     []Signer   `json:"signers"`
}

// Signer represents a person who will sign the document
type Signer struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// CreateDocumentResponse represents the response after creating a document
type CreateDocumentResponse struct {
	DocumentID uuid.UUID           `json:"documentId"`
	Title      string              `json:"title"`
	Status     db.DocumentStatus   `json:"status"`
	Signers    []SignerInformation `json:"signers"`
}

// SignerInformation contains signer information in the response
type SignerInformation struct {
	Email     string          `json:"email"`
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Status    db.SignerStatus `json:"status"`
	Hash      string          `json:"hash"`
}

// Create handles document creation requests
func (h *DocumentHandler) Create(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method != http.MethodPost {
		response.MethodNotAllowed(w)
		return
	}

	// Get user from context
	claims, ok := r.Context().Value(UserClaimsKey).(*auth.JWTClaims)
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		log.Printf("Error parsing multipart form: %v", err)
		response.BadRequest(w, "Invalid multipart form", err.Error())
		return
	}

	// Get document metadata
	metadataStr := r.FormValue("metadata")
	if metadataStr == "" {
		response.BadRequest(w, "Missing document metadata", "")
		return
	}

	var createReq CreateDocumentRequest
	if err := json.Unmarshal([]byte(metadataStr), &createReq); err != nil {
		log.Printf("Error unmarshaling document metadata: %v", err)
		response.BadRequest(w, "Invalid document metadata", err.Error())
		return
	}

	// Validate required fields
	if createReq.Title == "" {
		response.BadRequest(w, "Title is required", "")
		return
	}

	if len(createReq.Signers) == 0 {
		response.BadRequest(w, "At least one signer is required", "")
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("document")
	if err != nil {
		log.Printf("Error getting document file: %v", err)
		response.BadRequest(w, "Missing document file", err.Error())
		return
	}
	defer file.Close()

	// Calculate file hash for integrity verification
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		log.Printf("Error calculating file hash: %v", err)
		response.InternalServerError(w, err)
		return
	}
	fileHash := hex.EncodeToString(hasher.Sum(nil))

	// Rewind the file reader for upload
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		log.Printf("Error rewinding file reader: %v", err)
		response.InternalServerError(w, err)
		return
	}

	// Upload file to S3
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	fileInfo, err := h.Storage.UploadFile(r.Context(), file, header.Filename, contentType)
	if err != nil {
		log.Printf("Error uploading file to S3: %v", err)
		response.InternalServerError(w, err)
		return
	}

	// Start a database transaction
	tx := h.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		response.InternalServerError(w, tx.Error)
		return
	}

	// Rollback transaction in case of error
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	// Create document in database
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		tx.Rollback()
		log.Printf("Error parsing user ID: %v", err)
		response.InternalServerError(w, err)
		return
	}

	document := db.Document{
		Title:       createReq.Title,
		Description: createReq.Description,
		ContentURL:  fileInfo.URL,
		ContentHash: fileHash,
		Status:      db.DocumentStatusDraft,
		UserID:      userID,
		ExpiresAt:   createReq.ExpiresAt,
	}

	if err := tx.Create(&document).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating document: %v", err)
		response.InternalServerError(w, err)
		return
	}

	// Create signers
	signerInfos := make([]SignerInformation, 0, len(createReq.Signers))
	for _, signerReq := range createReq.Signers {
		// Generate a unique hash for signer access
		signerHash := generateSignerHash(document.ID, signerReq.Email)

		signer := db.Signer{
			DocumentID: document.ID,
			Email:      signerReq.Email,
			FirstName:  signerReq.FirstName,
			LastName:   signerReq.LastName,
			Hash:       signerHash,
			Status:     db.SignerStatusPending,
		}

		if err := tx.Create(&signer).Error; err != nil {
			tx.Rollback()
			log.Printf("Error creating signer: %v", err)
			response.InternalServerError(w, err)
			return
		}

		signerInfos = append(signerInfos, SignerInformation{
			Email:     signer.Email,
			FirstName: signer.FirstName,
			LastName:  signer.LastName,
			Status:    signer.Status,
			Hash:      signer.Hash,
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		response.InternalServerError(w, err)
		return
	}

	// Return response
	resp := CreateDocumentResponse{
		DocumentID: document.ID,
		Title:      document.Title,
		Status:     document.Status,
		Signers:    signerInfos,
	}

	response.Created(w, resp)
}

// generateSignerHash generates a unique hash for signer access
func generateSignerHash(documentID uuid.UUID, email string) string {
	hasher := sha256.New()
	hasher.Write([]byte(documentID.String()))
	hasher.Write([]byte(email))
	hasher.Write([]byte(time.Now().String()))
	return hex.EncodeToString(hasher.Sum(nil)[:16]) // Use first 16 bytes (32 chars)
}
