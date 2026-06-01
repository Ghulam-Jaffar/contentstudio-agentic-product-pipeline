package mongodb

import (
	"context"
	"testing"
	"time"

	mongo3 "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_SeedRepo_NewAndClear(t *testing.T) {
	repo := NewSeedRepo()
	if repo == nil {
		t.Fatal("expected non-nil repo")
	}

	repo.SeedAccounts(mongo3.SocialIntegration{
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	accounts, _ := repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, nil)
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}

	repo.Clear()
	accounts, _ = repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, nil)
	if len(accounts) != 0 {
		t.Fatalf("expected 0 accounts after clear, got %d", len(accounts))
	}
}

func Test_SeedRepo_SeedAccounts(t *testing.T) {
	repo := NewSeedRepo()

	repo.SeedAccounts(
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_1",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			Type:               "page",
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformInstagram,
			PlatformIdentifier: "ig_1",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			Type:               "business",
			AccessToken:        "encrypted_token",
		},
	)

	fbAccounts, _ := repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, []string{"page"})
	if len(fbAccounts) != 1 {
		t.Fatalf("expected 1 Facebook account, got %d", len(fbAccounts))
	}

	igAccounts, _ := repo.GetValidAccounts(context.Background(), mongo3.PlatformInstagram, []string{"business"})
	if len(igAccounts) != 1 {
		t.Fatalf("expected 1 Instagram account, got %d", len(igAccounts))
	}
}

func Test_SeedRepo_FindByID(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	found, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find account")
	}
	if found.PlatformIdentifier != "fb_1" {
		t.Fatalf("expected fb_1, got %s", found.PlatformIdentifier)
	}

	notFound, err := repo.FindByID(context.Background(), primitive.NewObjectID())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for non-existent ID")
	}
}

func Test_SeedRepo_GetByPlatformID(t *testing.T) {
	repo := NewSeedRepo()

	repo.SeedAccounts(mongo3.SocialIntegration{
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_123",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	found, err := repo.GetByPlatformID(context.Background(), mongo3.PlatformFacebook, "fb_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found == nil {
		t.Fatal("expected to find account")
	}

	notFound, err := repo.GetByPlatformID(context.Background(), mongo3.PlatformFacebook, "unknown")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Fatal("expected nil for non-existent platform ID")
	}
}

func Test_SeedRepo_GetValidAccounts(t *testing.T) {
	repo := NewSeedRepo()

	repo.SeedAccounts(
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_1",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			Type:               "page",
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_2",
			State:              mongo3.StateSyncing,
			Validity:           mongo3.ValidityValid,
			Type:               "profile",
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_3",
			State:              mongo3.StateDeleted,
			Validity:           mongo3.ValidityValid,
			Type:               "page",
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_4",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityInvalid,
			Type:               "page",
		},
	)

	all, _ := repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, nil)
	if len(all) != 2 {
		t.Fatalf("expected 2 valid accounts, got %d", len(all))
	}

	pages, _ := repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, []string{"page"})
	if len(pages) != 1 {
		t.Fatalf("expected 1 page account, got %d", len(pages))
	}
}

func Test_SeedRepo_GetAccountsByWorkspace(t *testing.T) {
	repo := NewSeedRepo()
	wsID := primitive.NewObjectID()

	repo.SeedAccounts(
		mongo3.SocialIntegration{
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_1",
			WorkspaceID:        wsID,
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
		},
		mongo3.SocialIntegration{
			PlatformType:       mongo3.PlatformInstagram,
			PlatformIdentifier: "ig_1",
			WorkspaceID:        wsID,
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
		},
		mongo3.SocialIntegration{
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_deleted",
			WorkspaceID:        wsID,
			State:              mongo3.StateDeleted,
			Validity:           mongo3.ValidityValid,
		},
	)

	all, err := repo.GetAccountsByWorkspace(context.Background(), wsID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(all))
	}

	fbOnly, err := repo.GetAccountsByWorkspace(context.Background(), wsID, []string{mongo3.PlatformFacebook})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fbOnly) != 1 {
		t.Fatalf("expected 1 Facebook account, got %d", len(fbOnly))
	}

	otherWS, _ := repo.GetAccountsByWorkspace(context.Background(), primitive.NewObjectID(), nil)
	if len(otherWS) != 0 {
		t.Fatalf("expected 0 accounts for other workspace, got %d", len(otherWS))
	}
}

