package mongodb

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongo3 "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func tblLogger() zerolog.Logger { return zerolog.New(io.Discard) }

/* small helper so every case gets a fresh harness that is properly started */
func runMT(t *testing.T, name string, fn func(*mt.T)) {
	t.Helper()
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
	h.Run(name, fn) // IMPORTANT: this “starts” the harness and initializes h.DB
}

func Test_FindByID_Table(t *testing.T) {
	id := primitive.NewObjectID()
	cases := []struct {
		name      string
		mock      func(*mt.T, string)
		expectNil bool
		expectErr bool
		expectID  primitive.ObjectID
	}{
		{
			name: "found",
			mock: func(m *mt.T, ns string) {
				doc := bson.D{
					{Key: "_id", Value: id},
					{Key: "platform_type", Value: mongo3.PlatformFacebook},
					{Key: "platform_identifier", Value: "fb_1"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectID: id,
		},
		{
			name: "not found -> (nil,nil)",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectNil: true,
		},
		{
			name: "driver error",
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "boom", Code: 1}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.FindByID(context.Background(), id)

				if tc.expectErr {
					if err == nil {
						m.Fatalf("want error, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if tc.expectNil {
					if got != nil {
						m.Fatalf("want nil, got %+v", got)
					}
					return
				}
				if got == nil || got.ID != tc.expectID {
					m.Fatalf("want id=%s, got %+v", tc.expectID.Hex(), got)
				}
			})
		})
	}
}

func Test_GetByPlatformID_Table(t *testing.T) {
	cases := []struct {
		name      string
		pt, pid   string
		mock      func(*mt.T, string)
		expectNil bool
		expectErr bool
		expectPID string
	}{
		{
			name: "primary hit",
			pt:   mongo3.PlatformInstagram, pid: "ig_9",
			mock: func(m *mt.T, ns string) {
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformInstagram},
					{Key: "platform_identifier", Value: "ig_9"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "ig_9",
		},
		{
			name: "fallback legacy hit",
			pt:   mongo3.PlatformFacebook, pid: "fb_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch)) // primary empty
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformFacebook},
					{Key: "facebook_id", Value: "fb_legacy"},
					{Key: "platform_identifier", Value: "fb_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch, doc))
			},
			expectPID: "fb_legacy",
		},
		{
			name: "not found",
			pt:   mongo3.PlatformLinkedIn, pid: "ln_0",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch)) // primary
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch)) // legacy
			},
			expectNil: true,
		},
		{
			name: "driver error",
			pt:   mongo3.PlatformInstagram, pid: "ig_err",
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 2}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetByPlatformID(context.Background(), tc.pt, tc.pid)

				if tc.expectErr {
					if err == nil {
						m.Fatalf("want error, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if tc.expectNil {
					if got != nil {
						m.Fatalf("want nil, got %+v", got)
					}
					return
				}
				if got == nil || got.PlatformIdentifier != tc.expectPID {
					m.Fatalf("want pid=%s, got %+v", tc.expectPID, got)
				}
			})
		})
	}
}

/* ========== GetValidAccounts (table) ========== */

