package api

type ListeningAnalyticsResponse struct {
	Status                bool                  `json:"status"`
	Engagement            EngagementMetrics     `json:"engagement"`
	MentionsOverTime      []TopicTimeSeries     `json:"mentionsOverTime"`
	SentimentTrend        []SentimentTimeSeries `json:"sentimentTrend"`
	SentimentDistribution []DistributionItem    `json:"sentimentDistribution"`
	PlatformDistribution  []DistributionItem    `json:"platformDistribution"`
	TagDistribution       []DistributionItem    `json:"tagDistribution"`
}

type EngagementMetrics struct {
	TotalMentions          int     `json:"totalMentions"`
	PositiveSentimentShare float64 `json:"positiveSentimentShare"`
	AvgDailyMentions       float64 `json:"avgDailyMentions"`
	TopicsTracked          int     `json:"topicsTracked"`
}

type TopicTimeSeries struct {
	TopicID   string            `json:"topicId"`
	TopicName string            `json:"topicName"`
	Data      []TimeSeriesPoint `json:"data"`
}

type SentimentTimeSeries struct {
	Sentiment string            `json:"sentiment"`
	Data      []TimeSeriesPoint `json:"data"`
}

type TimeSeriesPoint struct {
	Date  string `json:"date"` // ISO Date 2006-01-02
	Value int    `json:"value"`
}

type DistributionItem struct {
	Label string `json:"label"`
	Value int    `json:"value"`
}
