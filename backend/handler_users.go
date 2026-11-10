package backend

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/internal/services/users"
)

// Request headers / client info structure from json_data
type requestHeader struct {
	Authorization string `json:"authorization"`
	Server        string `json:"server"`
	Timestamp     string `json:"timestamp"`
}

type requestUserInfo struct {
	Authorization string `json:"authorization"`
}

type userPayloadData struct {
	Schema  string          `json:"schema"`
	Payload json.RawMessage `json:"payload"`
}

// Struct for json_data/users/CreateProfile.json
type createProfileRequestEnvelope struct {
	Header      requestHeader `json:"header"`
	RequestData struct {
		Data userPayloadData `json:"data"`
	} `json:"request_data"`
}

// Struct for json_data/users/CreateContestantProfile.json
type createContestantProfileRequestEnvelope struct {
	Header      requestHeader `json:"header"`
	RequestData struct {
		UserInfo requestUserInfo `json:"user_info"`
		Data     userPayloadData `json:"data"`
	} `json:"request_data"`
}

// Struct for json_data/users/Login.json
type loginRequestEnvelope struct {
	Header requestHeader `json:"header"`
	Data   struct {
		Mode     string `json:"mode"`
		User     string `json:"user"`
		Password string `json:"password"`
		Options  struct {
			ReturnKeys []string `json:"return_keys"`
		} `json:"options"`
	} `json:"data"`
}

// handleCreateUser processes user profile creation.
func (b *Backend) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req createProfileRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	payloadBytes, err := json.Marshal(req.RequestData.Data.Payload)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	params := &users.CreateUserProfileParams{
		Schema:  req.RequestData.Data.Schema,
		Payload: payloadBytes,
	}

	if err := b.services.User.CreateProfile(r.Context(), params); err != nil {
		writeError(w, uuid.Nil, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, uuid.Nil, map[string]string{"status": "created"}, http.StatusCreated)
}

// handleCreateContestantProfile processes contestant profile creation.
func (b *Backend) handleCreateContestantProfile(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req createContestantProfileRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	// For authenticated user ID: using user ID from context or header
	userID := ClientIDFromContext(r.Context())

	payloadBytes, err := json.Marshal(req.RequestData.Data.Payload)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	params := &users.CreateContestantProfileParams{
		Schema:  req.RequestData.Data.Schema,
		Payload: payloadBytes,
	}

	if err := b.services.User.CreateContestantProfile(r.Context(), userID, params); err != nil {
		writeError(w, userID, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, userID, map[string]string{"status": "created"}, http.StatusCreated)
}

// handleLoginUser processes user login.
func (b *Backend) handleLoginUser(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req loginRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	loginReq := users.LoginRequest{
		Mode:       req.Data.Mode,
		Identifier: req.Data.User,
		Password:   []byte(req.Data.Password),
	}
	if len(req.Data.Options.ReturnKeys) > 0 {
		loginReq.ResultOpts = &users.LoginReqResultOpts{
			Schema: req.Data.Options.ReturnKeys,
		}
	}

	key := []byte(b.config.Monnify().SecretKey())
	if len(key) == 0 {
		key = []byte("default-secret-key-32-bytes-long!")
	}
	res, err := b.services.User.Login(r.Context(), key, loginReq)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusUnauthorized)
		return
	}

	writeSuccess(w, uuid.Nil, res, http.StatusOK)
}
