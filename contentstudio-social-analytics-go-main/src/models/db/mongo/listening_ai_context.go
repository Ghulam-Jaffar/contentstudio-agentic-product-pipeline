package mongo

// AICompetitor represents one competitor entry in a topic's AI context.
type AICompetitor struct {
	Name     string   `bson:"name"     json:"name"`
	Keywords []string `bson:"keywords" json:"keywords"`
}

// AIContext is the structured AI relevance context attached to a listening
// topic. It is denormalized per topic (β architecture) — copied from brand
// discovery output at topic creation time. Empty values are treated as "no
// context" by downstream consumers.
type AIContext struct {
	BrandName     string         `bson:"brand_name"     json:"brand_name"`
	BrandKeywords []string       `bson:"brand_keywords" json:"brand_keywords"`
	Industry      string         `bson:"industry"       json:"industry"`
	Competitors   []AICompetitor `bson:"competitors"    json:"competitors"`
}

// IsEmpty reports whether the AIContext has no usable signal. Used by
// downstream consumers to decide whether to fall back to topic-level fields.
func (c AIContext) IsEmpty() bool {
	return c.BrandName == "" &&
		len(c.BrandKeywords) == 0 &&
		c.Industry == "" &&
		len(c.Competitors) == 0
}

// TopicContextSnapshot is the per-topic snapshot read in one round trip and
// passed to the AI agents enrichment service. Lives in the model package to
// avoid an import cycle between the repo and the enrichment service.
type TopicContextSnapshot struct {
	AIContext     AIContext
	Hint          string
	TopicName     string
	TopicType     string
	TopicKeywords []string
}
