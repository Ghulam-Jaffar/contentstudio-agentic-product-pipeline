package api

import (
	"encoding/json"
	"testing"
)

func TestCompetitorWorkRequest_Struct(t *testing.T) {
	request := CompetitorWorkRequest{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "facebook",
	}

	if request.ReportID != "report123" {
		t.Fatalf("expected ReportID 'report123', got %s", request.ReportID)
	}
	if request.PageID != "page456" {
		t.Fatalf("expected PageID 'page456', got %s", request.PageID)
	}
	if request.Channel != "facebook" {
		t.Fatalf("expected Channel 'facebook', got %s", request.Channel)
	}
}

func TestCompetitorWorkRequest_FacebookChannel(t *testing.T) {
	request := CompetitorWorkRequest{
		ReportID: "report123",
		PageID:   "fb_page_789",
		Channel:  "facebook",
	}

	if request.Channel != "facebook" {
		t.Fatalf("expected Channel 'facebook', got %s", request.Channel)
	}
}

func TestCompetitorWorkRequest_InstagramChannel(t *testing.T) {
	request := CompetitorWorkRequest{
		ReportID: "report456",
		PageID:   "ig_page_123",
		Channel:  "instagram",
	}

	if request.Channel != "instagram" {
		t.Fatalf("expected Channel 'instagram', got %s", request.Channel)
	}
}

func TestCompetitorWorkRequest_JSON_Marshal(t *testing.T) {
	request := CompetitorWorkRequest{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "facebook",
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result CompetitorWorkRequest
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.ReportID != request.ReportID {
		t.Fatalf("expected ReportID %s, got %s", request.ReportID, result.ReportID)
	}
	if result.PageID != request.PageID {
		t.Fatalf("expected PageID %s, got %s", request.PageID, result.PageID)
	}
	if result.Channel != request.Channel {
		t.Fatalf("expected Channel %s, got %s", request.Channel, result.Channel)
	}
}

func TestCompetitorWorkRequest_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"report_id": "report789",
		"page_id": "page012",
		"channel": "instagram"
	}`

	var request CompetitorWorkRequest
	err := json.Unmarshal([]byte(jsonData), &request)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if request.ReportID != "report789" {
		t.Fatalf("expected ReportID 'report789', got %s", request.ReportID)
	}
	if request.PageID != "page012" {
		t.Fatalf("expected PageID 'page012', got %s", request.PageID)
	}
	if request.Channel != "instagram" {
		t.Fatalf("expected Channel 'instagram', got %s", request.Channel)
	}
}
