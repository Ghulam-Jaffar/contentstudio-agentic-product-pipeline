package mongodb

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mt "go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func Test_NewShareableLinkRepository(t *testing.T) {
	h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
	h.Run("create repository", func(m *mt.T) {
		repo := NewShareableLinkRepository(m.DB, testLogger())
		if repo == nil {
			t.Fatal("expected non-nil repository")
		}
		if repo.shareLinksCollection == nil {
			t.Fatal("expected non-nil share links collection")
		}
		if repo.usersCollection == nil {
			t.Fatal("expected non-nil users collection")
		}
	})
}

func Test_FindActiveUserIDByLinkID(t *testing.T) {
	userOID := primitive.NewObjectID()
	userHex := userOID.Hex()

	cases := []struct {
		name      string
		linkID    string
		mock      func(*mt.T)
		wantUser  string
		expectErr bool
	}{
		{
			name:   "valid link with string user id and user exists",
			linkID: "link-string-user",
			mock: func(m *mt.T) {
				shareLinkNS := mt.TestDb + ".analytics_share_links"
				usersNS := mt.TestDb + ".users"

				m.AddMockResponses(
					mt.CreateCursorResponse(0, shareLinkNS, mt.FirstBatch,
						bson.D{{Key: "link_id", Value: "link-string-user"}, {Key: "user_id", Value: userHex}},
					),
					mt.CreateCursorResponse(0, usersNS, mt.FirstBatch,
						bson.D{{Key: "_id", Value: userHex}},
					),
				)
			},
			wantUser: userHex,
		},
		{
			name:   "valid link with object id user and user exists",
			linkID: "link-oid-user",
			mock: func(m *mt.T) {
				shareLinkNS := mt.TestDb + ".analytics_share_links"
				usersNS := mt.TestDb + ".users"

				m.AddMockResponses(
					mt.CreateCursorResponse(0, shareLinkNS, mt.FirstBatch,
						bson.D{{Key: "link_id", Value: "link-oid-user"}, {Key: "user_id", Value: userOID}},
					),
					mt.CreateCursorResponse(0, usersNS, mt.FirstBatch,
						bson.D{{Key: "_id", Value: userOID}},
					),
				)
			},
			wantUser: userHex,
		},
		{
			name:   "missing link returns empty user without error",
			linkID: "missing-link",
			mock: func(m *mt.T) {
				shareLinkNS := mt.TestDb + ".analytics_share_links"
				m.AddMockResponses(mt.CreateCursorResponse(0, shareLinkNS, mt.FirstBatch))
			},
			wantUser: "",
		},
		{
			name:   "link lookup error returns error",
			linkID: "link-error",
			mock: func(m *mt.T) {
				m.AddMockResponses(mt.CreateCommandErrorResponse(mt.CommandError{
					Message: "find failed",
					Code:    1,
				}))
			},
			expectErr: true,
		},
		{
			name:   "user missing returns empty user without error",
			linkID: "link-user-missing",
			mock: func(m *mt.T) {
				shareLinkNS := mt.TestDb + ".analytics_share_links"
				usersNS := mt.TestDb + ".users"
				m.AddMockResponses(
					mt.CreateCursorResponse(0, shareLinkNS, mt.FirstBatch,
						bson.D{{Key: "link_id", Value: "link-user-missing"}, {Key: "user_id", Value: userHex}},
					),
					mt.CreateCursorResponse(0, usersNS, mt.FirstBatch),
				)
			},
			wantUser: "",
		},
		{
			name:   "user lookup error returns error",
			linkID: "link-user-error",
			mock: func(m *mt.T) {
				shareLinkNS := mt.TestDb + ".analytics_share_links"
				m.AddMockResponses(
					mt.CreateCursorResponse(0, shareLinkNS, mt.FirstBatch,
						bson.D{{Key: "link_id", Value: "link-user-error"}, {Key: "user_id", Value: userHex}},
					),
					mt.CreateCommandErrorResponse(mt.CommandError{
						Message: "user find failed",
						Code:    2,
					}),
				)
			},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := mt.New(t, mt.NewOptions().ClientType(mt.Mock))
			h.Run(tc.name, func(m *mt.T) {
				repo := NewShareableLinkRepository(m.DB, testLogger())
				tc.mock(m)

				gotUser, err := repo.FindActiveUserIDByLinkID(context.Background(), tc.linkID)
				if tc.expectErr {
					if err == nil {
						t.Fatal("expected error, got nil")
					}
					return
				}
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if gotUser != tc.wantUser {
					t.Fatalf("expected user %q, got %q", tc.wantUser, gotUser)
				}
			})
		})
	}
}

func Test_normalizeUserID(t *testing.T) {
	oid := primitive.NewObjectID()

	cases := []struct {
		name string
		in   interface{}
		want string
	}{
		{name: "object id", in: oid, want: oid.Hex()},
		{name: "string id", in: "abc123", want: "abc123"},
		{name: "unsupported type", in: 123, want: ""},
		{name: "nil", in: nil, want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeUserID(tc.in)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
