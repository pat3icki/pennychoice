package backend

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pat3icki/pennychoice/internal/services/campaign"
)

type campaignUserInfoID struct {
	Organisation string `json:"organisation"`
	Campaign     string `json:"campaign"`
}

type campaignUserInfo struct {
	Authorization string             `json:"authorization"`
	ID            campaignUserInfoID `json:"id"`
}

// Struct for json_data/campaign/Create.json and Activiate.json
type campaignRequestEnvelope struct {
	Header      requestHeader `json:"header"`
	RequestData struct {
		UserInfo campaignUserInfo `json:"user_info"`
		Data     userPayloadData  `json:"data"`
	} `json:"request_data"`
}

// Struct for json_data/campaign/campaign.json
type updateVotePeriodRequest struct {
	Command     string `json:"command"`
	RequestData struct {
		ClientInfo struct {
			Token        string `json:"token"`
			Organisation string `json:"organisation"`
			Campaign     string `json:"campaign"`
		} `json:"client_info"`
		Data struct {
			VoteStartsAt string `json:"vote_starts_at"`
			VoteEndsAt   string `json:"vote_ends_at"`
		} `json:"data"`
	} `json:"request_data"`
}

func (b *Backend) handleCreateCampaign(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req campaignRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	userID := ClientIDFromContext(r.Context())
	orgID, _ := uuid.Parse(req.RequestData.UserInfo.ID.Organisation)

	auth := campaign.Auth{
		User:         userID,
		Organisation: orgID,
	}

	payloadBytes, err := json.Marshal(req.RequestData.Data.Payload)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	params := campaign.CreateCampaignParams{
		Schema:  req.RequestData.Data.Schema,
		Payload: payloadBytes,
	}

	if err := b.services.Campaign.Create(r.Context(), auth, params); err != nil {
		writeError(w, userID, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, userID, map[string]string{"status": "created"}, http.StatusCreated)
}

func (b *Backend) handleActivateCampaign(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req campaignRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	userID := ClientIDFromContext(r.Context())
	orgID, _ := uuid.Parse(req.RequestData.UserInfo.ID.Organisation)
	campaignID, err := uuid.Parse(req.RequestData.UserInfo.ID.Campaign)
	if err != nil {
		// try route URL param fallback
		paramID := chi.URLParam(r, "id")
		campaignID, err = uuid.Parse(paramID)
		if err != nil {
			writeError(w, userID, err, http.StatusBadRequest)
			return
		}
	}

	auth := campaign.Auth{
		User:         userID,
		Organisation: orgID,
	}

	if err := b.services.Campaign.Activiate(r.Context(), auth, campaignID); err != nil {
		writeError(w, userID, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, userID, map[string]string{"status": "activated"}, http.StatusOK)
}

func (b *Backend) handlePauseCampaign(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req campaignRequestEnvelope
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	userID := ClientIDFromContext(r.Context())
	orgID, _ := uuid.Parse(req.RequestData.UserInfo.ID.Organisation)
	campaignID, err := uuid.Parse(req.RequestData.UserInfo.ID.Campaign)
	if err != nil {
		paramID := chi.URLParam(r, "id")
		campaignID, err = uuid.Parse(paramID)
		if err != nil {
			writeError(w, userID, err, http.StatusBadRequest)
			return
		}
	}

	auth := campaign.Auth{
		User:         userID,
		Organisation: orgID,
	}

	if err := b.services.Campaign.Pause(r.Context(), auth, campaignID); err != nil {
		writeError(w, userID, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, userID, map[string]string{"status": "paused"}, http.StatusOK)
}

func (b *Backend) handleUpdateVotePeriodCampaign(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req updateVotePeriodRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeError(w, uuid.Nil, err, http.StatusBadRequest)
		return
	}

	userID := ClientIDFromContext(r.Context())
	orgID, _ := uuid.Parse(req.RequestData.ClientInfo.Organisation)
	campaignID, err := uuid.Parse(req.RequestData.ClientInfo.Campaign)
	if err != nil {
		paramID := chi.URLParam(r, "id")
		campaignID, err = uuid.Parse(paramID)
		if err != nil {
			writeError(w, userID, err, http.StatusBadRequest)
			return
		}
	}

	auth := campaign.Auth{
		User:         userID,
		Organisation: orgID,
	}

	startsAt, _ := time.Parse(time.RFC3339, req.RequestData.Data.VoteStartsAt)
	endsAt, _ := time.Parse(time.RFC3339, req.RequestData.Data.VoteEndsAt)

	votePeriod := campaign.VotePeriod{
		Start: startsAt,
		End:   endsAt,
	}

	if err := b.services.Campaign.UpdateVotePeriods(r.Context(), auth, campaignID, votePeriod); err != nil {
		writeError(w, userID, err, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, userID, map[string]string{"status": "updated"}, http.StatusOK)
}
