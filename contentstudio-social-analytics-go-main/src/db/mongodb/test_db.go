package mongodb

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"sync"
	"time"

	mongo3 "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedRepo is an in-memory UnifiedSocialRepository for tests.
// It supports seeding data and exercising real read/write paths without a DB.
type SeedRepo struct {
	mu          sync.RWMutex
	byID        map[primitive.ObjectID]*mongo3.SocialIntegration
	byPlatform  map[string]map[string]*mongo3.SocialIntegration // platformType -> platformIdentifier -> *SI
	byWorkspace map[primitive.ObjectID][]*mongo3.SocialIntegration
	jobMetadata []TwitterJobMetadataPayload
}

func NewSeedRepo() *SeedRepo {
	return &SeedRepo{
		byID:        make(map[primitive.ObjectID]*mongo3.SocialIntegration),
		byPlatform:  make(map[string]map[string]*mongo3.SocialIntegration),
		byWorkspace: make(map[primitive.ObjectID][]*mongo3.SocialIntegration),
	}
}

// SeedAccounts inserts/updates one or more accounts into the repo.
func (r *SeedRepo) SeedAccounts(accs ...mongo3.SocialIntegration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	for i := range accs {
		a := accs[i] // copy
		if a.ID.IsZero() {
			a.ID, _ = primitive.ObjectIDFromHex("65119038a5e02c3fc1c45c9d")
		}
		if a.CreatedAt == nil {
			a.CreatedAt = &mongo3.MongoTime{Time: now}
		}
		if a.UpdatedAt == nil {
			a.UpdatedAt = &mongo3.MongoTime{Time: now}
		}
		// Keep platform_identifier populated
		if a.PlatformIdentifier == "" {
			a.PlatformIdentifier = a.GetPlatformID()
		}

		if a.AccessToken != "" {
			key := "01234567890123456789012345678901"
			base64Key := base64.StdEncoding.EncodeToString([]byte(key))

			token := "token123"
			accessToken, _ := EncryptToken(token, base64Key)
			a.AccessToken = accessToken
		}

		// store by id
		r.byID[a.ID] = &a

		// store by platform
		if _, ok := r.byPlatform[a.PlatformType]; !ok {
			r.byPlatform[a.PlatformType] = make(map[string]*mongo3.SocialIntegration)
		}
		r.byPlatform[a.PlatformType][a.PlatformIdentifier] = &a

		// store by workspace
		r.byWorkspace[a.WorkspaceID] = append(r.byWorkspace[a.WorkspaceID], &a)
	}
}

// Clear removes all seeded data.
func (r *SeedRepo) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.byID = make(map[primitive.ObjectID]*mongo3.SocialIntegration)
	r.byPlatform = make(map[string]map[string]*mongo3.SocialIntegration)
	r.byWorkspace = make(map[primitive.ObjectID][]*mongo3.SocialIntegration)
	r.jobMetadata = nil
}

// ---- UnifiedSocialRepository methods (in-memory) ----

func (r *SeedRepo) FindByID(ctx context.Context, id primitive.ObjectID) (*mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if a, ok := r.byID[id]; ok {
		cp := *a
		return &cp, nil
	}
	return nil, nil
}

func (r *SeedRepo) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if m, ok := r.byPlatform[platformType]; ok {
		if a, ok := m[platformID]; ok {
			cp := *a
			return &cp, nil
		}
	}
	// simple legacy fallbacks using the same identifier (you can customize if needed)
	if m, ok := r.byPlatform[platformType]; ok {
		if a, ok := m[platformID]; ok {
			cp := *a
			return &cp, nil
		}
	}
	return nil, nil
}

func (r *SeedRepo) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var out []mongo3.SocialIntegration

	allowTypes := make(map[string]bool)
	for _, t := range accountTypes {
		allowTypes[t] = true
	}

	for _, a := range r.byID {
		if a.PlatformType != platformType {
			continue
		}
		if len(accountTypes) > 0 && !allowTypes[a.Type] {
			continue
		}
		if a.Validity != mongo3.ValidityValid {
			continue
		}
		switch a.State {
		case mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed:
			cp := *a
			out = append(out, cp)
		}
	}
	return out, nil
}

