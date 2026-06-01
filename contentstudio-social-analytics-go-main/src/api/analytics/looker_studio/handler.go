// Package looker_studio provides the HTTP handler for generating Looker Studio deep links.
package looker_studio

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/middleware"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

const lookerStudioBaseURL = "https://lookerstudio.google.com/datasources/create"
const lookerStudioTemplateBaseURL = "https://lookerstudio.google.com/reporting/create"
const defaultDatasourceAlias = "ds0"

// ApiKeyFinder looks up a user's active API key by user ID.
type ApiKeyFinder interface {
	FindActiveByUserID(ctx context.Context, userID string) (*mongoModels.ApiKey, error)
}

// Handler handles HTTP requests for Looker Studio connector integration.
type Handler struct {
	cfg        config.LookerStudioConfig
	apiKeyRepo ApiKeyFinder
	logger     zerolog.Logger
}

// NewHandler creates a new LookerStudio handler.
func NewHandler(cfg config.LookerStudioConfig, apiKeyRepo ApiKeyFinder, logger zerolog.Logger) *Handler {
	return &Handler{
		cfg:        cfg,
		apiKeyRepo: apiKeyRepo,
		logger:     logger.With().Str("handler", "looker-studio").Logger(),
	}
}

type connectResponse struct {
	Status bool   `json:"status"`
	URL    string `json:"url"`
}

// connectParams holds the parsed request parameters from either GET query string or POST JSON body.
type connectParams struct {
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	Platform      string `json:"platform"`
	AccountID     string `json:"account_id"`
	AccountName   string `json:"account_name"`
	TemplateID    string `json:"template_id"`
}

func parseConnectParams(r *http.Request) connectParams {
	if r.Method == http.MethodPost {
		var p connectParams
		json.NewDecoder(r.Body).Decode(&p)
		return p
	}
	q := r.URL.Query()
	return connectParams{
		WorkspaceID:   q.Get("workspace_id"),
		WorkspaceName: q.Get("workspace_name"),
		Platform:      q.Get("platform"),
		AccountID:     q.Get("account_id"),
		AccountName:   q.Get("account_name"),
		TemplateID:    q.Get("template_id"),
	}
}

// HandleConnect handles GET and POST /analytics/looker-studio/connect.
//
// GET (deep link flow): platform, workspace_id, account_id are all required.
// Returns a fully pre-filled Looker Studio URL so the user lands on a configured
// data source without manual steps.
//
// POST (wizard flow): only workspace_id is required.
// Returns a connector URL that opens the Looker Studio setup wizard pre-filled
// with the workspace and API key so the user selects platform and account manually.
func (h *Handler) HandleConnect(w http.ResponseWriter, r *http.Request) {
	if h.cfg.ConnectorID == "" {
		h.logger.Error().Msg("APP_LOOKER_STUDIO_CONNECTOR_ID not configured")
		httputil.WriteError(w, h.logger, httputil.NewInternalError("Looker Studio connector is not configured"))
		return
	}

	p := parseConnectParams(r)

	if p.WorkspaceID == "" {
		httputil.WriteError(w, h.logger, httputil.NewBadRequestError("workspace_id is required"))
		return
	}
	if r.Method == http.MethodGet && (p.Platform == "" || p.AccountID == "") {
		httputil.WriteError(w, h.logger, httputil.NewBadRequestError("platform, workspace_id and account_id are required"))
		return
	}

	claims := middleware.GetClaims(r.Context())
	if claims == nil || claims.Subject == "" {
		httputil.WriteError(w, h.logger, httputil.NewUnauthorizedError("unauthenticated"))
		return
	}

	apiKey, err := h.apiKeyRepo.FindActiveByUserID(r.Context(), claims.Subject)
	if err != nil {
		h.logger.Error().Err(err).Str("user_id", claims.Subject).Msg("Failed to look up API key")
		httputil.WriteError(w, h.logger, httputil.NewInternalError("Failed to retrieve API key"))
		return
	}
	if apiKey == nil {
		httputil.WriteError(w, h.logger, httputil.NewBadRequestError("No active API key found — generate one from ContentStudio Settings"))
		return
	}

	var deepLink string
	if p.TemplateID != "" {
		dsAlias := defaultDatasourceAlias
		deepLink = lookerStudioTemplateBaseURL +
			"?c.reportId=" + url.QueryEscape(p.TemplateID) +
			"&ds." + dsAlias + ".connector=community" +
			"&ds." + dsAlias + ".connectorId=" + url.QueryEscape(h.cfg.ConnectorID) +
			"&ds." + dsAlias + ".access_token=" + url.QueryEscape(apiKey.Key) +
			"&ds." + dsAlias + ".workspace_id=" + url.QueryEscape(p.WorkspaceID) +
			"&ds." + dsAlias + ".platform=" + url.QueryEscape(p.Platform) +
			"&ds." + dsAlias + ".account_id=" + url.QueryEscape(p.AccountID)
	} else {
		cfg := map[string]string{
			"access_token": apiKey.Key,
			"workspace_id": p.WorkspaceID,
		}
		if p.Platform != "" {
			cfg["platform"] = p.Platform
		}
		if p.AccountID != "" {
			cfg["account_id"] = p.AccountID
		}
		if p.WorkspaceName != "" {
			cfg["workspace_name"] = p.WorkspaceName
		}
		if p.AccountName != "" {
			cfg["account_name"] = p.AccountName
		}
		connectorConfig, err := json.Marshal(cfg)
		if err != nil {
			h.logger.Error().Err(err).Msg("Failed to marshal connector config")
			httputil.WriteError(w, h.logger, httputil.NewInternalError("Failed to build connector config"))
			return
		}
		deepLink = lookerStudioBaseURL +
			"?connectorId=" + url.QueryEscape(h.cfg.ConnectorID) +
			"&connectorConfig=" + url.QueryEscape(string(connectorConfig))
	}

	h.logger.Info().
		Str("user_id", claims.Subject).
		Str("platform", p.Platform).
		Str("workspace_id", p.WorkspaceID).
		Str("template_id", p.TemplateID).
		Msg("Generated Looker Studio deep link")

	httputil.WriteJSON(w, http.StatusOK, connectResponse{Status: true, URL: deepLink})
}
