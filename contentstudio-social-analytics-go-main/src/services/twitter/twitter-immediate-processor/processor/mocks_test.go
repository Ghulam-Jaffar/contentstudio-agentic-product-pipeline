package processor

// ================== Mock Implementations ==================

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// MockTwitterTweetFetcher mocks the TwitterTweetFetcher interface.
type MockTwitterTweetFetcher struct {
	FetchUserTweetsFunc func(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error)
	FetchUserInfoFunc   func(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error)
}

func (m *MockTwitterTweetFetcher) FetchUserTweets(ctx context.Context, userID, oauthToken, oauthTokenSecret string, maxResults int, paginationToken string) (*social.TwitterTweetsResponse, error) {
	if m.FetchUserTweetsFunc != nil {
		return m.FetchUserTweetsFunc(ctx, userID, oauthToken, oauthTokenSecret, maxResults, paginationToken)
	}
	return &social.TwitterTweetsResponse{}, nil
}

func (m *MockTwitterTweetFetcher) FetchUserInfo(ctx context.Context, userIDs []string, oauthToken, oauthTokenSecret string) (*social.TwitterUserResponse, error) {
	if m.FetchUserInfoFunc != nil {
		return m.FetchUserInfoFunc(ctx, userIDs, oauthToken, oauthTokenSecret)
	}
	return &social.TwitterUserResponse{}, nil
}

// MockTwitterPostSink mocks the TwitterPostSink interface.
type MockTwitterPostSink struct {
	BulkInsertTwitterPostsFunc    func(ctx context.Context, posts []*chmodels.TwitterPosts) error
	BulkInsertTwitterInsightsFunc func(ctx context.Context, insights []*chmodels.TwitterInsights) error
	InsertedPosts                 []*chmodels.TwitterPosts
	InsertedInsights              []*chmodels.TwitterInsights
}

func (m *MockTwitterPostSink) BulkInsertTwitterPosts(ctx context.Context, posts []*chmodels.TwitterPosts) error {
	m.InsertedPosts = append(m.InsertedPosts, posts...)
	if m.BulkInsertTwitterPostsFunc != nil {
		return m.BulkInsertTwitterPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockTwitterPostSink) BulkInsertTwitterInsights(ctx context.Context, insights []*chmodels.TwitterInsights) error {
	m.InsertedInsights = append(m.InsertedInsights, insights...)
	if m.BulkInsertTwitterInsightsFunc != nil {
		return m.BulkInsertTwitterInsightsFunc(ctx, insights)
	}
	return nil
}

// MockSocialRepository mocks the SocialRepository interface.
type MockSocialRepository struct {
	FindByIDFunc                 func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
	UpdateFunc                   func(ctx context.Context, id primitive.ObjectID, updates bson.M) error
	InsertTwitterJobMetadataFunc func(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error
	RecordProcessingErrorFunc    func(ctx context.Context, id primitive.ObjectID, errorMessage string) error
	ClearProcessingErrorFunc     func(ctx context.Context, id primitive.ObjectID) error
	ClearProcessingErrorCalls    []primitive.ObjectID
}

func (m *MockSocialRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockSocialRepository) Update(ctx context.Context, id primitive.ObjectID, updates bson.M) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil
}

func (m *MockSocialRepository) InsertTwitterJobMetadata(ctx context.Context, payload mongodb.TwitterJobMetadataPayload) error {
	if m.InsertTwitterJobMetadataFunc != nil {
		return m.InsertTwitterJobMetadataFunc(ctx, payload)
	}
	return nil
}

func (m *MockSocialRepository) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	if m.RecordProcessingErrorFunc != nil {
		return m.RecordProcessingErrorFunc(ctx, id, errorMessage)
	}
	return nil
}

func (m *MockSocialRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	m.ClearProcessingErrorCalls = append(m.ClearProcessingErrorCalls, id)
	if m.ClearProcessingErrorFunc != nil {
		return m.ClearProcessingErrorFunc(ctx, id)
	}
	return nil
}

// MockPusherNotifier mocks the PusherClientInterface.
type MockPusherNotifier struct {
	TriggerFunc func(channel, event string, data interface{}) error
	Triggered   []struct {
		Channel string
		Event   string
		Data    interface{}
	}
}

func (m *MockPusherNotifier) Trigger(channel, event string, data interface{}) error {
	m.Triggered = append(m.Triggered, struct {
		Channel string
		Event   string
		Data    interface{}
	}{channel, event, data})
	if m.TriggerFunc != nil {
		return m.TriggerFunc(channel, event, data)
	}
	return nil
}

// MockEmailNotifier mocks the NotifierInterface.
type MockEmailNotifier struct {
	SendFunc func(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
	Sent     []struct {
		UserID, WorkspaceID, Platform, AccountID, AccountName string
		IsCompetitor                                          bool
	}
}

func (m *MockEmailNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	m.Sent = append(m.Sent, struct {
		UserID, WorkspaceID, Platform, AccountID, AccountName string
		IsCompetitor                                          bool
	}{userID, workspaceID, platform, accountID, accountName, isCompetitor})
	if m.SendFunc != nil {
		return m.SendFunc(userID, workspaceID, platform, accountID, accountName, isCompetitor)
	}
	return nil
}
