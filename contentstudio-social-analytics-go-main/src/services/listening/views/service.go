// Package views persists user-defined listening dashboard layouts in Mongo.
// Separate from mentions because these are user-settings, not analytics data,
// and do not belong in the ClickHouse write path.
package views

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

type mentionCounter interface {
	CountMentions(ctx context.Context, filter *clickhouse.MentionFilter) (int, error)
}

type topicWorkspaceResolver interface {
	ListTopicsByWorkspace(ctx context.Context, workspaceID string) ([]*mongoModels.ListeningTopic, error)
}

type Service struct {
	repo           *mongodb.ListeningViewsRepository
	mentionCounter mentionCounter
	topics         topicWorkspaceResolver
	logger         zerolog.Logger
}

func NewService(
	repo *mongodb.ListeningViewsRepository,
	mentionCounter mentionCounter,
	topics topicWorkspaceResolver,
	logger zerolog.Logger,
) *Service {
	return &Service{
		repo:           repo,
		mentionCounter: mentionCounter,
		topics:         topics,
		logger:         logger.With().Str("service", "listening-views").Logger(),
	}
}

func (s *Service) ListViews(ctx context.Context, workspaceID string) ([]apiModels.ViewResponse, error) {
	if err := s.repo.SeedSystemViews(ctx, workspaceID); err != nil {
		s.logger.Warn().Err(err).Str("workspace_id", workspaceID).Msg("Failed to seed system views")
	}

	views, err := s.repo.ListViews(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("views.Service.ListViews: %w", err)
	}

	// Topic lookup is best-effort — it only feeds the count query for views
	// that don't pin explicit topic IDs. A transient MongoDB error shouldn't
	// 500 the whole endpoint; views should still render with count=0 so the
	// sidebar remains usable.
	workspaceTopicIDs, topicsErr := s.listWorkspaceTopicIDs(ctx, workspaceID)
	if topicsErr != nil {
		s.logger.Warn().
			Err(topicsErr).
			Str("workspace_id", workspaceID).
			Msg("Failed to list workspace topics for view counts; counts may be 0 for views without explicit topic IDs")
		workspaceTopicIDs = nil
	}

	result := make([]apiModels.ViewResponse, 0, len(views))
	for _, v := range views {
		count, countErr := s.countForView(ctx, v, workspaceTopicIDs)
		if countErr != nil {
			s.logger.Warn().
				Err(countErr).
				Str("workspace_id", workspaceID).
				Str("view_id", v.ID.Hex()).
				Msg("Failed to populate listening view count")
		}
		result = append(result, toViewResponse(v, count))
	}
	return result, nil
}

func (s *Service) CreateView(ctx context.Context, req *apiModels.ViewRequest) (*apiModels.ViewResponse, error) {
	if req.WorkspaceID == "" {
		return nil, httputil.NewValidationError("workspace_id is required")
	}
	if req.Name == "" {
		return nil, httputil.NewValidationError("name is required")
	}

	view := &mongoModels.ListeningView{
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Icon:        req.Icon,
		Type:        "user",
		FilterPreset: mongoModels.ListeningViewFilterPreset{
			TopicIDs:           req.FilterPreset.TopicIDs,
			Platforms:          req.FilterPreset.Platforms,
			Sentiments:         req.FilterPreset.Sentiments,
			AITags:             req.FilterPreset.AITags,
			MinFollowers:       req.FilterPreset.MinFollowers,
			MinTotalEngagement: req.FilterPreset.MinTotalEngagement,
			Language:           req.FilterPreset.Language,
		},
	}

	created, err := s.repo.CreateView(ctx, view)
	if err != nil {
		return nil, fmt.Errorf("views.Service.CreateView: %w", err)
	}

	resp := toViewResponse(*created, 0)
	return &resp, nil
}

func (s *Service) UpdateView(ctx context.Context, viewID string, req *apiModels.ViewRequest) (*apiModels.ViewResponse, error) {
	existing, err := s.repo.GetViewByID(ctx, viewID)
	if err != nil {
		return nil, fmt.Errorf("views.Service.UpdateView: get: %w", err)
	}
	if existing == nil {
		return nil, httputil.NewValidationError("view not found")
	}
	// All three types (system, preset, user) are user-editable. The only thing
	// system views are protected from is deletion — see DeleteView below.

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Icon != "" {
		update["icon"] = req.Icon
	}
	update["filter_preset"] = mongoModels.ListeningViewFilterPreset{
		TopicIDs:           req.FilterPreset.TopicIDs,
		Platforms:          req.FilterPreset.Platforms,
		Sentiments:         req.FilterPreset.Sentiments,
		AITags:             req.FilterPreset.AITags,
		MinFollowers:       req.FilterPreset.MinFollowers,
		MinTotalEngagement: req.FilterPreset.MinTotalEngagement,
		Language:           req.FilterPreset.Language,
	}

	updated, err := s.repo.UpdateView(ctx, viewID, update)
	if err != nil {
		return nil, fmt.Errorf("views.Service.UpdateView: %w", err)
	}
	if updated == nil {
		return nil, httputil.NewValidationError("view not found after update")
	}

	resp := toViewResponse(*updated, 0)
	return &resp, nil
}