func Test_GetValidAccounts_Table(t *testing.T) {
	doc1 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_1"},
		{Key: "type", Value: "page"},
		{Key: "state", Value: mongo3.StateAdded},
		{Key: "validity", Value: mongo3.ValidityValid},
	}
	doc2 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_2"},
		{Key: "type", Value: "page"},
		{Key: "state", Value: mongo3.StateProcessed},
		{Key: "validity", Value: mongo3.ValidityValid},
	}

	cases := []struct {
		name      string
		pt        string
		at        []string
		mock      func(*mt.T, string)
		wantCount int
		expectErr bool
	}{
		{
			name: "two docs",
			pt:   mongo3.PlatformFacebook, at: []string{"page"},
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1, doc2),
					mt.CreateCursorResponse(1, ns, mt.NextBatch),
				)
			},
			wantCount: 2,
		},
		{
			name: "driver error",
			pt:   mongo3.PlatformFacebook, at: nil,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 3}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetValidAccounts(context.Background(), tc.pt, tc.at)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					m.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_Update_Table(t *testing.T) {
	id := primitive.NewObjectID()

	cases := []struct {
		name       string
		updates    bson.M
		mock       func(*mt.T)
		expectErr  bool
		errIsNoDoc bool
	}{
		{
			name:    "no-op",
			updates: bson.M{},
			mock:    func(*mt.T) {},
		},
		{
			name:    "success",
			updates: bson.M{"state": mongo3.StateProcessed},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:    "not found -> ErrNoDocuments",
			updates: bson.M{"state": mongo3.StateProcessed},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 0},
					bson.E{Key: "nModified", Value: 0},
				))
			},
			expectErr: true, errIsNoDoc: true,
		},
		{
			name:    "driver error",
			updates: bson.M{"state": mongo3.StateProcessed},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "update fail", Code: 6}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := repo.Update(context.Background(), id, tc.updates)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					if tc.errIsNoDoc && err != mongo.ErrNoDocuments {
						m.Fatalf("want mongo.ErrNoDocuments, got %v", err)
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateAnalyticsTimestamp_Table(t *testing.T) {
	cases := []struct {
		name      string
		typ       string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:      "invalid type",
			typ:       "nope",
			mock:      func(*mt.T) {},
			expectErr: true,
		},
		{
			name: "success",
			typ:  "analytics",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := repo.UpdateAnalyticsTimestamp(context.Background(), primitive.NewObjectID(), tc.typ, time.Now())
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateTokens_Table(t *testing.T) {
	cases := []struct {
		name      string
		tokens    map[string]string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:   "valid token key",
			tokens: map[string]string{"access_token": "abc", "foo": "bar"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:      "no valid keys -> error",
			tokens:    map[string]string{"foo": "bar"},
			mock:      func(*mt.T) {},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := repo.UpdateTokens(context.Background(), primitive.NewObjectID(), tc.tokens)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateState_UpdateValidity_Table(t *testing.T) {
	type testCase struct {
		name      string
		call      func(UnifiedSocialRepository) error
		mock      func(*mt.T)
		expectErr bool
	}
	cases := []testCase{
		{
			name: "UpdateState success",
			call: func(r UnifiedSocialRepository) error {
				return r.UpdateState(context.Background(), primitive.NewObjectID(), mongo3.StateProcessed)
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "UpdateValidity success",
			call: func(r UnifiedSocialRepository) error {
				return r.UpdateValidity(context.Background(), primitive.NewObjectID(), mongo3.ValidityValid)
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "driver error",
			call: func(r UnifiedSocialRepository) error {
				return r.UpdateValidity(context.Background(), primitive.NewObjectID(), mongo3.ValidityInvalid)
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "fail", Code: 8}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := tc.call(repo)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_Create_Delete_Table(t *testing.T) {
	newOID := primitive.NewObjectID()

	cases := []struct {
		name      string
		call      func(UnifiedSocialRepository) error
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name: "Create success",
			call: func(r UnifiedSocialRepository) error {
				_, err := r.Create(context.Background(), &mongo3.SocialIntegration{
					PlatformType:       mongo3.PlatformFacebook,
					PlatformIdentifier: "fb_x",
				})
				return err
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "insertedId", Value: newOID},
				))
			},
		},
		{
			name: "Delete success",
			call: func(r UnifiedSocialRepository) error {
				return r.Delete(context.Background(), primitive.NewObjectID())
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "Create driver error",
			call: func(r UnifiedSocialRepository) error {
				_, err := r.Create(context.Background(), &mongo3.SocialIntegration{
					PlatformType:       mongo3.PlatformLinkedIn,
					PlatformIdentifier: "ln_x",
				})
				return err
			},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "insert fail", Code: 9}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := tc.call(repo)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_GetAccountsByWorkspace_Table(t *testing.T) {
	wsID := primitive.NewObjectID()
	doc1 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_1"},
		{Key: "workspace_id", Value: wsID},
		{Key: "state", Value: mongo3.StateAdded},
		{Key: "validity", Value: mongo3.ValidityValid},
	}
	doc2 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformInstagram},
		{Key: "platform_identifier", Value: "ig_1"},
		{Key: "workspace_id", Value: wsID},
		{Key: "state", Value: mongo3.StateProcessed},
		{Key: "validity", Value: mongo3.ValidityValid},
	}

	cases := []struct {
		name      string
		wsID      primitive.ObjectID
		platforms []string
		mock      func(*mt.T, string)
		wantCount int
		expectErr bool
	}{
		{
			name:      "two docs no platform filter",
			wsID:      wsID,
			platforms: nil,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1, doc2),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 2,
		},
		{
			name:      "with platform filter",
			wsID:      wsID,
			platforms: []string{mongo3.PlatformFacebook},
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:      "driver error",
			wsID:      wsID,
			platforms: nil,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 10}))
			},
			expectErr: true,
		},
		{
			name:      "empty result",
			wsID:      primitive.NewObjectID(),
			platforms: nil,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(0, ns, mt.FirstBatch),
				)
			},
			wantCount: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetAccountsByWorkspace(context.Background(), tc.wsID, tc.platforms)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					m.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_GetAccountsNeedingUpdate_Table(t *testing.T) {
	doc1 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_1"},
		{Key: "state", Value: mongo3.StateAdded},
		{Key: "validity", Value: mongo3.ValidityValid},
		{Key: "last_analytics_updated_at", Value: time.Now().Add(-48 * time.Hour)},
	}

	cases := []struct {
		name            string
		platformType    string
		lastUpdateField string
		hours           int
		mock            func(*mt.T, string)
		wantCount       int
		expectErr       bool
	}{
		{
			name:            "found accounts needing update",
			platformType:    mongo3.PlatformFacebook,
			lastUpdateField: "last_analytics_updated_at",
			hours:           24,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:            "no accounts found",
			platformType:    mongo3.PlatformInstagram,
			lastUpdateField: "last_analytics_updated_at",
			hours:           24,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(0, ns, mt.FirstBatch),
				)
			},
			wantCount: 0,
		},
		{
			name:            "driver error",
			platformType:    mongo3.PlatformFacebook,
			lastUpdateField: "last_analytics_updated_at",
			hours:           24,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 11}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetAccountsNeedingUpdate(context.Background(), tc.platformType, tc.lastUpdateField, tc.hours)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					m.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_GetAccountsNeedingUpdatePaginated_Table(t *testing.T) {
	doc1 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformLinkedIn},
		{Key: "platform_identifier", Value: "ln_1"},
		{Key: "type", Value: "page"},
		{Key: "state", Value: mongo3.StateAdded},
		{Key: "validity", Value: mongo3.ValidityValid},
	}
	doc2 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformLinkedIn},
		{Key: "platform_identifier", Value: "ln_2"},
		{Key: "type", Value: "page"},
		{Key: "state", Value: mongo3.StateSyncing},
		{Key: "validity", Value: mongo3.ValidityValid},
	}

	cases := []struct {
		name         string
		platformType string
		accountTypes []string
		hours        int
		skip, limit  int64
		mock         func(*mt.T, string)
		wantCount    int
		expectErr    bool
	}{
		{
			name:         "paginated results",
			platformType: mongo3.PlatformLinkedIn,
			accountTypes: []string{"page"},
			hours:        24,
			skip:         0,
			limit:        10,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1, doc2),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 2,
		},
		{
			name:         "with multiple account types",
			platformType: mongo3.PlatformLinkedIn,
			accountTypes: []string{"page", "profile"},
			hours:        24,
			skip:         0,
			limit:        10,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:         "empty account types",
			platformType: mongo3.PlatformLinkedIn,
			accountTypes: []string{},
			hours:        24,
			skip:         0,
			limit:        10,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:         "driver error",
			platformType: mongo3.PlatformLinkedIn,
			accountTypes: []string{"page"},
			hours:        24,
			skip:         0,
			limit:        10,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 12}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetAccountsNeedingUpdatePaginated(context.Background(), tc.platformType, tc.accountTypes, tc.hours, tc.skip, tc.limit)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					m.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_CountAccountsNeedingUpdate_Table(t *testing.T) {
	cases := []struct {
		name         string
		platformType string
		accountTypes []string
		hours        int
		mock         func(*mt.T, string)
		expectErr    bool
	}{
		{
			name:         "driver error",
			platformType: mongo3.PlatformFacebook,
			accountTypes: []string{"page"},
			hours:        24,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "count fail", Code: 13}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, "")
				_, err := repo.CountAccountsNeedingUpdate(context.Background(), tc.platformType, tc.accountTypes, tc.hours)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_GetByPlatformID_LegacyFallback_Table(t *testing.T) {
	cases := []struct {
		name      string
		pt, pid   string
		mock      func(*mt.T, string)
		expectNil bool
		expectErr bool
		expectPID string
	}{
		{
			name: "fallback Instagram legacy",
			pt:   mongo3.PlatformInstagram, pid: "ig_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformInstagram},
					{Key: "instagram_id", Value: "ig_legacy"},
					{Key: "platform_identifier", Value: "ig_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "ig_legacy",
		},
		{
			name: "fallback LinkedIn legacy",
			pt:   mongo3.PlatformLinkedIn, pid: "ln_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformLinkedIn},
					{Key: "linkedin_id", Value: "ln_legacy"},
					{Key: "platform_identifier", Value: "ln_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "ln_legacy",
		},
		{
			name: "fallback Twitter legacy",
			pt:   mongo3.PlatformTwitter, pid: "tw_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformTwitter},
					{Key: "twitter_id", Value: "tw_legacy"},
					{Key: "platform_identifier", Value: "tw_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "tw_legacy",
		},
		{
			name: "fallback GMB legacy",
			pt:   mongo3.PlatformGMB, pid: "gmb_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformGMB},
					{Key: "location_id", Value: "gmb_legacy"},
					{Key: "platform_identifier", Value: "gmb_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "gmb_legacy",
		},
		{
			name: "fallback Pinterest legacy",
			pt:   mongo3.PlatformPinterest, pid: "pin_legacy",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
				doc := bson.D{
					{Key: "_id", Value: primitive.NewObjectID()},
					{Key: "platform_type", Value: mongo3.PlatformPinterest},
					{Key: "pinterest_id", Value: "pin_legacy"},
					{Key: "platform_identifier", Value: "pin_legacy"},
					{Key: "state", Value: mongo3.StateAdded},
					{Key: "validity", Value: mongo3.ValidityValid},
				}
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			expectPID: "pin_legacy",
		},
		{
			name: "unknown platform returns nil",
			pt:   "unknown_platform", pid: "unknown_id",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectNil: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetByPlatformID(context.Background(), tc.pt, tc.pid)

				if tc.expectErr {
					if err == nil {
						m.Fatalf("want error, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if tc.expectNil {
					if got != nil {
						m.Fatalf("want nil, got %+v", got)
					}
					return
				}
				if got == nil || got.PlatformIdentifier != tc.expectPID {
					m.Fatalf("want pid=%s, got %+v", tc.expectPID, got)
				}
			})
		})
	}
}

func Test_GetValidAccounts_MultipleTypes_Table(t *testing.T) {
	doc1 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_1"},
		{Key: "type", Value: "page"},
		{Key: "state", Value: mongo3.StateAdded},
		{Key: "validity", Value: mongo3.ValidityValid},
	}
	doc2 := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "platform_type", Value: mongo3.PlatformFacebook},
		{Key: "platform_identifier", Value: "fb_2"},
		{Key: "type", Value: "profile"},
		{Key: "state", Value: mongo3.StateSyncing},
		{Key: "validity", Value: mongo3.ValidityValid},
	}

	cases := []struct {
		name         string
		pt           string
		accountTypes []string
		mock         func(*mt.T, string)
		wantCount    int
		expectErr    bool
	}{
		{
			name:         "multiple account types",
			pt:           mongo3.PlatformFacebook,
			accountTypes: []string{"page", "profile"},
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1, doc2),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 2,
		},
		{
			name:         "empty account types",
			pt:           mongo3.PlatformFacebook,
			accountTypes: []string{},
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1, doc2),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 2,
		},
		{
			name:         "single page type",
			pt:           mongo3.PlatformFacebook,
			accountTypes: []string{"page"},
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc1),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".social_integrations"
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m, ns)
				got, err := repo.GetValidAccounts(context.Background(), tc.pt, tc.accountTypes)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
					return
				}
				if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					m.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_UpdateTokens_AllKeys_Table(t *testing.T) {
	cases := []struct {
		name      string
		tokens    map[string]string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:   "refresh_token key",
			tokens: map[string]string{"refresh_token": "refresh123"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:   "long_access_token key",
			tokens: map[string]string{"long_access_token": "long123"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:   "oauth_token key",
			tokens: map[string]string{"oauth_token": "oauth123"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:   "oauth_token_secret key",
			tokens: map[string]string{"oauth_token_secret": "secret123"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:   "expires_at key ignored",
			tokens: map[string]string{"expires_at": "2025-01-01", "access_token": "abc"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:   "multiple valid keys",
			tokens: map[string]string{"access_token": "a", "refresh_token": "r", "oauth_token": "o"},
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := repo.UpdateTokens(context.Background(), primitive.NewObjectID(), tc.tokens)
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateAnalyticsTimestamp_AllTypes_Table(t *testing.T) {
	cases := []struct {
		name      string
		typ       string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name: "insights type",
			typ:  "insights",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "fans type",
			typ:  "fans",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "video type",
			typ:  "video",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "group type",
			typ:  "group",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name: "link_preview type",
			typ:  "link_preview",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger())

				tc.mock(m)
				err := repo.UpdateAnalyticsTimestamp(context.Background(), primitive.NewObjectID(), tc.typ, time.Now())
				if tc.expectErr {
					if err == nil {
						m.Fatalf("want err, got nil")
					}
				} else if err != nil {
					m.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_Create_WithExistingID_Table(t *testing.T) {
	existingID := primitive.NewObjectID()

	runMT(t, "create with existing ID", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger())

		m.AddMockResponses(mt.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "insertedId", Value: existingID},
		))

		account := &mongo3.SocialIntegration{
			ID:                 existingID,
			PlatformType:       mongo3.PlatformFacebook,
			PlatformIdentifier: "fb_existing",
		}

		id, err := repo.Create(context.Background(), account)
		if err != nil {
			m.Fatalf("unexpected err: %v", err)
		}
		if id != existingID {
			m.Fatalf("want id=%s, got %s", existingID.Hex(), id.Hex())
		}
	})
}

func Test_Create_WithTimestamps_Table(t *testing.T) {
	now := time.Now().UTC()
	createdAt := &mongo3.MongoTime{Time: now.Add(-24 * time.Hour)}
	updatedAt := &mongo3.MongoTime{Time: now}

	runMT(t, "create with timestamps", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger())

		newID := primitive.NewObjectID()
		m.AddMockResponses(mt.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "insertedId", Value: newID},
		))

		account := &mongo3.SocialIntegration{
			PlatformType:       mongo3.PlatformInstagram,
			PlatformIdentifier: "ig_with_ts",
			CreatedAt:          createdAt,
			UpdatedAt:          updatedAt,
		}

		_, err := repo.Create(context.Background(), account)
		if err != nil {
			m.Fatalf("unexpected err: %v", err)
		}
	})
}

func Test_Create_WithLegacyPlatformID_Table(t *testing.T) {
	runMT(t, "create with legacy platform ID", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger())

		newID := primitive.NewObjectID()
		m.AddMockResponses(mt.CreateSuccessResponse(
			bson.E{Key: "ok", Value: 1},
			bson.E{Key: "n", Value: 1},
			bson.E{Key: "insertedId", Value: newID},
		))

		account := &mongo3.SocialIntegration{
			PlatformType: mongo3.PlatformFacebook,
			FacebookID:   "fb_legacy_id",
		}

		_, err := repo.Create(context.Background(), account)
		if err != nil {
			m.Fatalf("unexpected err: %v", err)
		}
	})
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_MongoDB_ErrorsAreWarnLevel(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	runMT(t, "logging contract warn level", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, log.Logger)

		// Trigger a driver error on FindByID
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "boom", Code: 1}))
		_, err := repo.FindByID(context.Background(), primitive.NewObjectID())
		if err == nil {
			m.Fatal("expected error from FindByID")
		}

		output := buf.String()
		if !strings.Contains(output, "WRN") {
			m.Fatalf("expected WRN in log output, got: %s", output)
		}
		if strings.Contains(output, "ERR") {
			m.Fatalf("unexpected ERR-level log in output: %s", output)
		}
		if len(*captureRecords) != 0 {
			m.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
		}
	})
}