func Test_SeedRepo_GetAccountsNeedingUpdate(t *testing.T) {
	repo := NewSeedRepo()

	oldTime := time.Now().Add(-48 * time.Hour)
	repo.SeedAccounts(
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_old",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": oldTime,
			},
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_null",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_recent",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": time.Now(),
			},
		},
	)

	accounts, err := repo.GetAccountsNeedingUpdate(context.Background(), mongo3.PlatformFacebook, "last_analytics_updated_at", 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 2 {
		t.Fatalf("expected 2 accounts needing update, got %d", len(accounts))
	}
}

func Test_SeedRepo_GetAccountsNeedingUpdatePaginated(t *testing.T) {
	repo := NewSeedRepo()

	for i := 0; i < 5; i++ {
		repo.SeedAccounts(mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_" + string(rune('a'+i)),
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			Type:               "page",
		})
	}

	page1, err := repo.GetAccountsNeedingUpdatePaginated(context.Background(), mongo3.PlatformFacebook, []string{"page"}, 24, 0, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page1) != 2 {
		t.Fatalf("expected 2 accounts on page 1, got %d", len(page1))
	}

	page2, err := repo.GetAccountsNeedingUpdatePaginated(context.Background(), mongo3.PlatformFacebook, []string{"page"}, 24, 2, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(page2) != 2 {
		t.Fatalf("expected 2 accounts on page 2, got %d", len(page2))
	}

	beyondEnd, err := repo.GetAccountsNeedingUpdatePaginated(context.Background(), mongo3.PlatformFacebook, []string{"page"}, 24, 10, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(beyondEnd) != 0 {
		t.Fatalf("expected 0 accounts beyond end, got %d", len(beyondEnd))
	}
}

func Test_SeedRepo_CountAccountsNeedingUpdate(t *testing.T) {
	repo := NewSeedRepo()

	for i := 0; i < 5; i++ {
		repo.SeedAccounts(mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_" + string(rune('a'+i)),
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			Type:               "page",
		})
	}

	count, err := repo.CountAccountsNeedingUpdate(context.Background(), mongo3.PlatformFacebook, []string{"page"}, 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected count 5, got %d", count)
	}
}

func Test_SeedRepo_Update(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.Update(context.Background(), id, primitive.M{"state": mongo3.StateProcessed})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found.State != mongo3.StateProcessed {
		t.Fatalf("expected state %s, got %s", mongo3.StateProcessed, found.State)
	}

	err = repo.Update(context.Background(), primitive.NewObjectID(), primitive.M{"state": "x"})
	if err == nil {
		t.Fatal("expected error for non-existent ID")
	}
}

