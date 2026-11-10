package campaign

import "time"

func GenerateCampaignTimes() (votesStartsAt, votesEndsAt time.Time) {
	now := time.Now().UTC()
	start := now.Add(24 * time.Hour)
	end := start.Add(30 * 24 * time.Hour)
	return start, end
}
