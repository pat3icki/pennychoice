package backend

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	apperrors "github.com/pat3icki/pennychoice/backend/errors"
)

// Response is the standard envelope for all backend API responses.
type Response struct {
	Client    ClientInfo      `json:"client"`
	Data      json.RawMessage `json:"data,omitempty"`
	Error     *ErrorInfo      `json:"error,omitempty"`
	Timestamp string          `json:"timestamp,omitempty"`
}

// ClientInfo carries authenticated user context for the frontend.
type ClientInfo struct {
	ID           uuid.UUID `json:"id"`
	Organisation uuid.UUID `json:"organisation"`
}

// ErrorInfo carries structured error information.
type ErrorInfo struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

// writeSuccess writes a successful response with the given data and status code.
func writeSuccess(w http.ResponseWriter, clientID uuid.UUID, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var raw json.RawMessage
	if data != nil {
		b, err := json.Marshal(data)
		if err != nil {
			// Fallback to internal error if marshaling fails (should not happen)
			writeError(w, clientID, apperrors.Internal("backend", "writeSuccess", err), http.StatusInternalServerError)
			return
		}
		raw = b
	}

	resp := Response{
		Client: ClientInfo{ID: clientID},
		Data:   raw,
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		// Can't do much here, log maybe
	}
}

// writeError writes an error response with the given error and status code.
// It maps the error type to the appropriate ErrorInfo.Type.
func writeError(w http.ResponseWriter, clientID uuid.UUID, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errType := apperrors.GetType(err)
	msg := err.Error()
	field := ""

	// If the error is a typed app error, extract additional info
	if appErr, ok := err.(apperrors.AppError); ok {
		field = appErr.Field()
	}

	// Map to envelope type
	var envelopeType string
	switch errType {
	case apperrors.TypeValidation:
		envelopeType = "validation"
	case apperrors.TypeInput:
		envelopeType = "input"
	case apperrors.TypeNotFound:
		envelopeType = "not_found"
	case apperrors.TypePermission:
		envelopeType = "permission"
	case apperrors.TypeConflict:
		envelopeType = "conflict"
	case apperrors.TypeDatabase, apperrors.TypeInternal:
		envelopeType = "internal"
	default:
		envelopeType = "internal"
	}

	resp := Response{
		Client: ClientInfo{ID: clientID},
		Error: &ErrorInfo{
			Type:    envelopeType,
			Message: msg,
			Field:   field,
		},
	}
	_ = json.NewEncoder(w).Encode(resp)
}
