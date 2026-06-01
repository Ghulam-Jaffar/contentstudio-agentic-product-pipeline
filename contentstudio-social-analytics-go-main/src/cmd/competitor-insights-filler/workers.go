package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	chRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

const (
	// BatchSize controls how many records to accumulate before inserting
	BatchSize = 500
)

// JobConfig holds the configuration for the gap fill job
type jobConfig struct {
	Platform     string
	Workers      int
	LookbackDays int
	From         time.Time
	To           time.Time
}

// CompetitorKey uniquely identifies a competitor (page_id or instagram_account_id)
type competitorKey struct {
	ID       string
	Platform string
}

// InsightRow represents a single insight record with date and metrics
// Includes all fields needed for copying to synthetic records
type insightRow struct {
	Date              time.Time
	Followers         int64
	Fans              int64    // Facebook only (total_fan_count)
	FollowingCount    int64    // Instagram only
	Biography         string   // Facebook
	ProfilePictureURL string   // Both platforms
	PageName          string   // Both platforms
	PageCategory      string   // Facebook only
	Emails            []string // Facebook only
	Birthday          string   // Facebook only
	WereHereCount     int64    // Facebook only
	CoverPhotoURL     string   // Facebook only
	Permalink         string   // Facebook only
}

// GapStats tracks statistics about gap filling
type gapStats struct {
	CompetitorsProcessed int32
	GapSegments          int32
	RecordsGenerated     int32
	RecordsInserted      int32
	Errors               int32
}

// CompetitorGapResult holds the result of gap detection for a single competitor
type competitorGapResult struct {
	Key              competitorKey
	ExistingRecords  int
	GapsDetected     int
	SyntheticRecords []interface{}
	Error            error
}

