package mongodb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func Test_NewApiKeyRepository(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
	h.Run("creates repository with collection", func(m *mt.T) {
		repo := NewApiKeyRepository(m.DB, testLogger())
		if repo == nil {
			t.Fatal("expected non-nil repository")
		}
		if repo.collection == nil {
			t.Fatal("expected non-nil collection")
		}
	})
}

func Test_FindValidByKey(t *testing.T) {
	validID := primitive.NewObjectID()
	validKey := "cs_abc123validkey"

	doc := bson.D{
		{Key: "_id", Value: validID},
		{Key: "key", Value: validKey},
		{Key: "revoked", Value: false},
		{Key: "ai_creation", Value: false},
	}

	cases := []struct {
		name      string
		key       string
		mock      func(*mt.T, string)
		wantKey   string
		wantNil   bool
		expectErr bool
	}{
		{
			name: "valid key found",
			key:  validKey,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			wantKey: validKey,
		},
		{
			name: "key not found returns nil without error",
			key:  "cs_unknownkey",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			wantNil: true,
		},
		{
			name: "driver error returns error",
			key:  validKey,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{
					Message: "find failed",
					Code:    1,
				}))
			},
			expectErr: true,
		},
		{
			name: "empty key returns nil without error",
			key:  "",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			wantNil: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".api_keys"
				repo := NewApiKeyRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.FindValidByKey(context.Background(), tc.key)

				if tc.expectErr {
					if err == nil {
						t.Fatal("want error, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tc.wantNil {
					if got != nil {
						t.Fatalf("want nil, got %+v", got)
					}
					return
				}
				if got == nil {
					t.Fatal("want non-nil ApiKey, got nil")
				}
				if got.Key != tc.wantKey {
					t.Fatalf("want key %q, got %q", tc.wantKey, got.Key)
				}
			})
		})
	}
}
