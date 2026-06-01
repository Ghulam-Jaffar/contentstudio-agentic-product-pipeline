package api

// ListeningWorkRequest is the HTTP request body for triggering a listening work order.
type ListeningWorkRequest struct {
	TopicID      string `json:"topic_id"`
	WorkspaceID  string `json:"workspace_id"`
	SuperAdminID string `json:"super_admin_id"`
	SyncType     string `json:"sync_type"`
}