func (r *SeedRepo) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allow := map[string]bool{}
	for _, p := range platforms {
		allow[p] = true
	}

	var out []mongo3.SocialIntegration
	for _, a := range r.byWorkspace[workspaceID] {
		if len(platforms) > 0 && !allow[a.PlatformType] {
			continue
		}
		if a.State == mongo3.StateDeleted {
			continue
		}
		cp := *a
		out = append(out, cp)
	}
	return out, nil
}

func (r *SeedRepo) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	// We allow tests to stash last_*_updated_at inside ExtraData using the same field key.
	var out []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != platformType {
			continue
		}
		if a.Validity != mongo3.ValidityValid || a.State != mongo3.StateAdded {
			continue
		}
		var ts time.Time
		if a.ExtraData != nil {
			if v, ok := a.ExtraData[lastUpdateField]; ok {
				switch t := v.(type) {
				case time.Time:
					ts = t
				case *time.Time:
					if t != nil {
						ts = *t
					}
				case mongo3.MongoTime:
					ts = t.Time
				case *mongo3.MongoTime:
					if t != nil {
						ts = t.Time
					}
				}
			}
		}
		// include if missing or older than cutoff
		if ts.IsZero() || ts.Before(cutoff) {
			cp := *a
			out = append(out, cp)
		}
	}
	// Sort oldest first if you need stable ordering; omitted for brevity.
	return out, nil
}

// GetAccountsNeedingUpdatePaginated returns paginated accounts needing update
func (r *SeedRepo) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	allowTypes := make(map[string]bool)
	for _, t := range accountTypes {
		allowTypes[t] = true
	}

	var all []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != platformType {
			continue
		}
		if len(accountTypes) > 0 && !allowTypes[a.Type] {
			continue
		}
		if a.Validity != mongo3.ValidityValid || a.State != mongo3.StateAdded {
			continue
		}

		var ts time.Time
		if a.ExtraData != nil {
			if v, ok := a.ExtraData["last_analytics_updated_at"]; ok {
				switch t := v.(type) {
				case time.Time:
					ts = t
				case *time.Time:
					if t != nil {
						ts = *t
					}
				case mongo3.MongoTime:
					ts = t.Time
				case *mongo3.MongoTime:
					if t != nil {
						ts = t.Time
					}
				}
			}
		}

		if ts.IsZero() || ts.Before(cutoff) {
			cp := *a
			all = append(all, cp)
		}
	}

	// Apply pagination
	if skip >= int64(len(all)) {
		return []mongo3.SocialIntegration{}, nil
	}

	end := skip + limit
	if end > int64(len(all)) {
		end = int64(len(all))
	}

	return all[skip:end], nil
}

// CountAccountsNeedingUpdate returns count of accounts needing update
func (r *SeedRepo) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	accounts, err := r.GetAccountsNeedingUpdatePaginated(ctx, platformType, accountTypes, hours, 0, 999999)
	if err != nil {
		return 0, err
	}
	return int64(len(accounts)), nil
}

// GetAccountsNeedingUpdateByID returns accounts needing update using ID-based pagination
func (r *SeedRepo) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allowTypes := make(map[string]bool)
	for _, t := range accountTypes {
		allowTypes[t] = true
	}

	// Collect all matching accounts sorted by ID
	var all []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != platformType {
			continue
		}
		if len(accountTypes) > 0 && !allowTypes[a.Type] {
			continue
		}
		if a.Validity != mongo3.ValidityValid {
			continue
		}
		switch a.State {
		case mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed:
			// Filter by lastID (only include accounts with ID > lastID)
			if lastID != primitive.NilObjectID && a.ID.Hex() <= lastID.Hex() {
				continue
			}
			cp := *a
			all = append(all, cp)
		}
	}

	// Sort by ID for consistent pagination
	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[i].ID.Hex() > all[j].ID.Hex() {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	// Apply limit
	if int64(len(all)) > limit {
		all = all[:limit]
	}

	return all, nil
}