func Test_SeedRepo_UpdateAnalyticsTimestamp(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	now := time.Now().UTC()
	err := repo.UpdateAnalyticsTimestamp(context.Background(), id, "analytics", now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = repo.UpdateAnalyticsTimestamp(context.Background(), id, "invalid_type", now)
	if err == nil {
		t.Fatal("expected error for invalid timestamp type")
	}
}

func Test_SeedRepo_UpdateTokens(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.UpdateTokens(context.Background(), id, map[string]string{
		"access_token":  "new_token",
		"refresh_token": "refresh",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = repo.UpdateTokens(context.Background(), id, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty tokens")
	}

	err = repo.UpdateTokens(context.Background(), id, map[string]string{"unknown_key": "value"})
	if err == nil {
		t.Fatal("expected error for invalid token keys")
	}
}

func Test_SeedRepo_UpdateState(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.UpdateState(context.Background(), id, mongo3.StateProcessed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found.State != mongo3.StateProcessed {
		t.Fatalf("expected state %s, got %s", mongo3.StateProcessed, found.State)
	}
}

func Test_SeedRepo_UpdateValidity(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.UpdateValidity(context.Background(), id, mongo3.ValidityInvalid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found.Validity != mongo3.ValidityInvalid {
		t.Fatalf("expected validity %s, got %s", mongo3.ValidityInvalid, found.Validity)
	}
}

func Test_SeedRepo_Create(t *testing.T) {
	repo := NewSeedRepo()

	account := &mongo3.SocialIntegration{
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_new",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	}

	id, err := repo.Create(context.Background(), account)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.IsZero() {
		t.Fatal("expected non-zero ID")
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found == nil {
		t.Fatal("expected to find created account")
	}

	_, err = repo.Create(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil account")
	}
}

func Test_SeedRepo_Delete(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.Delete(context.Background(), id)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found.State != mongo3.StateDeleted {
		t.Fatalf("expected state %s, got %s", mongo3.StateDeleted, found.State)
	}
}

func Test_EncryptToken(t *testing.T) {
	key := "01234567890123456789012345678901"
	base64Key := "MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTIzNDU2Nzg5MDE="

	encrypted, err := EncryptToken("my_secret_token", base64Key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if encrypted == "" {
		t.Fatal("expected non-empty encrypted token")
	}

	_, err = EncryptToken("token", "invalid_base64")
	if err == nil {
		t.Fatal("expected error for invalid base64 key")
	}

	shortKey := "MTIzNDU2Nzg5MDEyMzQ1Ng=="
	_, err = EncryptToken("token", shortKey)
	if err == nil {
		t.Fatal("expected error for short key")
	}

	_ = key
}

func Test_SeedRepo_Update_ExtraDataFields(t *testing.T) {
	repo := NewSeedRepo()
	id := primitive.NewObjectID()

	repo.SeedAccounts(mongo3.SocialIntegration{
		ID:                 id,
		PlatformType:       mongo3.PlatformFacebook,
		PlatformIdentifier: "fb_1",
		State:              mongo3.StateAdded,
		Validity:           mongo3.ValidityValid,
	})

	err := repo.Update(context.Background(), id, primitive.M{
		"long_access_token":  "long_token_value",
		"oauth_token":        "oauth_value",
		"oauth_token_secret": "secret_value",
		"custom_field":       "custom_value",
		"updated_at":         time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found, _ := repo.FindByID(context.Background(), id)
	if found.ExtraData == nil {
		t.Fatal("expected ExtraData to be populated")
	}
	if found.ExtraData["long_access_token"] != "long_token_value" {
		t.Fatal("expected long_access_token to be set in ExtraData")
	}
}

func Test_SeedRepo_GetAccountsNeedingUpdate_TimestampTypes(t *testing.T) {
	repo := NewSeedRepo()

	oldTime := time.Now().Add(-48 * time.Hour)
	ptrTime := &oldTime
	mongoTime := mongo3.MongoTime{Time: oldTime}
	ptrMongoTime := &mongoTime

	repo.SeedAccounts(
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_time",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": oldTime,
			},
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_ptr_time",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": ptrTime,
			},
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_mongo_time",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": mongoTime,
			},
		},
		mongo3.SocialIntegration{
			ID:                 primitive.NewObjectID(),
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_ptr_mongo_time",
			State:              mongo3.StateAdded,
			Validity:           mongo3.ValidityValid,
			ExtraData: map[string]interface{}{
				"last_analytics_updated_at": ptrMongoTime,
			},
		},
	)

	accounts, err := repo.GetAccountsNeedingUpdate(context.Background(), mongo3.PlatformFacebook, "last_analytics_updated_at", 24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(accounts) != 4 {
		t.Fatalf("expected 4 accounts, got %d", len(accounts))
	}
}