// produceCompetitors fetches distinct competitors from ClickHouse and sends them to the channel
func produceCompetitors(
	ctx context.Context,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	out chan<- competitorKey,
) error {

	op := log.
		Operation("produce_competitors").
		WithSentryTags(map[string]string{
			"platform": cfg.Platform,
		})

	defer func() {
		op.Complete(nil, "")
	}()

	var query string

	switch cfg.Platform {
	case PlatformFacebook:
		query = `
			SELECT DISTINCT page_id
			FROM contentstudiobackend.facebook_competitor_insights
			WHERE inserted_at >= ?
			  AND inserted_at <= ?
			ORDER BY page_id
		`

	case PlatformInstagram:
		query = `
			SELECT DISTINCT instagram_account_id
			FROM contentstudiobackend.instagram_competitor_insights
			WHERE inserted_at >= ?
			  AND inserted_at <= ?
			ORDER BY instagram_account_id
		`

	default:
		return fmt.Errorf("produceCompetitors: unsupported platform: %s", cfg.Platform)
	}

	log.Info().
		Str("platform", cfg.Platform).
		Time("from", cfg.From).
		Time("to", cfg.To).
		Msg("Querying for distinct competitors")

	rows, err := conn.Query(ctx, query, cfg.From, cfg.To)
	if err != nil {
		op.Complete(err, "query_failed")
		return fmt.Errorf("produceCompetitors: failed to query competitors: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var id string
		if err := rows.Scan(&id); err != nil {
			log.Error().Err(err).Msg("Failed to scan competitor ID")
			continue
		}

		if id == "" {
			continue
		}

		out <- competitorKey{
			ID:       id,
			Platform: cfg.Platform,
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("produceCompetitors: error iterating competitors: %w", err)
	}

	log.Info().
		Int("count", count).
		Str("platform", cfg.Platform).
		Msg("Finished producing competitors")

	return nil
}

// collectResults collects results from workers and forwards synthetic records to inserter
func collectResults(
	ctx context.Context,
	log *logger.Logger,
	in <-chan competitorGapResult,
	out chan<- interface{},
	stats *gapStats,
) {
	for result := range in {
		select {
		case <-ctx.Done():
			return
		default:
		}

		atomic.AddInt32(&stats.CompetitorsProcessed, 1)

		if result.Error != nil {
			atomic.AddInt32(&stats.Errors, 1)
			log.Error().
				Err(result.Error).
				Str("competitor_id", result.Key.ID).
				Str("platform", result.Key.Platform).
				Msg("Error processing competitor")
			continue
		}

		if result.GapsDetected > 0 {
			atomic.AddInt32(&stats.GapSegments, int32(result.GapsDetected))
			atomic.AddInt32(&stats.RecordsGenerated, int32(len(result.SyntheticRecords)))

			log.Info().
				Str("competitor_id", result.Key.ID).
				Int("existing_records", result.ExistingRecords).
				Int("gaps_detected", result.GapsDetected).
				Int("synthetic_records", len(result.SyntheticRecords)).
				Msg("Gaps detected for competitor")

			// Forward synthetic records to inserter
			for _, record := range result.SyntheticRecords {
				select {
				case <-ctx.Done():
					return
				case out <- record:
				}
			}
		}
	}
}

// runWorker processes competitors from the input channel
func runWorker(
	ctx context.Context,
	workerID int,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	in <-chan competitorKey,
	out chan<- competitorGapResult,
) {
	log.Debug().
		Int("worker_id", workerID).
		Msg("Worker started")

	processed := 0

	for {
		select {
		case <-ctx.Done():
			log.Debug().
				Int("worker_id", workerID).
				Int("processed", processed).
				Msg("Worker cancelled")
			return

		case key, ok := <-in:
			if !ok {
				log.Debug().
					Int("worker_id", workerID).
					Int("processed", processed).
					Msg("Worker finished")
				return
			}

			result := processCompetitor(ctx, cfg, conn, log, key)
			processed++

			select {
			case <-ctx.Done():
				return
			case out <- result:
			}
		}
	}
}

// processCompetitor handles gap detection and synthetic record generation for a single competitor
func processCompetitor(
	ctx context.Context,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	key competitorKey,
) competitorGapResult {
	result := competitorGapResult{
		Key: key,
	}

	op := log.
		Operation("process_competitor").
		WithSentryTags(map[string]string{
			"platform":      key.Platform,
			"competitor_id": key.ID,
		})

	defer func() {
		if result.Error != nil {
			op.Complete(result.Error, "")
		} else {
			op.Complete(nil, "")
		}
	}()

	// Fetch existing records
	records, err := fetchRecords(ctx, conn, cfg, key)
	if err != nil {
		result.Error = err
		return result
	}

	result.ExistingRecords = len(records)

	log.Debug().
		Str("competitor_id", key.ID).
		Str("platform", key.Platform).
		Int("records_count", len(records)).
		Msg("Fetched records for competitor")

	// Need at least 2 records to detect gaps
	if len(records) < 2 {
		log.Debug().
			Str("competitor_id", key.ID).
			Int("records_count", len(records)).
			Msg("Insufficient records for gap detection (need at least 2)")
		return result
	}

	// Log first and last record dates for debugging
	if len(records) > 0 {
		log.Debug().
			Str("competitor_id", key.ID).
			Time("first_date", records[0].Date).
			Time("last_date", records[len(records)-1].Date).
			Msg("Record date range")
	}

	// Generate synthetic records for gaps
	synthetic, gapsFound := generateMissingRecords(log, records, key)
	result.GapsDetected = gapsFound
	result.SyntheticRecords = synthetic

	if gapsFound > 0 {
		log.Info().
			Str("competitor_id", key.ID).
			Int("gaps_found", gapsFound).
			Int("synthetic_records", len(synthetic)).
			Msg("Gaps detected")
	}

	return result
}

// fetchRecords retrieves all insight records for a competitor, ordered by date
// Now includes all fields for copying to synthetic records
func fetchRecords(
	ctx context.Context,
	conn clickhouse.Conn,
	cfg jobConfig,
	key competitorKey,
) ([]insightRow, error) {
	var query string

	switch cfg.Platform {
	case PlatformFacebook:
		query = `
			SELECT 
				toDate(inserted_at) as date,
				followers_count,
				total_fan_count,
				biography,
				profile_picture_url,
				page_name,
				page_category,
				emails,
				birthday,
				were_here_count,
				cover_photo_url,
				permalink
			FROM contentstudiobackend.facebook_competitor_insights
			WHERE page_id = ?
			  AND inserted_at >= ?
			  AND inserted_at <= ?
			ORDER BY inserted_at ASC
			LIMIT 1 BY toDate(inserted_at)
		`

	case PlatformInstagram:
		query = `
			SELECT 
				toDate(inserted_at) as date,
				total_followed_by_count,
				total_following_count,
				profile_picture_url,
				page_name
			FROM contentstudiobackend.instagram_competitor_insights
			WHERE instagram_account_id = ?
			  AND inserted_at >= ?
			  AND inserted_at <= ?
			ORDER BY inserted_at ASC
			LIMIT 1 BY toDate(inserted_at)
		`

	default:
		return nil, fmt.Errorf("fetchRecords: unsupported platform: %s", cfg.Platform)
	}

	rows, err := conn.Query(ctx, query, key.ID, cfg.From, cfg.To)
	if err != nil {
		return nil, fmt.Errorf("fetchRecords: failed to query records for %s: %w", key.ID, err)
	}
	defer rows.Close()

	var records []insightRow
	seenDates := make(map[string]bool)

	for rows.Next() {
		var r insightRow
		if cfg.Platform == PlatformFacebook {
			if err := rows.Scan(
				&r.Date,
				&r.Followers,
				&r.Fans,
				&r.Biography,
				&r.ProfilePictureURL,
				&r.PageName,
				&r.PageCategory,
				&r.Emails,
				&r.Birthday,
				&r.WereHereCount,
				&r.CoverPhotoURL,
				&r.Permalink,
			); err != nil {
				return nil, fmt.Errorf("fetchRecords: failed to scan facebook record: %w", err)
			}
		} else {
			if err := rows.Scan(
				&r.Date,
				&r.Followers,
				&r.FollowingCount,
				&r.ProfilePictureURL,
				&r.PageName,
			); err != nil {
				return nil, fmt.Errorf("fetchRecords: failed to scan instagram record: %w", err)
			}
		}

		// Normalize to midnight UTC to ensure consistent date comparisons
		r.Date = time.Date(r.Date.Year(), r.Date.Month(), r.Date.Day(), 0, 0, 0, 0, time.UTC)

		// Deduplicate by date (take first occurrence)
		dateKey := r.Date.Format("2006-01-02")
		if !seenDates[dateKey] {
			seenDates[dateKey] = true
			records = append(records, r)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fetchRecords: error iterating records: %w", err)
	}

	return records, nil
}

// generateMissingRecords detects gaps and creates synthetic records using linear interpolation
// Returns the synthetic records and the number of gaps found
// Copies all metadata fields from the previous record
func generateMissingRecords(log *logger.Logger, records []insightRow, key competitorKey) ([]interface{}, int) {
	var synthetic []interface{}
	gapsFound := 0

	// Iterate through consecutive pairs of records
	for i := 0; i < len(records)-1; i++ {
		curr := records[i]
		next := records[i+1]

		// Calculate the number of missing days between curr and next
		daysDiff := int(next.Date.Sub(curr.Date).Hours() / 24)
		missingDays := daysDiff - 1

		// Log each pair being checked
		log.Debug().
			Str("competitor_id", key.ID).
			Time("curr_date", curr.Date).
			Time("next_date", next.Date).
			Int("days_diff", daysDiff).
			Int("missing_days", missingDays).
			Msg("Checking date pair for gaps")

		// No gap if records are consecutive days
		if missingDays <= 0 {
			continue
		}

		gapsFound++

		log.Info().
			Str("competitor_id", key.ID).
			Time("gap_start", curr.Date).
			Time("gap_end", next.Date).
			Int("missing_days", missingDays).
			Msg("Gap detected")

		// Calculate interpolation steps
		// Formula: step = (end - start) / (gaps + 1)
		deltaFollowers := next.Followers - curr.Followers
		deltaFans := next.Fans - curr.Fans
		deltaFollowing := next.FollowingCount - curr.FollowingCount

		stepFollowers := float64(deltaFollowers) / float64(missingDays+1)
		stepFans := float64(deltaFans) / float64(missingDays+1)
		stepFollowing := float64(deltaFollowing) / float64(missingDays+1)

		// Generate synthetic records for each missing day
		for d := 1; d <= missingDays; d++ {
			syntheticDate := curr.Date.AddDate(0, 0, d)

			// Linear interpolation: value = start + (step * day_offset)
			syntheticFollowers := curr.Followers + int64(stepFollowers*float64(d))
			syntheticFans := curr.Fans + int64(stepFans*float64(d))
			syntheticFollowing := curr.FollowingCount + int64(stepFollowing*float64(d))

			// Ensure non-negative values
			if syntheticFollowers < 0 {
				syntheticFollowers = 0
			}
			if syntheticFans < 0 {
				syntheticFans = 0
			}
			if syntheticFollowing < 0 {
				syntheticFollowing = 0
			}

			// Create metadata map for synthetic records
			metadata := map[string]string{
				"message": "inserted synthetic data",
			}

			// Create platform-specific model
			switch key.Platform {
			case PlatformFacebook:
				synthetic = append(synthetic, &models.FacebookCompetitorInsights{
					RecordID:          generateRecordID(key.ID, syntheticDate),
					PageID:            key.ID,
					FollowersCount:    syntheticFollowers,
					TotalFanCount:     syntheticFans,
					TalkingAboutThis:  0, // Keep as 0 for synthetic
					Biography:         curr.Biography,
					ProfilePictureURL: curr.ProfilePictureURL,
					PageName:          curr.PageName,
					PageCategory:      curr.PageCategory,
					Emails:            curr.Emails,
					Birthday:          curr.Birthday,
					WereHereCount:     curr.WereHereCount,
					CoverPhotoURL:     curr.CoverPhotoURL,
					Permalink:         curr.Permalink,
					Metadata:          metadata,
					InsertedAt:        syntheticDate.Truncate(time.Hour),
				})

			case PlatformInstagram:
				synthetic = append(synthetic, &models.InstagramCompetitorInsights{
					RecordID:             generateRecordID(key.ID, syntheticDate),
					InstagramAccountID:   key.ID,
					TotalFollowedByCount: syntheticFollowers,
					TotalFollowingCount:  syntheticFollowing,
					ProfilePictureURL:    curr.ProfilePictureURL,
					PageName:             curr.PageName,
					Metadata:             metadata,
					InsertedAt:           syntheticDate.Truncate(time.Hour),
				})
			}
		}
	}

	return synthetic, gapsFound
}

// runInserter batches and inserts synthetic records into ClickHouse
func runInserter(
	ctx context.Context,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	in <-chan interface{},
	stats *gapStats,
) {
	switch cfg.Platform {
	case PlatformFacebook:
		insertFacebookRecords(ctx, cfg, conn, log, in, stats)
	case PlatformInstagram:
		insertInstagramRecords(ctx, cfg, conn, log, in, stats)
	default:
		log.Error().Str("platform", cfg.Platform).Msg("Unknown platform in inserter")
	}
}

// insertFacebookRecords handles Facebook-specific batch inserts
func insertFacebookRecords(
	ctx context.Context,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	in <-chan interface{},
	stats *gapStats,
) {
	// Create a wrapper client to use the Client methods
	client := &chRepo.Client{
		Conn:   conn,
		Logger: log.Logger,
	}
	batch := make([]*models.FacebookCompetitorInsights, 0, BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := client.InsertCompetitorInsights(ctx, batch); err != nil {
			op := log.
				Operation("insert_facebook_batch").
				WithSentryTags(map[string]string{
					"platform": "facebook",
				}).
				WithSentryExtras(map[string]interface{}{
					"batch_size": len(batch),
				})

			op.Complete(err, "")

			log.Error().
				Err(err).
				Int("batch_size", len(batch)).
				Msg("Failed to insert Facebook batch")

			atomic.AddInt32(&stats.Errors, 1)
		} else {
			atomic.AddInt32(&stats.RecordsInserted, int32(len(batch)))
			log.Debug().
				Int("batch_size", len(batch)).
				Msg("Inserted Facebook batch")
		}

		batch = batch[:0]
	}

	for record := range in {
		select {
		case <-ctx.Done():
			flush()
			return
		default:
		}

		fbRecord, ok := record.(*models.FacebookCompetitorInsights)
		if !ok {
			log.Error().Msg("Invalid record type for Facebook")
			continue
		}

		batch = append(batch, fbRecord)

		if len(batch) >= BatchSize {
			flush()
		}
	}

	// Flush remaining records
	flush()
	log.Info().Msg("Facebook inserter completed")
}

// insertInstagramRecords handles Instagram-specific batch inserts
func insertInstagramRecords(
	ctx context.Context,
	cfg jobConfig,
	conn clickhouse.Conn,
	log *logger.Logger,
	in <-chan interface{},
	stats *gapStats,
) {
	// Create a wrapper client to use the Client methods
	client := &chRepo.Client{
		Conn:   conn,
		Logger: log.Logger,
	}
	batch := make([]*models.InstagramCompetitorInsights, 0, BatchSize)

	flush := func() {
		if len(batch) == 0 {
			return
		}

		if err := client.InsertInstagramCompetitorInsights(ctx, batch); err != nil {
			op := log.
				Operation("insert_instagram_batch").
				WithSentryTags(map[string]string{
					"platform": "instagram",
				}).
				WithSentryExtras(map[string]interface{}{
					"batch_size": len(batch),
				})

			op.Complete(err, "")

			log.Error().
				Err(err).
				Int("batch_size", len(batch)).
				Msg("Failed to insert Instagram batch")

			atomic.AddInt32(&stats.Errors, 1)
		} else {
			atomic.AddInt32(&stats.RecordsInserted, int32(len(batch)))
			log.Debug().
				Int("batch_size", len(batch)).
				Msg("Inserted Instagram batch")
		}

		batch = batch[:0]
	}

	for record := range in {
		select {
		case <-ctx.Done():
			flush()
			return
		default:
		}

		igRecord, ok := record.(*models.InstagramCompetitorInsights)
		if !ok {
			log.Error().Msg("Invalid record type for Instagram")
			continue
		}

		batch = append(batch, igRecord)

		if len(batch) >= BatchSize {
			flush()
		}
	}

	// Flush remaining records
	flush()
	log.Info().Msg("Instagram inserter completed")
}

func generateRecordID(pageID string, timestamp time.Time) string {
	date := timestamp.UTC().Format("2006-01-02")
	str := fmt.Sprintf("%s_%s", pageID, date)
	hash := md5.Sum([]byte(str))
	return hex.EncodeToString(hash[:])
}
