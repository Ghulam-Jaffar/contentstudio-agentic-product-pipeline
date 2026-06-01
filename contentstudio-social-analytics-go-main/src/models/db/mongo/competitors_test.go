package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestCompetitor_GetCompetitorIDAsString_String(t *testing.T) {
	c := &Competitor{
		CompetitorID: "12345",
	}

	result := c.GetCompetitorIDAsString()
	if result != "12345" {
		t.Errorf("expected '12345', got '%s'", result)
	}
}

func TestCompetitor_GetCompetitorIDAsString_Int64(t *testing.T) {
	c := &Competitor{
		CompetitorID: int64(12345),
	}

	result := c.GetCompetitorIDAsString()
	if result != "12345" {
		t.Errorf("expected '12345', got '%s'", result)
	}
}

func TestCompetitor_GetCompetitorIDAsString_Float64(t *testing.T) {
	c := &Competitor{
		CompetitorID: float64(12345),
	}

	result := c.GetCompetitorIDAsString()
	if result != "12345" {
		t.Errorf("expected '12345', got '%s'", result)
	}
}

func TestCompetitor_GetCompetitorIDAsString_OtherType(t *testing.T) {
	c := &Competitor{
		CompetitorID: []byte("12345"),
	}

	result := c.GetCompetitorIDAsString()
	if result != "[49 50 51 52 53]" {
		t.Errorf("expected '[49 50 51 52 53]', got '%s'", result)
	}
}

func TestCompetitor_Struct(t *testing.T) {
	id := primitive.NewObjectID()
	c := Competitor{
		ID:           id,
		CompetitorID: "comp123",
		Name:         "Test Competitor",
		Slug:         "test-competitor",
		State:        "active",
		Image:        "https://example.com/image.jpg",
		Error:        "",
		IsActive:     true,
		PlatformType: "facebook",
	}

	if c.ID != id {
		t.Errorf("expected ID %v, got %v", id, c.ID)
	}
	if c.Name != "Test Competitor" {
		t.Errorf("expected name 'Test Competitor', got '%s'", c.Name)
	}
	if c.PlatformType != "facebook" {
		t.Errorf("expected platform 'facebook', got '%s'", c.PlatformType)
	}
}

func TestCompetitorReport_Struct(t *testing.T) {
	id := primitive.NewObjectID()
	workspaceID := primitive.NewObjectID()
	userID := primitive.NewObjectID()

	report := CompetitorReport{
		ID:              id,
		WorkspaceID:     workspaceID,
		Name:            "Test Report",
		Competitors:     []string{"comp1", "comp2", "comp3"},
		CreatedByUserID: userID,
	}

	if report.ID != id {
		t.Errorf("expected ID %v, got %v", id, report.ID)
	}
	if report.Name != "Test Report" {
		t.Errorf("expected name 'Test Report', got '%s'", report.Name)
	}
	if len(report.Competitors) != 3 {
		t.Errorf("expected 3 competitors, got %d", len(report.Competitors))
	}
}

func TestUser_Struct(t *testing.T) {
	id := primitive.NewObjectID()
	user := User{
		ID:        id,
		Email:     "test@example.com",
		FirstName: "John",
		LastName:  "Doe",
	}

	if user.ID != id {
		t.Errorf("expected ID %v, got %v", id, user.ID)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if user.FirstName != "John" {
		t.Errorf("expected first name 'John', got '%s'", user.FirstName)
	}
}
