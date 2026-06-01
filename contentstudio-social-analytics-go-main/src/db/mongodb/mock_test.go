package mongodb

import (
	"context"
	"errors"
	"testing"
	"time"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMockUnifiedSocialRepository_FindByID(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()

	// Test with nil function
	result, err := mock.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccount := &mongomodels.SocialIntegration{ID: id}
	mock.FindByIDFunc = func(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error) {
		return expectedAccount, nil
	}
	result, err = mock.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expectedAccount {
		t.Fatal("expected account to match")
	}
}

func TestMockUnifiedSocialRepository_GetByPlatformID(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}

	// Test with nil function
	result, err := mock.GetByPlatformID(context.Background(), "facebook", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccount := &mongomodels.SocialIntegration{ID: primitive.NewObjectID()}
	mock.GetByPlatformIDFunc = func(ctx context.Context, platformType, platformID string) (*mongomodels.SocialIntegration, error) {
		return expectedAccount, nil
	}
	result, err = mock.GetByPlatformID(context.Background(), "facebook", "123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expectedAccount {
		t.Fatal("expected account to match")
	}
}

func TestMockUnifiedSocialRepository_GetValidAccounts(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}

	// Test with nil function
	result, err := mock.GetValidAccounts(context.Background(), "facebook", []string{"page"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccounts := []mongomodels.SocialIntegration{{ID: primitive.NewObjectID()}}
	mock.GetValidAccountsFunc = func(ctx context.Context, platformType string, accountTypes []string) ([]mongomodels.SocialIntegration, error) {
		return expectedAccounts, nil
	}
	result, err = mock.GetValidAccounts(context.Background(), "facebook", []string{"page"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
}

func TestMockUnifiedSocialRepository_GetAccountsByWorkspace(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	workspaceID := primitive.NewObjectID()

	// Test with nil function
	result, err := mock.GetAccountsByWorkspace(context.Background(), workspaceID, []string{"facebook"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccounts := []mongomodels.SocialIntegration{{ID: primitive.NewObjectID()}}
	mock.GetAccountsByWorkspaceFunc = func(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongomodels.SocialIntegration, error) {
		return expectedAccounts, nil
	}
	result, err = mock.GetAccountsByWorkspace(context.Background(), workspaceID, []string{"facebook"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
}

func TestMockUnifiedSocialRepository_GetAccountsNeedingUpdate(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}

	// Test with nil function
	result, err := mock.GetAccountsNeedingUpdate(context.Background(), "facebook", "last_analytics_update", 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccounts := []mongomodels.SocialIntegration{{ID: primitive.NewObjectID()}}
	mock.GetAccountsNeedingUpdateFunc = func(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongomodels.SocialIntegration, error) {
		return expectedAccounts, nil
	}
	result, err = mock.GetAccountsNeedingUpdate(context.Background(), "facebook", "last_analytics_update", 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 account, got %d", len(result))
	}
}

func TestMockUnifiedSocialRepository_GetAccountsNeedingUpdatePaginated(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}

	// Test with nil function
	result, err := mock.GetAccountsNeedingUpdatePaginated(context.Background(), "facebook", []string{"page"}, 24, 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedAccounts := []mongomodels.SocialIntegration{{ID: primitive.NewObjectID()}, {ID: primitive.NewObjectID()}}
	mock.GetAccountsNeedingUpdatePaginatedFunc = func(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongomodels.SocialIntegration, error) {
		return expectedAccounts, nil
	}
	result, err = mock.GetAccountsNeedingUpdatePaginated(context.Background(), "facebook", []string{"page"}, 24, 0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(result))
	}
}

func TestMockUnifiedSocialRepository_CountAccountsNeedingUpdate(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}

	// Test with nil function
	count, err := mock.CountAccountsNeedingUpdate(context.Background(), "facebook", []string{"page"}, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	// Test with custom function
	mock.CountAccountsNeedingUpdateFunc = func(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
		return 42, nil
	}
	count, err = mock.CountAccountsNeedingUpdate(context.Background(), "facebook", []string{"page"}, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 42 {
		t.Fatalf("expected 42, got %d", count)
	}
}

func TestMockUnifiedSocialRepository_Update(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()
	updates := primitive.M{"$set": primitive.M{"state": "active"}}

	// Test with nil function
	err := mock.Update(context.Background(), id, updates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.UpdateFunc = func(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
		return nil
	}
	err = mock.Update(context.Background(), id, updates)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.UpdateFunc = func(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
		return errors.New("update failed")
	}
	err = mock.Update(context.Background(), id, updates)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockUnifiedSocialRepository_UpdateAnalyticsTimestamp(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()
	timestamp := time.Now()

	// Test with nil function
	err := mock.UpdateAnalyticsTimestamp(context.Background(), id, "last_analytics_update", timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.UpdateAnalyticsTimestampFunc = func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
		return nil
	}
	err = mock.UpdateAnalyticsTimestamp(context.Background(), id, "last_analytics_update", timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockUnifiedSocialRepository_UpdateTokens(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()
	tokens := map[string]string{"access_token": "new_token"}

	// Test with nil function
	err := mock.UpdateTokens(context.Background(), id, tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.UpdateTokensFunc = func(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
		return nil
	}
	err = mock.UpdateTokens(context.Background(), id, tokens)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockUnifiedSocialRepository_UpdateState(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()

	// Test with nil function
	err := mock.UpdateState(context.Background(), id, "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.UpdateStateFunc = func(ctx context.Context, id primitive.ObjectID, state string) error {
		return nil
	}
	err = mock.UpdateState(context.Background(), id, "active")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockUnifiedSocialRepository_UpdateValidity(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()

	// Test with nil function
	err := mock.UpdateValidity(context.Background(), id, "valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.UpdateValidityFunc = func(ctx context.Context, id primitive.ObjectID, newValidity string) error {
		return nil
	}
	err = mock.UpdateValidity(context.Background(), id, "valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMockUnifiedSocialRepository_Create(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	account := &mongomodels.SocialIntegration{}

	// Test with nil function
	id, err := mock.Create(context.Background(), account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != primitive.NilObjectID {
		t.Fatal("expected NilObjectID")
	}

	// Test with custom function
	expectedID := primitive.NewObjectID()
	mock.CreateFunc = func(ctx context.Context, account *mongomodels.SocialIntegration) (primitive.ObjectID, error) {
		return expectedID, nil
	}
	id, err = mock.Create(context.Background(), account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != expectedID {
		t.Fatal("expected ID to match")
	}
}

func TestMockUnifiedSocialRepository_Delete(t *testing.T) {
	mock := &MockUnifiedSocialRepository{}
	id := primitive.NewObjectID()

	// Test with nil function
	err := mock.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.DeleteFunc = func(ctx context.Context, id primitive.ObjectID) error {
		return nil
	}
	err = mock.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.DeleteFunc = func(ctx context.Context, id primitive.ObjectID) error {
		return errors.New("delete failed")
	}
	err = mock.Delete(context.Background(), id)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockUnifiedSocialRepository_ImplementsInterface(t *testing.T) {
	var _ UnifiedSocialRepository = (*MockUnifiedSocialRepository)(nil)
}