func TestLoggingContract_MongoDB_ErrorsReturnedToCaller(t *testing.T) {
	log, _ := logger.NewTestLoggerWithHook()

	runMT(t, "logging contract errors returned", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, log.Logger)

		// FindByID driver error
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 1}))
		_, err := repo.FindByID(context.Background(), primitive.NewObjectID())
		if err == nil {
			m.Fatal("FindByID: expected error to be returned to caller, got nil")
		}

		// GetByPlatformID driver error
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 2}))
		_, err = repo.GetByPlatformID(context.Background(), mongo3.PlatformFacebook, "fb_err")
		if err == nil {
			m.Fatal("GetByPlatformID: expected error to be returned to caller, got nil")
		}

		// GetValidAccounts driver error
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 3}))
		_, err = repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, []string{"page"})
		if err == nil {
			m.Fatal("GetValidAccounts: expected error to be returned to caller, got nil")
		}
	})
}

func Test_BuildNeedingUpdateFilter_IncludesStateFailed(t *testing.T) {
	runMT(t, "state filter includes Failed", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger()).(*unifiedSocialRepository)
		filter := repo.buildNeedingUpdateFilter(mongo3.PlatformFacebook, []string{"Page"}, 6)

		stateVal, exists := filter["state"]
		if !exists {
			m.Fatal("expected state filter to be present")
		}

		inFilter, ok := stateVal.(bson.M)
		if !ok {
			m.Fatalf("expected state to be bson.M, got %T", stateVal)
		}

		inValues, ok := inFilter["$in"].([]string)
		if !ok {
			m.Fatalf("expected $in to be []string, got %T", inFilter["$in"])
		}

		found := false
		for _, v := range inValues {
			if v == mongo3.StateFailed {
				found = true
				break
			}
		}
		if !found {
			m.Fatalf("expected state $in to include %q, got %v", mongo3.StateFailed, inValues)
		}
	})
}

