package listening

import "net/http"

// RegisterRoutes registers all listening API endpoints on the provided mux.
func RegisterRoutes(mux *http.ServeMux, mentions *MentionsHandler, views *ViewsHandler, analytics *AnalyticsHandler) {
	// Mentions endpoints
	mux.HandleFunc("GET /api/listening/mentions", mentions.HandleListMentions)
	mux.HandleFunc("PATCH /api/listening/mentions/{id}", mentions.HandlePatchMention)
	mux.HandleFunc("POST /api/listening/mentions/mark-all-read", mentions.HandleMarkAllRead)
	mux.HandleFunc("GET /api/listening/mentions/unread-count", mentions.HandleUnreadCount)

	// Analytics endpoints
	mux.HandleFunc("GET /api/listening/analytics", analytics.HandleGetAnalytics)
	mux.HandleFunc("GET /api/listening/analytics/export", analytics.HandleExportMentions)

	// Views endpoints
	mux.HandleFunc("GET /api/listening/views", views.HandleListViews)
	mux.HandleFunc("POST /api/listening/views", views.HandleCreateView)
	mux.HandleFunc("PUT /api/listening/views/{id}", views.HandleUpdateView)
	mux.HandleFunc("DELETE /api/listening/views/{id}", views.HandleDeleteView)
}
