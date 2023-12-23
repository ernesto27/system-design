package types

type Request struct {
	UserID int
	Text   string `json:"text"`
}

type JSONResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
