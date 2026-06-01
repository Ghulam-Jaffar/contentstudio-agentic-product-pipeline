package mongodb

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

func testLogger() *logger.Logger {
	log, _ := logger.NewTestLogger()
	return log
}

func Test_NewCompetitorRepository(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
	h.Run("create repository", func(m *mt.T) {
		repo := NewCompetitorRepository(m.DB, testLogger())
		if repo == nil {
			t.Fatal("expected non-nil repository")
		}
		if repo.collection == nil {
			t.Fatal("expected non-nil collection")
		}
		if repo.reportsCollection == nil {
			t.Fatal("expected non-nil reports collection")
		}
	})
}

func Test_GetByCompetitorID(t *testing.T) {
	compID := "comp_123"
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "competitor_id", Value: compID},
		{Key: "state", Value: "active"},
		{Key: "image", Value: "http://example.com/img.png"},
		{Key: "error", Value: ""},
	}

	cases := []struct {
		name      string
		compID    string
		mock      func(*mt.T, string)
		wantCount int
		expectErr bool
	}{
		{
			name:   "found competitors",
			compID: compID,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:   "no competitors found",
			compID: "unknown",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(0, ns, mt.FirstBatch),
				)
			},
			wantCount: 0,
		},
		{
			name:   "driver error",
			compID: compID,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 1}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".competitors"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetByCompetitorID(context.Background(), tc.compID)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					t.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_GetByID(t *testing.T) {
	validID := primitive.NewObjectID()
	doc := bson.D{
		{Key: "_id", Value: validID},
		{Key: "competitor_id", Value: "comp_123"},
		{Key: "name", Value: "Test Competitor"},
		{Key: "state", Value: "active"},
	}

	cases := []struct {
		name      string
		id        string
		mock      func(*mt.T, string)
		expectErr bool
	}{
		{
			name: "found by ID",
			id:   validID.Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
		},
		{
			name:      "invalid object ID",
			id:        "invalid-id",
			mock:      func(m *mt.T, ns string) {},
			expectErr: true,
		},
		{
			name: "not found",
			id:   primitive.NewObjectID().Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectErr: true,
		},
		{
			name: "driver error",
			id:   validID.Hex(),
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 2}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".competitors"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetByID(context.Background(), tc.id)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if got == nil {
					t.Fatal("want non-nil competitor")
				}
			})
		})
	}
}

func Test_UpdateState(t *testing.T) {
	validID := primitive.NewObjectID()

	cases := []struct {
		name      string
		id        primitive.ObjectID
		state     string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:  "success",
			id:    validID,
			state: "active",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:  "not found",
			id:    validID,
			state: "active",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 0},
					bson.E{Key: "nModified", Value: 0},
				))
			},
			expectErr: true,
		},
		{
			name:  "driver error",
			id:    validID,
			state: "active",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "update fail", Code: 3}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m)
				err := repo.UpdateState(context.Background(), tc.id, tc.state)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateField(t *testing.T) {
	validID := primitive.NewObjectID()
	now := time.Now()

	cases := []struct {
		name      string
		id        primitive.ObjectID
		timestamp time.Time
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:      "success",
			id:        validID,
			timestamp: now,
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:      "not found",
			id:        validID,
			timestamp: now,
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 0},
					bson.E{Key: "nModified", Value: 0},
				))
			},
			expectErr: true,
		},
		{
			name:      "driver error",
			id:        validID,
			timestamp: now,
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "update fail", Code: 4}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m)
				err := repo.UpdateField(context.Background(), tc.id, tc.timestamp)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_AddError(t *testing.T) {
	validID := primitive.NewObjectID()

	cases := []struct {
		name      string
		id        primitive.ObjectID
		errorMsg  string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:     "success",
			id:       validID,
			errorMsg: "API rate limit exceeded",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:     "not found",
			id:       validID,
			errorMsg: "error",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 0},
					bson.E{Key: "nModified", Value: 0},
				))
			},
			expectErr: true,
		},
		{
			name:     "driver error",
			id:       validID,
			errorMsg: "error",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "update fail", Code: 5}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m)
				err := repo.AddError(context.Background(), tc.id, tc.errorMsg)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_UpdateImage(t *testing.T) {
	validID := primitive.NewObjectID()

	cases := []struct {
		name      string
		id        primitive.ObjectID
		imageURL  string
		mock      func(*mt.T)
		expectErr bool
	}{
		{
			name:     "success",
			id:       validID,
			imageURL: "https://example.com/new-image.png",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 1},
					bson.E{Key: "nModified", Value: 1},
				))
			},
		},
		{
			name:     "not found",
			id:       validID,
			imageURL: "https://example.com/image.png",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateSuccessResponse(
					bson.E{Key: "ok", Value: 1},
					bson.E{Key: "n", Value: 0},
					bson.E{Key: "nModified", Value: 0},
				))
			},
			expectErr: true,
		},
		{
			name:     "driver error",
			id:       validID,
			imageURL: "https://example.com/image.png",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "update fail", Code: 6}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m)
				err := repo.UpdateImage(context.Background(), tc.id, tc.imageURL)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
			})
		})
	}
}

