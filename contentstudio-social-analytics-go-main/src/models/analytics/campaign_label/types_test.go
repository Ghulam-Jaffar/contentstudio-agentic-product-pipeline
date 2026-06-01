package campaign_label

import "testing"

func TestGetFlagSetup_DefaultsToAllPlatformsWhenNoAccountsSelected(t *testing.T) {
	req := &CampaignLabelRequest{}

	flags := req.GetFlagSetup()

	for _, platform := range []string{
		"facebook",
		"instagram",
		"linkedin",
		"pinterest",
		"youtube",
		"tiktok",
	} {
		if !flags[platform] {
			t.Fatalf("expected %s flag to be enabled when no accounts are selected", platform)
		}
	}
}

func TestGetFlagSetup_RespectsSelectedAccounts(t *testing.T) {
	req := &CampaignLabelRequest{
		FacebookAccounts: []string{"fb-1"},
		TiktokAccounts:   []string{"tt-1"},
	}

	flags := req.GetFlagSetup()

	if !flags["facebook"] {
		t.Fatal("expected facebook flag to be enabled")
	}
	if !flags["tiktok"] {
		t.Fatal("expected tiktok flag to be enabled")
	}
	if flags["instagram"] || flags["linkedin"] || flags["pinterest"] {
		t.Fatal("expected only selected platform flags to be enabled")
	}
	if !flags["youtube"] {
		t.Fatal("expected youtube to remain enabled")
	}
}
