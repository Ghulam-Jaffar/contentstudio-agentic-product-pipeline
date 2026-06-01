package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// MockUnifiedSocialRepository is a mock implementation of UnifiedSocialRepository for testing.
// It can be used by other packages that need to mock MongoDB operations.
type MockUnifiedSocialRepository struct {
	FindByIDFunc                                 func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
	GetByPlatformIDFunc                          func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error)
	GetValidAccountsFunc                         func(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error)
	GetAccountsByWorkspaceFunc                   func(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error)
	GetAccountsNeedingUpdateFunc                 func(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error)
	GetAccountsNeedingUpdatePaginatedFunc        func(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error)
	GetAccountsNeedingUpdateByIDFunc             func(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error)
	GetValidAccountsByIDFunc                     func(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error)
	CountValidAccountsFunc                       func(ctx context.Context, platformType string, accountTypes []string) (int64, error)
	GetAccountsByPlatformIDsFunc                 func(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error)
	CountAccountsNeedingUpdateFunc               func(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error)
	GetYouTubeAccountsNeedingUpdatePaginatedFunc func(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongomodels.SocialIntegration, error)
	GetYouTubeAccountsNeedingUpdateByIDFunc      func(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error)
	CountYouTubeAccountsNeedingUpdateFunc        func(ctx context.Context, hours int, consentDays int) (int64, error)
	UpdateFunc                                   func(ctx context.Context, id primitive.ObjectID, updates primitive.M) error
	UpdateAnalyticsTimestampFunc                 func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error
	UpdateTokensFunc                             func(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error
	UpdateStateFunc                              func(ctx context.Context, id primitive.ObjectID, state string) error
	UpdateValidityFunc                           func(ctx context.Context, id primitive.ObjectID, newValidity string) error
	RecordProcessingErrorFunc                    func(ctx context.Context, id primitive.ObjectID, errorMessage string) error
	ClearProcessingErrorFunc                     func(ctx context.Context, id primitive.ObjectID) error
	CreateFunc                                   func(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error)
	DeleteFunc                                   func(ctx context.Context, id primitive.ObjectID) error
	InsertTwitterJobMetadataFunc                 func(ctx context.Context, payload TwitterJobMetadataPayload) error
}

func (m *MockUnifiedSocialRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
	if m.GetByPlatformIDFunc != nil {
		return m.GetByPlatformIDFunc(ctx, platformType, platformID)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error) {
	if m.GetValidAccountsFunc != nil {
		return m.GetValidAccountsFunc(ctx, platformType, accountTypes)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.GetValidAccountsByIDFunc != nil {
		return m.GetValidAccountsByIDFunc(ctx, platformType, accountTypes, lastID, limit)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	if m.CountValidAccountsFunc != nil {
		return m.CountValidAccountsFunc(ctx, platformType, accountTypes)
	}
	return 0, nil
}

func (m *MockUnifiedSocialRepository) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongomodels.SocialIntegration, error) {
	if m.GetAccountsByPlatformIDsFunc != nil {
		return m.GetAccountsByPlatformIDsFunc(ctx, platformType, platformIDs)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error) {
	if m.GetAccountsByWorkspaceFunc != nil {
		return m.GetAccountsByWorkspaceFunc(ctx, workspaceID, platforms)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error) {
	if m.GetAccountsNeedingUpdateFunc != nil {
		return m.GetAccountsNeedingUpdateFunc(ctx, platformType, lastUpdateField, hours)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.GetAccountsNeedingUpdatePaginatedFunc != nil {
		return m.GetAccountsNeedingUpdatePaginatedFunc(ctx, platformType, accountTypes, hours, skip, limit)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	if m.CountAccountsNeedingUpdateFunc != nil {
		return m.CountAccountsNeedingUpdateFunc(ctx, platformType, accountTypes, hours)
	}
	return 0, nil
}

func (m *MockUnifiedSocialRepository) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.GetAccountsNeedingUpdateByIDFunc != nil {
		return m.GetAccountsNeedingUpdateByIDFunc(ctx, platformType, accountTypes, hours, lastID, limit)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.GetYouTubeAccountsNeedingUpdatePaginatedFunc != nil {
		return m.GetYouTubeAccountsNeedingUpdatePaginatedFunc(ctx, hours, consentDays, skip, limit)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongomodels.SocialIntegration, error) {
	if m.GetYouTubeAccountsNeedingUpdateByIDFunc != nil {
		return m.GetYouTubeAccountsNeedingUpdateByIDFunc(ctx, hours, consentDays, lastID, limit)
	}
	return nil, nil
}

func (m *MockUnifiedSocialRepository) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	if m.CountYouTubeAccountsNeedingUpdateFunc != nil {
		return m.CountYouTubeAccountsNeedingUpdateFunc(ctx, hours, consentDays)
	}
	return 0, nil
}

func (m *MockUnifiedSocialRepository) Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, id, updates)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
	if m.UpdateAnalyticsTimestampFunc != nil {
		return m.UpdateAnalyticsTimestampFunc(ctx, id, field, timestamp)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	if m.UpdateTokensFunc != nil {
		return m.UpdateTokensFunc(ctx, id, tokens)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) UpdateState(ctx context.Context, id primitive.ObjectID, state string) error {
	if m.UpdateStateFunc != nil {
		return m.UpdateStateFunc(ctx, id, state)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	if m.UpdateValidityFunc != nil {
		return m.UpdateValidityFunc(ctx, id, newValidity)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	if m.RecordProcessingErrorFunc != nil {
		return m.RecordProcessingErrorFunc(ctx, id, errorMessage)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	if m.ClearProcessingErrorFunc != nil {
		return m.ClearProcessingErrorFunc(ctx, id)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) Create(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, account)
	}
	return primitive.NilObjectID, nil
}

func (m *MockUnifiedSocialRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockUnifiedSocialRepository) InsertTwitterJobMetadata(ctx context.Context, payload TwitterJobMetadataPayload) error {
	if m.InsertTwitterJobMetadataFunc != nil {
		return m.InsertTwitterJobMetadataFunc(ctx, payload)
	}
	return nil
}

// Verify mock implements interface at compile time
var _ UnifiedSocialRepository = (*MockUnifiedSocialRepository)(nil)
