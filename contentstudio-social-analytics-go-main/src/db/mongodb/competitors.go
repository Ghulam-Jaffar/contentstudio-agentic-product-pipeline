package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// CompetitorRepository handles MongoDB operations for competitors.
type CompetitorRepository struct {
	collection        *mongo.Collection
	reportsCollection *mongo.Collection
	log               *logger.Logger
}

// NewCompetitorRepository initializes a new competitor repository.
func NewCompetitorRepository(db *mongo.Database, log *logger.Logger) *CompetitorRepository {
	return &CompetitorRepository{
		collection:        db.Collection("competitors"),
		reportsCollection: db.Collection("competitors_reports"),
		log:               log,
	}
}

// GetByCompetitorID retrieves competitors by their competitor_id.
func (r *CompetitorRepository) GetByCompetitorID(ctx context.Context, competitorID string) ([]*mongoModels.Competitor, error) {
	filter := bson.M{"competitor_id": competitorID}
	projection := bson.M{"_id": 1, "state": 1, "image": 1, "error": 1}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetByCompetitorID: failed to find competitors: %w", err)
	}
	defer cursor.Close(ctx)

	var competitors []*mongoModels.Competitor
	if err := cursor.All(ctx, &competitors); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetByCompetitorID: failed to decode competitors: %w", err)
	}

	return competitors, nil
}

// GetByID retrieves a competitor by Mongo ObjectID.
func (r *CompetitorRepository) GetByID(ctx context.Context, id string) (*mongoModels.Competitor, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetByID: invalid object ID: %w", err)
	}

	var competitor mongoModels.Competitor
	if err := r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&competitor); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetByID: failed to find competitor: %w", err)
	}

	return &competitor, nil
}

// GetAccounts retrieves all active competitors for a platform type linked to
// workspaces whose super-admin state is eligible for processing.
func (r *CompetitorRepository) GetAccounts(ctx context.Context, platformType string) ([]*mongoModels.Competitor, error) {
	r.log.Info().Msgf("[GetAccounts] Fetching accounts for platform: %s", platformType)

	validIDs, err := r.getValidCompetitorIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetAccounts: failed to get valid competitor IDs: %w", err)
	}

	filter := bson.M{
		"_id":          bson.M{"$in": validIDs},
		"is_active":    true,
		"platform_type": platformType,
		"state":        bson.M{"$ne": mongoModels.StateFailed},
	}
	projection := bson.M{"competitor_id": 1, "name": 1, "slug": 1}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetAccounts: failed to find filtered accounts: %w", err)
	}
	defer cursor.Close(ctx)

	var competitors []*mongoModels.Competitor
	if err := cursor.All(ctx, &competitors); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetAccounts: failed to decode filtered accounts: %w", err)
	}
	r.log.Info().Msgf("[GetAccounts] Accounts linked to workspace reports: %d", len(competitors))

	return competitors, nil
}

// GetActiveAccounts retrieves all active competitors for a platform type without workspace linkage filtering.
func (r *CompetitorRepository) GetActiveAccounts(ctx context.Context, platformType string) ([]*mongoModels.Competitor, error) {
	r.log.Info().Msgf("[GetActiveAccounts] Fetching active accounts for platform: %s", platformType)

	filter := bson.M{"is_active": true, "platform_type": platformType}
	projection := bson.M{"competitor_id": 1, "name": 1, "slug": 1}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetActiveAccounts: failed to find active accounts: %w", err)
	}
	defer cursor.Close(ctx)

	var competitors []*mongoModels.Competitor
	if err := cursor.All(ctx, &competitors); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetActiveAccounts: failed to decode active accounts: %w", err)
	}
	r.log.Info().Msgf("[GetActiveAccounts] Active accounts found: %d", len(competitors))

	return competitors, nil
}