func (s *Service) DeleteView(ctx context.Context, viewID string) error {
	existing, err := s.repo.GetViewByID(ctx, viewID)
	if err != nil {
		return fmt.Errorf("views.Service.DeleteView: get: %w", err)
	}
	if existing == nil {
		return httputil.NewValidationError("view not found")
	}
	if existing.Type == "system" {
		return httputil.NewForbiddenError("system views cannot be deleted")
	}

	if err := s.repo.DeleteView(ctx, viewID); err != nil {
		return fmt.Errorf("views.Service.DeleteView: %w", err)
	}
	return nil
}

func (s *Service) GetViewByID(ctx context.Context, viewID string) (*mongoModels.ListeningView, error) {
	return s.repo.GetViewByID(ctx, viewID)
}

func (s *Service) listWorkspaceTopicIDs(ctx context.Context, workspaceID string) ([]string, error) {
	if s.topics == nil {
		return nil, nil
	}

	topics, err := s.topics.ListTopicsByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	topicIDs := make([]string, 0, len(topics))
	for _, topic := range topics {
		if topic == nil {
			continue
		}

		topicID := topic.TopicID
		if topicID == "" {
			topicID = topic.ID.Hex()
		}
		if topicID != "" {
			topicIDs = append(topicIDs, topicID)
		}
	}

	return topicIDs, nil
}

func (s *Service) countForView(
	ctx context.Context,
	view mongoModels.ListeningView,
	workspaceTopicIDs []string,
) (int, error) {
	if s.mentionCounter == nil {
		return 0, nil
	}

	// Always scope the count query to the current workspace. A saved view may
	// carry topic IDs from when it was created (or from a migration), but we
	// must not count mentions for topics outside this workspace — ClickHouse
	// is shared across tenants and a stale/tampered topic_id would leak counts
	// across workspaces.
	topicIDs := intersectTopicIDs(view.FilterPreset.TopicIDs, workspaceTopicIDs)

	if len(topicIDs) == 0 {
		return 0, nil
	}

	return s.mentionCounter.CountMentions(ctx, &clickhouse.MentionFilter{
		TopicIDs:           topicIDs,
		Platforms:          view.FilterPreset.Platforms,
		Sentiments:         view.FilterPreset.Sentiments,
		AITags:             view.FilterPreset.AITags,
		ExcludeAITags:      view.FilterPreset.ExcludeAITags,
		Language:           view.FilterPreset.Language,
		MinFollowers:       view.FilterPreset.MinFollowers,
		MinTotalEngagement: view.FilterPreset.MinTotalEngagement,
	})
}

// intersectTopicIDs returns the workspace-scoped topic set to query.
//
// When the view pins no topics, every topic in the workspace is counted.
// When it pins topics, we only keep the ones that exist in the workspace so a
// saved view can never count mentions from topics that were never in this
// workspace (or were deleted). If we can't verify the workspace (topic lookup
// errored upstream), we fall back to whatever the view requested — count may
// be zero because those IDs won't match any mentions, but we don't leak.
func intersectTopicIDs(viewTopicIDs, workspaceTopicIDs []string) []string {
	if len(viewTopicIDs) == 0 {
		return workspaceTopicIDs
	}
	if len(workspaceTopicIDs) == 0 {
		return viewTopicIDs
	}

	allowed := make(map[string]struct{}, len(workspaceTopicIDs))
	for _, id := range workspaceTopicIDs {
		allowed[id] = struct{}{}
	}

	filtered := make([]string, 0, len(viewTopicIDs))
	for _, id := range viewTopicIDs {
		if _, ok := allowed[id]; ok {
			filtered = append(filtered, id)
		}
	}
	return filtered
}

func toViewResponse(v mongoModels.ListeningView, count int) apiModels.ViewResponse {
	return apiModels.ViewResponse{
		ID:          v.ID.Hex(),
		WorkspaceID: v.WorkspaceID,
		Name:        v.Name,
		Icon:        v.Icon,
		Type:        v.Type,
		SystemKey:   v.SystemKey,
		Count:       count,
		FilterPreset: apiModels.ViewFilterPreset{
			TopicIDs:           v.FilterPreset.TopicIDs,
			Platforms:          v.FilterPreset.Platforms,
			Sentiments:         v.FilterPreset.Sentiments,
			AITags:             v.FilterPreset.AITags,
			MinFollowers:       v.FilterPreset.MinFollowers,
			MinTotalEngagement: v.FilterPreset.MinTotalEngagement,
			Language:           v.FilterPreset.Language,
		},
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}