func Test_BuildYouTubeFilter_IncludesStateFailed(t *testing.T) {
	runMT(t, "youtube state filter includes Failed", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger()).(*unifiedSocialRepository)
		filter := repo.buildYouTubeFilter(30)

		stateVal, exists := filter["state"]
		if !exists {
			m.Fatal("expected state filter to be present")
		}

		inFilter, ok := stateVal.(bson.M)
		if !ok {
			m.Fatalf("expected state to be bson.M, got %T", stateVal)
		}

		inValues, ok := inFilter["$in"].([]string)
		if !ok {
			m.Fatalf("expected $in to be []string, got %T", inFilter["$in"])
		}

		found := false
		for _, v := range inValues {
			if v == mongo3.StateFailed {
				found = true
				break
			}
		}
		if !found {
			m.Fatalf("expected state $in to include %q, got %v", mongo3.StateFailed, inValues)
		}
	})
}

func Test_BuildNeedingUpdateFilter_ExcludesSuperAdminState(t *testing.T) {
	cases := []struct {
		name         string
		platformType string
		accountTypes []string
		hours        int
	}{
		{
			name:         "facebook with page type",
			platformType: mongo3.PlatformFacebook,
			accountTypes: []string{"Page"},
			hours:        6,
		},
		{
			name:         "instagram no account types",
			platformType: mongo3.PlatformInstagram,
			accountTypes: nil,
			hours:        6,
		},
		{
			name:         "linkedin multiple types",
			platformType: mongo3.PlatformLinkedIn,
			accountTypes: []string{"Page", "Profile"},
			hours:        6,
		},
		{
			name:         "tiktok",
			platformType: mongo3.PlatformTikTok,
			accountTypes: []string{"Profile"},
			hours:        6,
		},
		{
			name:         "twitter",
			platformType: mongo3.PlatformTwitter,
			accountTypes: nil,
			hours:        6,
		},
		{
			name:         "pinterest",
			platformType: mongo3.PlatformPinterest,
			accountTypes: nil,
			hours:        6,
		},
		{
			name:         "meta ads",
			platformType: mongo3.PlatformMetaAds,
			accountTypes: nil,
			hours:        6,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runMT(t, tc.name, func(m *mt.T) {
				repo := NewUnifiedSocialRepository(m.DB, tblLogger()).(*unifiedSocialRepository)
				filter := repo.buildNeedingUpdateFilter(tc.platformType, tc.accountTypes, tc.hours)

				validityVal, exists := filter["validity"]
				if !exists {
					m.Fatal("expected validity filter to be present")
				}
				if validityVal != mongo3.ValidityValid {
					m.Fatalf("expected validity filter %q, got %v", mongo3.ValidityValid, validityVal)
				}

				superAdminVal, exists := filter["super_admin_state"]
				if !exists {
					m.Fatal("expected super_admin_state filter to be present")
				}

				inFilter, ok := superAdminVal.(bson.M)
				if !ok {
					m.Fatalf("expected super_admin_state to be bson.M, got %T", superAdminVal)
				}

				inValues, ok := inFilter["$in"].([]string)
				if !ok {
					m.Fatalf("expected $in to be []string, got %T", inFilter["$in"])
				}

				if len(inValues) != 2 {
					m.Fatalf("expected 2 super_admin_state values, got %d", len(inValues))
				}
				if inValues[0] != mongo3.SuperAdminStateActive {
					m.Fatalf("expected first value %q, got %q", mongo3.SuperAdminStateActive, inValues[0])
				}
				if inValues[1] != mongo3.SuperAdminStatePastDue {
					m.Fatalf("expected second value %q, got %q", mongo3.SuperAdminStatePastDue, inValues[1])
				}
			})
		})
	}
}