// getValidCompetitorIDs fetches competitors linked to workspaces via reports
// whose super-admin state allows processing.
func (r *CompetitorRepository) getValidCompetitorIDs(ctx context.Context) ([]primitive.ObjectID, error) {
	pipeline := mongo.Pipeline{
		// Convert workspace_id string to ObjectID
		{{Key: "$addFields", Value: bson.D{{Key: "workspace_id_object", Value: bson.D{{Key: "$toObjectId", Value: "$workspace_id"}}}}}},
		// Join with workspace collection
		{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "workspace"}, {Key: "localField", Value: "workspace_id_object"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "workspace_info"}}}},
		{{Key: "$unwind", Value: "$workspace_info"}},
		{{Key: "$match", Value: bson.D{{Key: "workspace_info.super_admin_state", Value: bson.D{{Key: "$in", Value: bson.A{mongoModels.SuperAdminStateActive, mongoModels.SuperAdminStatePastDue}}}}}}},
		{{Key: "$unwind", Value: "$competitors"}},
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$toObjectId", Value: "$competitors"}}}}}},
	}

	cursor, err := r.reportsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.getValidCompetitorIDs: failed to aggregate valid competitor IDs: %w", err)
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.getValidCompetitorIDs: failed to decode valid competitor IDs: %w", err)
	}

	var validIDs []primitive.ObjectID
	for _, doc := range results {
		if id, ok := doc["_id"].(primitive.ObjectID); ok {
			validIDs = append(validIDs, id)
		}
	}

	return validIDs, nil
}

// UpdateState updates the 'state' field of a competitor.
func (r *CompetitorRepository) UpdateState(ctx context.Context, id primitive.ObjectID, state string) error {
	update := bson.M{"$set": bson.M{"state": state}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("CompetitorRepository.UpdateState: failed to update state: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CompetitorRepository.UpdateState: competitor not found")
	}
	return nil
}

// UpdateField updates the 'last_analytics_updated_at' timestamp.
func (r *CompetitorRepository) UpdateField(ctx context.Context, id primitive.ObjectID, timestamp time.Time) error {
	update := bson.M{"$set": bson.M{"last_analytics_updated_at": timestamp}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("CompetitorRepository.UpdateField: failed to update field: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CompetitorRepository.UpdateField: competitor not found")
	}
	return nil
}

// AddError sets the 'error' field for a competitor.
func (r *CompetitorRepository) AddError(ctx context.Context, id primitive.ObjectID, errorMsg string) error {
	update := bson.M{"$set": bson.M{"error": errorMsg}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("CompetitorRepository.AddError: failed to add error: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CompetitorRepository.AddError: competitor not found")
	}
	return nil
}

// UpdateImage sets the 'image' field for a competitor.
func (r *CompetitorRepository) UpdateImage(ctx context.Context, id primitive.ObjectID, imageURL string) error {
	update := bson.M{"$set": bson.M{"image": imageURL}}
	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return fmt.Errorf("CompetitorRepository.UpdateImage: failed to update image: %w", err)
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("CompetitorRepository.UpdateImage: competitor not found")
	}
	return nil
}

// GetReportsByCompetitorID retrieves reports that include a specific competitor.
func (r *CompetitorRepository) GetReportsByCompetitorID(ctx context.Context, competitorID string) ([]*mongoModels.CompetitorReport, error) {
	filter := bson.M{"competitors": bson.M{"$elemMatch": bson.M{"$in": []string{competitorID}}}}
	cursor, err := r.reportsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportsByCompetitorID: failed to find reports: %w", err)
	}
	defer cursor.Close(ctx)

	var reports []*mongoModels.CompetitorReport
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportsByCompetitorID: failed to decode reports: %w", err)
	}

	return reports, nil
}

// GetReportByID retrieves a single report by its MongoDB ObjectID.
func (r *CompetitorRepository) GetReportByID(ctx context.Context, reportID string) (*mongoModels.CompetitorReport, error) {
	objectID, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportByID: invalid report ID: %w", err)
	}

	var report mongoModels.CompetitorReport
	if err := r.reportsCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&report); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CompetitorRepository.GetReportByID: report not found")
		}
		return nil, fmt.Errorf("CompetitorRepository.GetReportByID: failed to find report: %w", err)
	}

	return &report, nil
}

// GetUserByID retrieves a user by their MongoDB ObjectID.
func (r *CompetitorRepository) GetUserByID(ctx context.Context, userID string) (*mongoModels.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetUserByID: invalid user ID: %w", err)
	}

	usersCollection := r.collection.Database().Collection("users")
	var user mongoModels.User
	if err := usersCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("CompetitorRepository.GetUserByID: user not found")
		}
		return nil, fmt.Errorf("CompetitorRepository.GetUserByID: failed to find user: %w", err)
	}

	return &user, nil
}