func Test_GetReportsByCompetitorID(t *testing.T) {
	compID := "comp_123"
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "workspace_id", Value: primitive.NewObjectID()},
		{Key: "name", Value: "Test Report"},
		{Key: "competitors", Value: bson.A{compID}},
	}

	cases := []struct {
		name      string
		compID    string
		mock      func(*mt.T, string)
		wantCount int
		expectErr bool
	}{
		{
			name:   "found reports",
			compID: compID,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc),
					mt.CreateCursorResponse(0, ns, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:   "no reports found",
			compID: "unknown",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(0, ns, mt.FirstBatch),
				)
			},
			wantCount: 0,
		},
		{
			name:   "driver error",
			compID: compID,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 7}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".competitors_reports"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetReportsByCompetitorID(context.Background(), tc.compID)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					t.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_GetReportByID(t *testing.T) {
	validID := primitive.NewObjectID()
	doc := bson.D{
		{Key: "_id", Value: validID},
		{Key: "workspace_id", Value: primitive.NewObjectID()},
		{Key: "name", Value: "Test Report"},
		{Key: "competitors", Value: bson.A{"comp_1", "comp_2"}},
	}

	cases := []struct {
		name      string
		id        string
		mock      func(*mt.T, string)
		expectErr bool
	}{
		{
			name: "found report",
			id:   validID.Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
		},
		{
			name:      "invalid object ID",
			id:        "invalid-id",
			mock:      func(m *mt.T, ns string) {},
			expectErr: true,
		},
		{
			name: "not found",
			id:   primitive.NewObjectID().Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectErr: true,
		},
		{
			name: "driver error",
			id:   validID.Hex(),
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 8}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".competitors_reports"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetReportByID(context.Background(), tc.id)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if got == nil {
					t.Fatal("want non-nil report")
				}
			})
		})
	}
}

func Test_GetUserByID(t *testing.T) {
	validID := primitive.NewObjectID()
	doc := bson.D{
		{Key: "_id", Value: validID},
		{Key: "email", Value: "test@example.com"},
		{Key: "first_name", Value: "John"},
		{Key: "last_name", Value: "Doe"},
	}

	cases := []struct {
		name      string
		id        string
		mock      func(*mt.T, string)
		expectErr bool
	}{
		{
			name: "found user",
			id:   validID.Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
		},
		{
			name:      "invalid object ID",
			id:        "invalid-id",
			mock:      func(m *mt.T, ns string) {},
			expectErr: true,
		},
		{
			name: "not found",
			id:   primitive.NewObjectID().Hex(),
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectErr: true,
		},
		{
			name: "driver error",
			id:   validID.Hex(),
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 9}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".users"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetUserByID(context.Background(), tc.id)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if got == nil {
					t.Fatal("want non-nil user")
				}
			})
		})
	}
}