func Test_BuildYouTubeFilter_IncludesSuperAdminState(t *testing.T) {
	runMT(t, "youtube super_admin_state filter", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger()).(*unifiedSocialRepository)
		filter := repo.buildYouTubeFilter(30)

		superAdminVal, exists := filter["super_admin_state"]
		if !exists {
			m.Fatal("expected super_admin_state filter to be present in YouTube filter")
		}

		inFilter, ok := superAdminVal.(bson.M)
		if !ok {
			m.Fatalf("expected super_admin_state to be bson.M, got %T", superAdminVal)
		}

		inValues, ok := inFilter["$in"].([]string)
		if !ok {
			m.Fatalf("expected $in to be []string, got %T", inFilter["$in"])
		}

		if len(inValues) != 2 {
			m.Fatalf("expected 2 super_admin_state values, got %d", len(inValues))
		}
		if inValues[0] != mongo3.SuperAdminStateActive {
			m.Fatalf("expected first value %q, got %q", mongo3.SuperAdminStateActive, inValues[0])
		}
		if inValues[1] != mongo3.SuperAdminStatePastDue {
			m.Fatalf("expected second value %q, got %q", mongo3.SuperAdminStatePastDue, inValues[1])
		}

		if _, exists := filter["platform_type"]; !exists {
			m.Fatal("expected platform_type filter to be present")
		}
		if _, exists := filter["preferences.last_youtube_consent_time"]; !exists {
			m.Fatal("expected consent time filter to be present")
		}
	})
}