// GetLastAnalyticsTime returns the last analytics update timestamp for a competitor.
func (r *CompetitorRepository) GetLastAnalyticsTime(ctx context.Context, competitorID string) (time.Time, error) {
	var comp struct {
		LastAnalyticsUpdatedAt time.Time `bson:"last_analytics_updated_at"`
	}
	err := r.collection.FindOne(ctx, bson.M{"competitor_id": competitorID}).Decode(&comp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return time.Time{}, nil // first run
		}
		return time.Time{}, fmt.Errorf("CompetitorRepository.GetLastAnalyticsTime: failed to get last analytics time: %w", err)
	}
	return comp.LastAnalyticsUpdatedAt, nil
}

// GetReportCompetitors fetches a report by ID, then resolves all competitor documents
// and returns a map of platform competitor_id → competitor metadata.
// This mirrors PHP's getReportCompetitors() used by all competitor analytics endpoints.
func (r *CompetitorRepository) GetReportCompetitors(ctx context.Context, reportID string) (map[string]*mongoModels.Competitor, error) {
	report, err := r.GetReportByID(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportCompetitors: %w", err)
	}

	if len(report.Competitors) == 0 {
		return map[string]*mongoModels.Competitor{}, nil
	}

	// Convert string IDs to ObjectIDs
	objectIDs := make([]primitive.ObjectID, 0, len(report.Competitors))
	for _, idStr := range report.Competitors {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			continue
		}
		objectIDs = append(objectIDs, oid)
	}

	filter := bson.M{"_id": bson.M{"$in": objectIDs}}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportCompetitors: failed to find competitors: %w", err)
	}
	defer cursor.Close(ctx)

	var competitors []*mongoModels.Competitor
	if err := cursor.All(ctx, &competitors); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportCompetitors: failed to decode competitors: %w", err)
	}

	result := make(map[string]*mongoModels.Competitor, len(competitors))
	for _, comp := range competitors {
		result[comp.GetCompetitorIDAsString()] = comp
	}
	return result, nil
}

// AddUpdateReport creates or updates a competitor report in MongoDB.
func (r *CompetitorRepository) AddUpdateReport(ctx context.Context, report *mongoModels.CompetitorReport) (*mongoModels.CompetitorReport, error) {
	if report.ID.IsZero() {
		// Insert new report
		result, err := r.reportsCollection.InsertOne(ctx, report)
		if err != nil {
			return nil, fmt.Errorf("CompetitorRepository.AddUpdateReport: failed to insert: %w", err)
		}
		report.ID = result.InsertedID.(primitive.ObjectID)
		return report, nil
	}
	// Update existing report
	update := bson.M{
		"$set": bson.M{
			"name":        report.Name,
			"competitors": report.Competitors,
		},
	}
	_, err := r.reportsCollection.UpdateByID(ctx, report.ID, update)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.AddUpdateReport: failed to update: %w", err)
	}
	return report, nil
}

// GetReportsByWorkspace fetches all competitor reports for a workspace and platform type.
func (r *CompetitorRepository) GetReportsByWorkspace(ctx context.Context, workspaceID, platformType string) ([]*mongoModels.CompetitorReport, error) {
	filter := bson.M{}
	if workspaceID != "" {
		oid, err := primitive.ObjectIDFromHex(workspaceID)
		if err != nil {
			return nil, fmt.Errorf("CompetitorRepository.GetReportsByWorkspace: invalid workspace_id: %w", err)
		}
		filter["workspace_id"] = oid
	}
	if platformType != "" {
		filter["platform_type"] = platformType
	}

	cursor, err := r.reportsCollection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportsByWorkspace: failed to find reports: %w", err)
	}
	defer cursor.Close(ctx)

	var reports []*mongoModels.CompetitorReport
	if err := cursor.All(ctx, &reports); err != nil {
		return nil, fmt.Errorf("CompetitorRepository.GetReportsByWorkspace: failed to decode: %w", err)
	}
	return reports, nil
}

// DeleteReport deletes a competitor report by ID.
func (r *CompetitorRepository) DeleteReport(ctx context.Context, reportID string) error {
	oid, err := primitive.ObjectIDFromHex(reportID)
	if err != nil {
		return fmt.Errorf("CompetitorRepository.DeleteReport: invalid report ID: %w", err)
	}
	result, err := r.reportsCollection.DeleteOne(ctx, bson.M{"_id": oid})
	if err != nil {
		return fmt.Errorf("CompetitorRepository.DeleteReport: failed to delete: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("CompetitorRepository.DeleteReport: report not found")
	}
	return nil
}