func (r *SeedRepo) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	allowTypes := make(map[string]bool)
	for _, t := range accountTypes {
		allowTypes[t] = true
	}

	var all []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != platformType {
			continue
		}
		if len(accountTypes) > 0 && !allowTypes[a.Type] {
			continue
		}
		if a.Validity != mongo3.ValidityValid {
			continue
		}
		switch a.State {
		case mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed:
		default:
			continue
		}
		if lastID != primitive.NilObjectID && a.ID.Hex() <= lastID.Hex() {
			continue
		}
		cp := *a
		all = append(all, cp)
	}

	for i := 0; i < len(all)-1; i++ {
		for j := i + 1; j < len(all); j++ {
			if all[i].ID.Hex() > all[j].ID.Hex() {
				all[i], all[j] = all[j], all[i]
			}
		}
	}

	if int64(len(all)) > limit {
		all = all[:limit]
	}
	return all, nil
}

func (r *SeedRepo) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	accounts, err := r.GetValidAccounts(ctx, platformType, accountTypes)
	if err != nil {
		return 0, err
	}
	return int64(len(accounts)), nil
}

// GetYouTubeAccountsNeedingUpdatePaginated returns paginated YouTube accounts needing update with consent time filter
func (r *SeedRepo) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	consentCutoff := time.Now().UTC().AddDate(0, 0, -consentDays)

	var all []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != mongo3.PlatformYouTube {
			continue
		}
		if a.Validity != mongo3.ValidityValid || a.State != mongo3.StateAdded {
			continue
		}

		// Check consent time
		if a.Preferences == nil {
			continue
		}
		consentTimeVal, exists := a.Preferences["last_youtube_consent_time"]
		if !exists {
			continue
		}

		var consentTime time.Time
		switch v := consentTimeVal.(type) {
		case string:
			parsed, err := time.Parse(time.RFC3339, v)
			if err != nil {
				parsed, err = time.Parse("2006-01-02T15:04:05.000Z", v)
				if err != nil {
					continue
				}
			}
			consentTime = parsed
		case time.Time:
			consentTime = v
		default:
			continue
		}

		if consentTime.Before(consentCutoff) {
			continue
		}

		// Check last update time
		var ts time.Time
		if a.ExtraData != nil {
			if v, ok := a.ExtraData["last_analytics_updated_at"]; ok {
				switch t := v.(type) {
				case time.Time:
					ts = t
				case *time.Time:
					if t != nil {
						ts = *t
					}
				case mongo3.MongoTime:
					ts = t.Time
				case *mongo3.MongoTime:
					if t != nil {
						ts = t.Time
					}
				}
			}
		}

		if ts.IsZero() || ts.Before(cutoff) {
			cp := *a
			all = append(all, cp)
		}
	}

	// Apply pagination
	if skip >= int64(len(all)) {
		return []mongo3.SocialIntegration{}, nil
	}

	end := skip + limit
	if end > int64(len(all)) {
		end = int64(len(all))
	}

	return all[skip:end], nil
}