func Test_RecordProcessingError(t *testing.T) {
	runMT(t, "record processing error", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger())
		id := primitive.NewObjectID()

		m.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.RecordProcessingError(context.Background(), id, "InstagramClient.doWithRetry: instagram API error: Error validating access token: session expired (OAuthException/190)")
		if err != nil {
			m.Fatalf("expected no error, got %v", err)
		}

		startedEvents := m.GetAllStartedEvents()
		if len(startedEvents) != 1 {
			m.Fatalf("expected exactly 1 Mongo command for _id-scoped error update, got %d", len(startedEvents))
		}
		if startedEvents[0].CommandName != "update" {
			m.Fatalf("expected update command, got %s", startedEvents[0].CommandName)
		}
	})
}

func Test_ClearProcessingError(t *testing.T) {
	runMT(t, "clear processing error", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, tblLogger())
		id := primitive.NewObjectID()

		m.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "nModified", Value: 1},
			{Key: "n", Value: 1},
		})

		err := repo.ClearProcessingError(context.Background(), id)
		if err != nil {
			m.Fatalf("expected no error, got %v", err)
		}

		started := m.GetStartedEvent()
		if started == nil {
			m.Fatal("expected a started event")
		}
		if started.CommandName != "update" {
			m.Fatalf("expected update command, got %s", started.CommandName)
		}
	})
}

