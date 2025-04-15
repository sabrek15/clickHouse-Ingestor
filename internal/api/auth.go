package api

import (
	"net/http"
	"time"
	"encoding/json"

	"github.com/sabrek15/clickhouse-ingestor/internal/storage"
	"github.com/sabrek15/clickhouse-ingestor/internal/filehandler"
	"github.com/sabrek15/clickhouse-ingestor/internal/auth"
)

type APIHandler struct {
	jwtValidator    *auth.JWTValidator
	clickhouseService *storage.ClickHouseService
	fileService     *filehandler.FileService
}

func NewAPIHandler(
	jwtValidator *auth.JWTValidator,
	clickhouseService *storage.ClickHouseService,
	fileService *filehandler.FileService,
) *APIHandler {
	return &APIHandler{
		jwtValidator:    jwtValidator,
		clickhouseService: clickhouseService,
		fileService:     fileService,
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func (h *APIHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	type loginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// In a real app, validate against your user database
	// This is a simplified example
	storedHash, err := auth.HashPassword("correct-password") // Replace with DB lookup
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Server error")
		return
	}

	if err := auth.CheckPassword(storedHash, req.Password); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Generate JWT token
	token, err := h.jwtValidator.GenerateToken(req.Username, 24*time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

func (h *APIHandler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		Database string `json:"database"`
		User     string `json:"user"`
		JWTToken string `json:"jwtToken"`
		Secure   bool   `json:"secure"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.jwtValidator.Validate(req.JWTToken); err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid JWT token: "+err.Error())
		return
	}

	err := h.clickhouseService.Connect(req.Host, req.Port, req.Database, req.User, req.JWTToken, req.Secure)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to connect: "+err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Connected successfully",
	})
}

// Schema discovery handler
func (h *APIHandler) HandleSchemaDiscovery(w http.ResponseWriter, r *http.Request) {
	sourceType := r.URL.Query().Get("source")

	switch sourceType {
	case "clickhouse":
		tables, err := h.clickhouseService.GetTables()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get tables: "+err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"tables": tables})

	case "file":
		filePath := r.URL.Query().Get("filePath")
		delimiter := r.URL.Query().Get("delimiter")
		
		if filePath == "" {
			respondWithError(w, http.StatusBadRequest, "filePath parameter is required")
			return
		}

		delimiterRune := ','
		if delimiter != "" {
			delimiterRune = []rune(delimiter)[0]
		}

		columns, err := h.fileService.ReadSchema(filePath, delimiterRune)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read file: "+err.Error())
			return
		}
		respondWithJSON(w, http.StatusOK, map[string]interface{}{"columns": columns})

	default:
		respondWithError(w, http.StatusBadRequest, "Invalid source type")
	}
}

// Data transfer handler
func (h *APIHandler) HandleDataTransfer(w http.ResponseWriter, r *http.Request) {
	 var req struct {
        SourceType   string            `json:"sourceType"`
        SourceParams map[string]string `json:"sourceParams"`
        TargetType   string            `json:"targetType"`
        TargetParams map[string]string `json:"targetParams"`
        Columns      []string          `json:"columns"`
        Table        string            `json:"table"` // For ClickHouse source
    }

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Implementation would go here
	// This is just a placeholder response
	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"recordCount": 0,
	})
}

// UI handlers
func (h *APIHandler) HandleHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/index.html")
}

func (h *APIHandler) HandleResults(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/results.html")
}