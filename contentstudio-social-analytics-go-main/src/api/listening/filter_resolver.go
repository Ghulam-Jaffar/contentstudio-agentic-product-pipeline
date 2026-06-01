// Package listening's MentionFilterResolver exists because API callers
// are allowed to send partial MentionFilters, but every downstream query
// (mentions, aggregates, exports) needs a fully-populated filter. Resolving
// this once at the edge keeps the query layer simple and consistent.
//
// Why this file does what it does:
//
//   - ViewID expansion: a "view" is a saved filter preset owned by a workspace.
//     When a caller passes only a ViewID, we hydrate the preset (topics,
//     platforms, sentiments, ai_tags, language, min_followers,
//     min_total_engagement) so the rest of the stack never has to know views
//     exist.
//
//   - Workspace cross-check: a view must not leak across workspaces. If the
//     caller supplies both, we reject mismatches with a forbidden error rather
//     than silently trusting one side.
//
//   - Caller fields win over preset fields: the preset only fills gaps
//     (len == 0 / zero-value). This lets users override a saved view on the
//     fly without forcing them to re-specify every field.
//
//   - Topic defaulting: listening queries are meaningless without a topic
//     scope, and enumerating every topic client-side is fragile. When TopicIDs
//     are omitted we require a WorkspaceID and expand to all of that
//     workspace's topics — mirroring the Python service's behavior.
//
//   - Defensive cloning: the resolved filter is mutated (slice appends, field
//     assignments). Cloning the caller's input prevents accidental aliasing
//     of request-scoped data into shared/cached state.
package listening

import (
	"context"
	"fmt"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

type viewResolver interface {
	GetViewByID(ctx context.Context, viewID string) (*mongoModels.ListeningView, error)
}

type topicWorkspaceResolver interface {
	ListTopicsByWorkspace(ctx context.Context, workspaceID string) ([]*mongoModels.ListeningTopic, error)
}

type MentionFilterResolver struct {
	views  viewResolver
	topics topicWorkspaceResolver
}

func NewMentionFilterResolver(
	views viewResolver,
	topics topicWorkspaceResolver,
) *MentionFilterResolver {
	return &MentionFilterResolver{
		views:  views,
		topics: topics,
	}
}

func (r *MentionFilterResolver) Resolve(
	ctx context.Context,
	filter *apiModels.MentionFilter,
) (*apiModels.MentionFilter, error) {
	if filter == nil {
		return nil, httputil.NewValidationError("mention filter is required")
	}

	resolved := cloneMentionFilter(filter)

	if resolved.ViewID != "" {
		if r.views == nil {
			return nil, httputil.NewValidationError("view filters are unavailable")
		}

		view, err := r.views.GetViewByID(ctx, resolved.ViewID)
		if err != nil {
			return nil, fmt.Errorf("MentionFilterResolver.Resolve: get view: %w", err)
		}
		if view == nil {
			return nil, httputil.NewValidationError("view not found")
		}

		if resolved.WorkspaceID == "" {
			resolved.WorkspaceID = view.WorkspaceID
		} else if view.WorkspaceID != "" && view.WorkspaceID != resolved.WorkspaceID {
			return nil, httputil.NewForbiddenError("view does not belong to this workspace")
		}

		applyViewPreset(resolved, view.FilterPreset)
	}

	if len(resolved.TopicIDs) == 0 {
		if resolved.WorkspaceID == "" {
			return nil, httputil.NewValidationError("workspace_id is required when topic_ids[] are omitted")
		}
		if r.topics == nil {
			return nil, httputil.NewValidationError("workspace topics are unavailable")
		}

		topics, err := r.topics.ListTopicsByWorkspace(ctx, resolved.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("MentionFilterResolver.Resolve: list workspace topics: %w", err)
		}

		resolved.TopicIDs = make([]string, 0, len(topics))
		for _, topic := range topics {
			if topic == nil {
				continue
			}

			topicID := topic.TopicID
			if topicID == "" {
				topicID = topic.ID.Hex()
			}
			if topicID != "" {
				resolved.TopicIDs = append(resolved.TopicIDs, topicID)
			}
		}

		if len(resolved.TopicIDs) == 0 {
			return nil, httputil.NewValidationError("no listening topics found for this workspace")
		}
	}

	return resolved, nil
}

func cloneMentionFilter(filter *apiModels.MentionFilter) *apiModels.MentionFilter {
	if filter == nil {
		return nil
	}

	cloned := *filter
	cloned.TopicIDs = append([]string(nil), filter.TopicIDs...)
	cloned.Platforms = append([]string(nil), filter.Platforms...)
	cloned.Sentiments = append([]string(nil), filter.Sentiments...)
	cloned.AITags = append([]string(nil), filter.AITags...)
	cloned.ExcludeAITags = append([]string(nil), filter.ExcludeAITags...)
	cloned.Language = append([]string(nil), filter.Language...)

	return &cloned
}

func applyViewPreset(
	filter *apiModels.MentionFilter,
	preset mongoModels.ListeningViewFilterPreset,
) {
	if len(filter.TopicIDs) == 0 && len(preset.TopicIDs) > 0 {
		filter.TopicIDs = append([]string(nil), preset.TopicIDs...)
	}
	if len(filter.Platforms) == 0 && len(preset.Platforms) > 0 {
		filter.Platforms = append([]string(nil), preset.Platforms...)
	}
	if len(filter.Sentiments) == 0 && len(preset.Sentiments) > 0 {
		filter.Sentiments = append([]string(nil), preset.Sentiments...)
	}
	if len(filter.AITags) == 0 && len(preset.AITags) > 0 {
		filter.AITags = append([]string(nil), preset.AITags...)
	}
	if len(filter.ExcludeAITags) == 0 && len(preset.ExcludeAITags) > 0 {
		filter.ExcludeAITags = append([]string(nil), preset.ExcludeAITags...)
	}
	if len(filter.Language) == 0 && len(preset.Language) > 0 {
		filter.Language = append([]string(nil), preset.Language...)
	}
	if filter.MinFollowers <= 0 && preset.MinFollowers > 0 {
		filter.MinFollowers = preset.MinFollowers
	}
	if filter.MinTotalEngagement <= 0 && preset.MinTotalEngagement > 0 {
		filter.MinTotalEngagement = preset.MinTotalEngagement
	}
}