func Test_HasProcessingErrorMeta(t *testing.T) {
	tests := []struct {
		name string
		meta interface{}
		want bool
	}{
		{
			name: "nil metadata",
			meta: nil,
			want: false,
		},
		{
			name: "empty map metadata",
			meta: map[string]interface{}{},
			want: false,
		},
		{
			name: "string error in map",
			meta: map[string]interface{}{"last_processing_error": "token expired"},
			want: true,
		},
		{
			name: "blank string error in map",
			meta: map[string]interface{}{"last_processing_error": "   "},
			want: false,
		},
		{
			name: "bson metadata",
			meta: bson.M{"last_processing_error": "quota exceeded"},
			want: true,
		},
		{
			name: "primitive metadata",
			meta: primitive.M{"last_processing_error": "auth failed"},
			want: true,
		},
		{
			name: "string map metadata",
			meta: map[string]string{"last_processing_error": "access denied"},
			want: true,
		},
		{
			name: "primitive document metadata",
			meta: primitive.D{{Key: "last_processing_error", Value: "session expired"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasProcessingErrorMeta(tt.meta); got != tt.want {
				t.Fatalf("hasProcessingErrorMeta() = %v, want %v", got, tt.want)
			}
			if got := HasProcessingErrorMeta(tt.meta); got != tt.want {
				t.Fatalf("HasProcessingErrorMeta() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ProcessingErrorAccountIDs(t *testing.T) {
	idWithError := primitive.NewObjectID()
	idWithoutError := primitive.NewObjectID()

	accounts := []mongo3.SocialIntegration{
		{
			ID:       idWithError,
			MetaData: map[string]interface{}{"last_processing_error": "token expired"},
		},
		{
			ID:       idWithoutError,
			MetaData: map[string]interface{}{},
		},
		{
			ID: primitive.NilObjectID,
			MetaData: map[string]interface{}{
				"last_processing_error": "should be ignored because id is zero",
			},
		},
	}

	ids := processingErrorAccountIDs(accounts)
	if len(ids) != 1 {
		t.Fatalf("processingErrorAccountIDs() returned %d ids, want 1", len(ids))
	}
	if ids[0] != idWithError {
		t.Fatalf("processingErrorAccountIDs() returned %s, want %s", ids[0].Hex(), idWithError.Hex())
	}
}

func Test_CleanErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "instagram session invalidated by password change",
			input:    "InstagramClient.doWithRetry: instagram API error: Error validating access token: The session has been invalidated because the user changed their password or Facebook has changed the session for security reasons. (OAuthException/190)",
			expected: "Access token expired: user changed their password",
		},
		{
			name:     "instagram session expired with timestamp",
			input:    "InstagramClient.doWithRetry: instagram API error: Error validating access token: Session has expired on Monday, 16-Feb-26 23:46:54 PST. The current time is Monday, 09-Mar-26 03:10:35 PDT. (OAuthException/190)",
			expected: "Access token expired: session expired",
		},
		{
			name:     "linkedin multiline with JSON body",
			input:    "linkedin token invalid or expired\nLinkedInClient.FetchPostsPaginated: linkedin posts error: status 401 body {\"status\":401,\"serviceErrorCode\":65602,\"code\":\"EXPIRED_ACCESS_TOKEN\",\"message\":\"The token used in the request has expired\"}",
			expected: "Access token expired: token expired",
		},
		{
			name:     "linkedin share stats JSON",
			input:    "LinkedInClient.FetchShareStatisticsRaw: linkedin share statistics error status 401: {\"status\":401,\"message\":\"The token used in the request has expired\"}",
			expected: "Access token expired: token expired",
		},
		{
			name:     "facebook permission error",
			input:    "FacebookClient.doWithRetry: facebook API error: (#100) Pages manage metadata permission is needed to manage the Page (Type: OAuthException, Code: 100)",
			expected: "Insufficient permissions: required permissions not granted",
		},
		{
			name:     "youtube quota exceeded",
			input:    "YouTubeClient.FetchChannels: youtube API error (status 403): The request cannot be completed because you have exceeded your quota.",
			expected: "Rate limit exceeded: too many API requests",
		},
		{
			name:     "youtube unauthorized",
			input:    "unauthorized: YouTubeClient.makeRequest: request failed with status 401: unauthorized",
			expected: "Unauthorized: invalid credentials",
		},
		{
			name:     "tiktok expired token",
			input:    "TikTokClient.FetchUserVideos: tiktok api error (status 200): access_token_invalid - The access_token is invalid or has expired",
			expected: "Access token expired: token expired",
		},
		{
			name:     "twitter expired token",
			input:    "TwitterClient.FetchUserTweets: twitter api unauthorized (401): invalid or expired token",
			expected: "Access token expired: token expired",
		},
		{
			name:     "pinterest JSON auth failure",
			input:    "PinterestClient.makeRequest: pinterest API unauthorized (status 401): {\"code\":2,\"message\":\"Authentication failed.\"}",
			expected: "Unauthorized: invalid credentials",
		},
		{
			name:     "gmb nested JSON error",
			input:    "GMBClient.FetchPerformanceMetrics: API error (status 401): {\"error\":{\"code\":401,\"message\":\"Request had invalid authentication credentials.\",\"status\":\"UNAUTHENTICATED\"}}",
			expected: "Unauthorized: invalid credentials",
		},
		{
			name:     "youtube token revoked",
			input:    "YouTubeClient.FetchChannels: youtube API error (status 401): Token has been expired or revoked.",
			expected: "Access token expired: token expired",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "simple passthrough",
			input:    "connection timeout",
			expected: "connection timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanErrorMessage(tt.input)
			if result != tt.expected {
				t.Errorf("cleanErrorMessage(%q)\n  got:  %q\n  want: %q", tt.name, result, tt.expected)
			}
		})
	}
}

func TestLoggingContract_MongoDB_NoCaptureException(t *testing.T) {
	log, _ := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	runMT(t, "logging contract no capture", func(m *mt.T) {
		repo := NewUnifiedSocialRepository(m.DB, log.Logger)

		// Trigger multiple errors across different operations
		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "boom1", Code: 1}))
		_, _ = repo.FindByID(context.Background(), primitive.NewObjectID())

		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "boom2", Code: 2}))
		_, _ = repo.GetByPlatformID(context.Background(), mongo3.PlatformFacebook, "fb_x")

		m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "boom3", Code: 3}))
		_, _ = repo.GetValidAccounts(context.Background(), mongo3.PlatformFacebook, []string{"page"})

		if len(*captureRecords) != 0 {
			m.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
		}
	})
}