// GetYouTubeAccountsNeedingUpdateByID returns YouTube accounts using ID-based pagination (cursor-free)
func (r *SeedRepo) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	consentCutoff := time.Now().UTC().AddDate(0, 0, -consentDays)

	var all []mongo3.SocialIntegration
	for _, a := range r.byID {
		if a.PlatformType != mongo3.PlatformYouTube {
			continue
		}
		if a.Validity != mongo3.ValidityValid || a.State != mongo3.StateAdded {
			continue
		}

		// ID-based pagination: skip accounts with ID <= lastID
		if lastID != primitive.NilObjectID && a.ID.Hex() <= lastID.Hex() {
			continue
		}

		// Check consent time
		if a.Preferences == nil {
			continue
		}
		consentTimeVal, exists := a.Preferences["last_youtube_consent_time"]
		if !exists {
			continue
		}

		var consentTime time.Time
		switch v := consentTimeVal.(type) {
		case string:
			parsed, err := time.Parse(time.RFC3339, v)
			if err != nil {
				parsed, err = time.Parse("2006-01-02T15:04:05.000Z", v)
				if err != nil {
					continue
				}
			}
			consentTime = parsed
		case time.Time:
			consentTime = v
		default:
			continue
		}

		if consentTime.Before(consentCutoff) {
			continue
		}

		// Check last update time
		var ts time.Time
		if a.ExtraData != nil {
			if v, ok := a.ExtraData["last_analytics_updated_at"]; ok {
				switch t := v.(type) {
				case time.Time:
					ts = t
				case *time.Time:
					if t != nil {
						ts = *t
					}
				case mongo3.MongoTime:
					ts = t.Time
				case *mongo3.MongoTime:
					if t != nil {
						ts = t.Time
					}
				}
			}
		}

		if ts.IsZero() || ts.Before(cutoff) {
			cp := *a
			all = append(all, cp)
		}
	}

	// Apply limit
	if int64(len(all)) > limit {
		all = all[:limit]
	}

	return all, nil
}

// CountYouTubeAccountsNeedingUpdate returns count of YouTube accounts needing update with consent time filter
func (r *SeedRepo) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	accounts, err := r.GetYouTubeAccountsNeedingUpdatePaginated(ctx, hours, consentDays, 0, 999999)
	if err != nil {
		return 0, err
	}
	return int64(len(accounts)), nil
}

func (r *SeedRepo) Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	a, ok := r.byID[id]
	if !ok {
		return errors.New("not found") // you can replace with mongo.ErrNoDocuments if preferred in tests
	}
	// Apply known fields
	for k, v := range updates {
		switch k {
		case "state":
			if s, _ := v.(string); s != "" {
				a.State = s
			}
		case "validity":
			if s, _ := v.(string); s != "" {
				a.Validity = s
			}
		case "access_token":
			if s, _ := v.(string); s != "" {
				a.AccessToken = s
			}
		case "refresh_token", "long_access_token", "oauth_token", "oauth_token_secret":
			if a.ExtraData == nil {
				a.ExtraData = map[string]interface{}{}
			}
			a.ExtraData[k] = v
		case "updated_at":
			// ignore; UpdatedAt maintained below
		default:
			// allow arbitrary fields to be stashed into ExtraData for flexibility
			if a.ExtraData == nil {
				a.ExtraData = map[string]interface{}{}
			}
			a.ExtraData[k] = v
		}
	}
	now := time.Now().UTC()
	a.UpdatedAt = &mongo3.MongoTime{Time: now}
	return nil
}

func (r *SeedRepo) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error {
	fieldMap := map[string]string{
		"analytics": "last_analytics_updated_at",
		"insights":              "last_insights_analytics_updated_at",
		"fans":                  "last_fans_analytics_updated_at",
		"video":                 "last_video_analytics_updated_at",
		"group":                 "last_group_analytics_updated_at",
		"link_preview":          "last_link_preview_updated_at",
	}
	field, ok := fieldMap[timestampType]
	if !ok {
		return errors.New("invalid timestamp type")
	}
	return r.Update(ctx, id, bson.M{field: mongo3.MongoTime{Time: timestamp}})
}

func (r *SeedRepo) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	if len(tokens) == 0 {
		return errors.New("no valid token fields to update")
	}
	u := bson.M{}
	for k, v := range tokens {
		switch k {
		case "access_token", "refresh_token", "long_access_token", "oauth_token", "oauth_token_secret":
			u[k] = v
		case "expires_at":
			// ignore parsing in seed
		default:
			// ignore unknown keys
		}
	}
	if len(u) == 0 {
		return errors.New("no valid token fields to update")
	}
	return r.Update(ctx, id, u)
}