func Test_GetLastAnalyticsTime(t *testing.T) {
	compID := "comp_123"
	now := time.Now().UTC().Truncate(time.Second)
	doc := bson.D{
		{Key: "_id", Value: primitive.NewObjectID()},
		{Key: "competitor_id", Value: compID},
		{Key: "last_analytics_updated_at", Value: now},
	}

	cases := []struct {
		name       string
		compID     string
		mock       func(*mt.T, string)
		wantTime   time.Time
		expectZero bool
		expectErr  bool
	}{
		{
			name:   "found with timestamp",
			compID: compID,
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(1, ns, mt.FirstBatch, doc))
			},
			wantTime: now,
		},
		{
			name:   "not found returns zero time",
			compID: "unknown",
			mock: func(m *mt.T, ns string) {
				m.AddMockResponses(mt.CreateCursorResponse(0, ns, mt.FirstBatch))
			},
			expectZero: true,
		},
		{
			name:   "driver error",
			compID: compID,
			mock: func(m *mt.T, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "find fail", Code: 10}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				ns := mt.TestDb + ".competitors"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, ns)
				got, err := repo.GetLastAnalyticsTime(context.Background(), tc.compID)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if tc.expectZero {
					if !got.IsZero() {
						t.Fatalf("want zero time, got %v", got)
					}
					return
				}
				if !got.Equal(tc.wantTime) {
					t.Fatalf("want %v, got %v", tc.wantTime, got)
				}
			})
		})
	}
}

func Test_GetAccounts(t *testing.T) {
	compID := primitive.NewObjectID()
	doc := bson.D{
		{Key: "_id", Value: compID},
		{Key: "competitor_id", Value: "comp_123"},
		{Key: "name", Value: "Test Competitor"},
		{Key: "slug", Value: "test-competitor"},
		{Key: "platform_type", Value: mongoModels.PlatformFacebook},
		{Key: "is_active", Value: true},
	}

	cases := []struct {
		name         string
		platformType string
		mock         func(*mt.T, string, string)
		wantCount    int
		expectErr    bool
	}{
		{
			name:         "found accounts",
			platformType: mongoModels.PlatformFacebook,
			mock: func(m *mt.T, reportsNS, competitorsNS string) {
				m.AddMockResponses(
					mt.CreateCursorResponse(1, reportsNS, mt.FirstBatch, bson.D{
						{Key: "_id", Value: compID},
					}),
					mt.CreateCursorResponse(0, reportsNS, mt.NextBatch),
				)
				m.AddMockResponses(
					mt.CreateCursorResponse(1, competitorsNS, mt.FirstBatch, doc),
					mt.CreateCursorResponse(0, competitorsNS, mt.NextBatch),
				)
			},
			wantCount: 1,
		},
		{
			name:         "aggregate error",
			platformType: mongoModels.PlatformFacebook,
			mock: func(m *mt.T, _ string, _ string) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{Message: "aggregate fail", Code: 11}))
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				reportsNS := mt.TestDb + ".competitors_reports"
				competitorsNS := mt.TestDb + ".competitors"
				repo := NewCompetitorRepository(m.DB, testLogger())

				tc.mock(m, reportsNS, competitorsNS)
				got, err := repo.GetAccounts(context.Background(), tc.platformType)
				if tc.expectErr {
					if err == nil {
						t.Fatal("want err, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if len(got) != tc.wantCount {
					t.Fatalf("want %d, got %d", tc.wantCount, len(got))
				}
			})
		})
	}
}

func Test_Competitor_GetCompetitorIDAsString(t *testing.T) {
	cases := []struct {
		name     string
		comp     *mongoModels.Competitor
		expected string
	}{
		{
			name:     "string ID",
			comp:     &mongoModels.Competitor{CompetitorID: "comp_123"},
			expected: "comp_123",
		},
		{
			name:     "int64 ID",
			comp:     &mongoModels.Competitor{CompetitorID: int64(12345)},
			expected: "12345",
		},
		{
			name:     "float64 ID",
			comp:     &mongoModels.Competitor{CompetitorID: float64(67890)},
			expected: "67890",
		},
		{
			name:     "other type ID",
			comp:     &mongoModels.Competitor{CompetitorID: []byte("test")},
			expected: "[116 101 115 116]",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := tc.comp.GetCompetitorIDAsString()
			if got != tc.expected {
				t.Fatalf("want %q, got %q", tc.expected, got)
			}
		})
	}
}
