package controllers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"firmaelectronica/pkg/auth"
	"firmaelectronica/pkg/db"
	"firmaelectronica/pkg/email"
	"firmaelectronica/pkg/response"
	"firmaelectronica/pkg/storage"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentHandler struct {
	DB      *db.DB
	Storage storage.Storage
	Email   email.Provider
	BaseURL string
}

func NewDocumentHandler(database *db.DB, storageService storage.Storage, emailProvider email.Provider, baseURL string) *DocumentHandler {
	return &DocumentHandler{
		DB:      database,
		Storage: storageService,
		Email:   emailProvider,
		BaseURL: baseURL,
	}
}

type CreateDocumentRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	Signers     []Signer   `json:"signers"`
}

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
			response.InternalServerError(w, fmt.Errorf("panic: %v", r))
			return
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
		signerHash := generateSignerHash(document.ID, signerReq.Email, h.BaseURL)

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

	if err := h.sendSignerNotifications(tx, document, signerInfos); err != nil {
		tx.Rollback()
		log.Printf("Error sending notification emails: %v", err)
		response.InternalServerError(w, err)
		return
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
func generateSignerHash(documentID uuid.UUID, email string, baseURL string) string {
	hasher := sha256.New()
	hasher.Write([]byte(documentID.String()))
	hasher.Write([]byte(email))
	hasher.Write([]byte(baseURL))
	hasher.Write([]byte(time.Now().String()))
	return hex.EncodeToString(hasher.Sum(nil)[:16]) // Use first 16 bytes (32 chars)
}

// sendSignerNotifications sends email notifications to all signers
func (h *DocumentHandler) sendSignerNotifications(tx *gorm.DB, document db.Document, signers []SignerInformation) error {
	if h.Email == nil {
		log.Println("Email service not configured, skipping signer notifications")
		return nil
	}

	ctx := context.Background()

	for _, signer := range signers {
		signURL := fmt.Sprintf("%s/sign/%s", h.BaseURL, signer.Hash)

		// Send email
		emailMsg := &email.Email{
			To:      []string{signer.Email},
			Subject: fmt.Sprintf("You have a document to sign: %s", document.Title),
			Body:    fmt.Sprintf("Hello %s,\n\nYou have been requested to sign the document \"%s\". Please visit %s to review and sign the document.\n\nThank you,\nFirma Electronica Team", signer.FirstName, document.Title, signURL),
			HTMLBody: fmt.Sprintf(`<html>
				<body>
					<h1>Document Signature Request</h1>
					<p>Hello %s,</p>
					<p>You have been requested to sign the document <strong>%s</strong>.</p>
					<p>Please click the button below to review and sign the document:</p>
					<div style="text-align: center; margin: 30px 0;">
						<a href="%s" style="background-color: #4CAF50; color: white; padding: 14px 20px; text-align: center; text-decoration: none; display: inline-block; border-radius: 4px; font-weight: bold;">
							Sign Document
						</a>
					</div>
					<p>Or copy and paste this link in your browser:</p>
					<p>%s</p>
					<p>Thank you,<br>Firma Electronica Team</p>
				</body>
			</html>`, signer.FirstName, document.Title, signURL, signURL),
		}

		messageID, err := h.Email.Send(ctx, emailMsg)
		if err != nil {
			log.Printf("Error sending notification email to %s: %v", signer.Email, err)
			return fmt.Errorf("failed to send email to %s: %w", signer.Email, err)
		}

		log.Printf("Notification email sent to %s (Message ID: %s)", signer.Email, messageID)

		// Get signer ID from database using transaction
		var dbSigner db.Signer
		if err := tx.Where("hash = ?", signer.Hash).First(&dbSigner).Error; err != nil {
			log.Printf("Error retrieving signer by hash: %v", err)
			return fmt.Errorf("failed to retrieve signer ID for hash %s: %w", signer.Hash, err)
		}

		// Record notification in the database using the transaction
		notification := db.Notification{
			SignerID:   dbSigner.ID,
			DocumentID: document.ID,
			Type:       db.NotificationTypeInvitation,
			Status:     db.NotificationStatusSent,
		}

		if err := tx.Create(&notification).Error; err != nil {
			log.Printf("Error recording notification in database: %v", err)
			return fmt.Errorf("failed to record notification: %w", err)
		}
	}

	return nil
}

// getSignerIDByHash retrieves the signer ID given the hash
func (h *DocumentHandler) getSignerIDByHash(hash string) uint {
	var signer db.Signer
	if err := h.DB.Where("hash = ?", hash).First(&signer).Error; err != nil {
		log.Printf("Error retrieving signer by hash: %v", err)
		return 0
	}
	return signer.ID
}
