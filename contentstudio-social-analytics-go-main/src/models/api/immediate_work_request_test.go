package api

import (
	"encoding/json"
	"testing"
)

func TestImmediateWorkRequest_Struct(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "acc123",
		Channel:   "facebook",
	}

	if request.AccountID != "acc123" {
		t.Fatalf("expected AccountID 'acc123', got %s", request.AccountID)
	}
	if request.Channel != "facebook" {
		t.Fatalf("expected Channel 'facebook', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_FacebookChannel(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "fb_acc_123",
		Channel:   "facebook",
	}

	if request.Channel != "facebook" {
		t.Fatalf("expected Channel 'facebook', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_InstagramChannel(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "ig_acc_456",
		Channel:   "instagram",
	}

	if request.Channel != "instagram" {
		t.Fatalf("expected Channel 'instagram', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_LinkedinChannel(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "li_acc_789",
		Channel:   "linkedin",
	}

	if request.Channel != "linkedin" {
		t.Fatalf("expected Channel 'linkedin', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_TiktokChannel(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "tt_acc_012",
		Channel:   "tiktok",
	}

	if request.Channel != "tiktok" {
		t.Fatalf("expected Channel 'tiktok', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_JSON_Marshal(t *testing.T) {
	request := ImmediateWorkRequest{
		AccountID: "acc123",
		Channel:   "facebook",
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result ImmediateWorkRequest
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.AccountID != request.AccountID {
		t.Fatalf("expected AccountID %s, got %s", request.AccountID, result.AccountID)
	}
	if result.Channel != request.Channel {
		t.Fatalf("expected Channel %s, got %s", request.Channel, result.Channel)
	}
}

func TestImmediateWorkRequest_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"account_id": "acc456",
		"channel": "instagram"
	}`

	var request ImmediateWorkRequest
	err := json.Unmarshal([]byte(jsonData), &request)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if request.AccountID != "acc456" {
		t.Fatalf("expected AccountID 'acc456', got %s", request.AccountID)
	}
	if request.Channel != "instagram" {
		t.Fatalf("expected Channel 'instagram', got %s", request.Channel)
	}
}

func TestImmediateWorkRequest_AllChannels(t *testing.T) {
	channels := []string{"facebook", "instagram", "linkedin", "tiktok"}

	for _, channel := range channels {
		request := ImmediateWorkRequest{
			AccountID: "acc_test",
			Channel:   channel,
		}

		if request.Channel != channel {
			t.Fatalf("expected Channel %s, got %s", channel, request.Channel)
		}
	}
}