func (r *SeedRepo) UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error {
	return r.Update(ctx, id, bson.M{"state": newState})
}

func (r *SeedRepo) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	return r.Update(ctx, id, bson.M{"validity": newValidity})
}

func (r *SeedRepo) RecordProcessingError(_ context.Context, _ primitive.ObjectID, _ string) error {
	return nil
}

func (r *SeedRepo) ClearProcessingError(_ context.Context, _ primitive.ObjectID) error {
	return nil
}

func (r *SeedRepo) Create(ctx context.Context, account *mongo3.SocialIntegration) (primitive.ObjectID, error) {
	if account == nil {
		return primitive.NilObjectID, errors.New("nil account")
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	if account.ID.IsZero() {
		account.ID = primitive.NewObjectID()
	}
	if account.CreatedAt == nil {
		account.CreatedAt = &mongo3.MongoTime{Time: now}
	}
	if account.UpdatedAt == nil {
		account.UpdatedAt = &mongo3.MongoTime{Time: now}
	}
	if account.PlatformIdentifier == "" {
		account.PlatformIdentifier = account.GetPlatformID()
	}

	id := account.ID
	cp := *account // store a copy
	r.byID[id] = &cp

	if _, ok := r.byPlatform[cp.PlatformType]; !ok {
		r.byPlatform[cp.PlatformType] = make(map[string]*mongo3.SocialIntegration)
	}
	r.byPlatform[cp.PlatformType][cp.PlatformIdentifier] = &cp
	r.byWorkspace[cp.WorkspaceID] = append(r.byWorkspace[cp.WorkspaceID], &cp)
	return id, nil
}

func (r *SeedRepo) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.UpdateState(ctx, id, mongo3.StateDeleted)
}

func (r *SeedRepo) InsertTwitterJobMetadata(ctx context.Context, payload TwitterJobMetadataPayload) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobMetadata = append(r.jobMetadata, payload)
	return nil
}

func (r *SeedRepo) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongo3.SocialIntegration, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	lookup := make(map[string]bool, len(platformIDs))
	for _, id := range platformIDs {
		lookup[id] = true
	}
	var out []mongo3.SocialIntegration
	if m, ok := r.byPlatform[platformType]; ok {
		for pid, a := range m {
			if lookup[pid] {
				cp := *a
				out = append(out, cp)
			}
		}
	}
	return out, nil
}

// EncryptedPayload is the JSON format expected by your decryption function
type EncryptedPayload struct {
	IV    string `json:"iv"`
	Value string `json:"value"` // Must match expected field name
}

// EncryptToken encrypts a plaintext token using AES-256-CBC and returns a base64-encoded JSON payload string.
func EncryptToken(plaintextToken string, base64EncodedKey string) (string, error) {
	// Decode base64 key
	key, err := base64.StdEncoding.DecodeString(base64EncodedKey)
	if err != nil {
		return "", err
	}
	if len(key) != 32 {
		return "", errors.New("key must be 32 bytes (for AES-256)")
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	// Generate random IV
	iv := make([]byte, aes.BlockSize) // AES block size = 16
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	// Pad the plaintext to be a multiple of block size
	paddedToken := pkcs7Pad([]byte(plaintextToken), aes.BlockSize)

	// Encrypt
	cipherText := make([]byte, len(paddedToken))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, paddedToken)

	// Build payload
	payload := EncryptedPayload{
		IV:    base64.StdEncoding.EncodeToString(iv),
		Value: base64.StdEncoding.EncodeToString(cipherText),
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Return the base64-encoded JSON payload
	return base64.StdEncoding.EncodeToString(jsonPayload), nil
}

// pkcs7Pad adds padding according to PKCS#7.
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - (len(data) % blockSize)
	padBytes := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padBytes...)
}
