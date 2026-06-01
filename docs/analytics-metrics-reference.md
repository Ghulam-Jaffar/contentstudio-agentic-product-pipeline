# ContentStudio Analytics — Metrics Reference

All metrics tracked per platform by the `contentstudio-social-analytics-go` pipeline (parser + processor + sink stages). Sourced from `/src/models/kafka/`.

---

## Overview Dashboard

Metrics surfaced in the cross-platform overview:

| Category | What's shown |
|---|---|
| **Audience** | Total followers/subscribers per connected account |
| **Engagement** | Total likes, comments, shares, reactions across all posts |
| **Reach / Impressions** | Combined reach and impression counts |
| **Posts Published** | Total posts across all platforms in selected date range |
| **Growth** | Follower/subscriber gain vs. prior period |
| **Sentiment** | Positive vs. negative sentiment (via social listening) |
| **Post Performance** | Top-performing posts by engagement rate |
| **Best Time to Post** | Active audience heatmap (day-of-week × hour-of-day) |

---

## Facebook

### Page-Level Metrics
| Metric | Description |
|---|---|
| `PageFans` | Total page likes (fans) |
| `PageFollows` | Total page followers |
| `PageFansCity / Country / Locale / Age / Gender / GenderAge` | Audience demographics breakdown |
| `PageFanAddsUnique` | New fans gained |
| `PageFanRemovesUnique` | Fans lost (unlikes) |
| `PageFansByLikeSource / UnlikeSource` | How fans were gained/lost |
| `PageTotalActions` | Total actions taken on the page |
| `PagePostEngagements` | Total post engagements |
| `PageActionsPostReactionsLikeTotal / LoveTotal / AngerTotal` | Reaction type breakdown |
| `PageImpressions / PageImpressionsUnique` | Total & unique impressions |
| `PageImpressionsOrganic / PageImpressionsPaid` | Organic vs paid impressions |
| `PageMediaView` | Media views on the page |
| `PageVideoViews / PageVideoViewsPaid / PageVideoViewsOrganic` | Video view breakdown |
| `PageVideoViewsAutoplayed / PageVideoViewsClickToPlay / PageVideoRepeatViews` | Video view type |
| `PageNegativeFeedback / PagePositiveFeedback` | Sentiment signals |
| `PageNegativeFeedbackByType / PagePositiveFeedbackByType` | Feedback breakdown by type |
| `PageViews` | Total page views |
| `ActiveUsers` | Active users on page |
| `PrimeTime` | Peak activity hours |
| `PageFansOnline` | Hourly online fans breakdown |
| `PostsCount` | Total posts published |
| `LikesCount` | Total likes on page posts |
| `TalkingAboutCount` | People talking about the page |
| `TypeCount` | Post type breakdown (link, photo, video) |
| `MessageCount` | Messages sent / received |
| `PositiveSentiment / NegativeSentiment` | Sentiment score |
| `DayOfWeek / Year / Month` | Date partitions |

### Post-Level Metrics
| Metric | Description |
|---|---|
| `Like / Love / Haha / Wow / Sad / Angry / Thankful` | Individual reaction counts |
| `Total` (reactions) | Total reaction count |
| `Comments` | Comment count |
| `Shares` | Share count |
| `TotalEngagement` | Likes + comments + shares |
| `PostClicks / PostClicksUnique` | Total & unique post clicks |
| `PostEngaged / PostEngagedUsers` | Engaged users count |
| `PostImpressions / PostImpressionsUnique` | Total & unique impressions |
| `PostImpressionsPaid / PostImpressionsPaidUnique` | Paid impressions |
| `PostImpressionsOrganic / PostImpressionsOrganicUnique` | Organic impressions |
| `PostImpressionsViral / PostImpressionsViralUnique` | Viral impressions |
| `TotalImpressions` | Sum of all impression sources |
| `PostMediaView / PostMediaViewAds` | Media views |
| `PostVideoViews / PostVideoViewTime / PostVideoPlayTime` | Video-specific stats |
| `PostNegativeFeedback / PostNegativeFeedbackUnique` | Negative feedback |
| `StatusType` | Post status type |
| `MediaType` | Post media type |
| `DayOfWeek / HourOfDay` | Publish timing |

### Video-Level Metrics
| Metric | Description |
|---|---|
| `TotalVideoFollowers` | Followers at time of video |
| `TotalVideoViews / TotalVideoViewsUnique` | Total & unique views |
| `TotalVideoViewsAutoplayed / TotalVideoViewsClickedToPlay` | View initiation type |
| `TotalVideoViewsOrganic / TotalVideoViewsOrganicUnique` | Organic views |
| `TotalVideoViewsPaid / TotalVideoViewsPaidUnique` | Paid views |
| `TotalVideoViewsSoundOn` | Views with sound on |
| `TotalVideoPlayCount` | Total play events |
| `TotalVideoAvgTimeWatched` | Average watch duration |
| `TotalVideoViewTotalTime` | Aggregate watch time |
| `TotalVideoViewTotalTimeOrganic / TotalVideoViewTotalTimePaid` | Watch time by distribution |
| `TotalVideoViewTimeByDistributionType / Country / Region / AgeBucketAndGender` | Demographic breakdowns |
| `TotalVideoCompleteViews / TotalVideoCompleteViewsUnique` | 100% completion views |
| `TotalVideoCompleteViewsAutoplayed / ClickedToPlay / Organic / Paid` | Completion by type |
| `TotalVideo10sViews / 10sViewsUnique / 10sOrganic / 10sPaid / 10sSoundOn` | 10-second view metrics |
| `TotalVideo15sViews` | 15-second views |
| `TotalVideo15minExcludesShorterViews / Unique` | 15-min view threshold |
| `TotalVideo30sViews / 30sOrganic / 30sPaid / 30sSoundOn` | 30-second view metrics |
| `TotalVideo60sExcludesShorterViews` | 60-second views |
| `TotalVideoImpressions / TotalVideoImpressionsUnique` | Video impressions |
| `TotalVideoImpressionsPaid / Organic / Viral / Fan / FanPaid` | Impression by source |
| `TotalVideoRetentionGraphAutoplayed / ClickedToPlay / GenderMale / GenderFemale` | Audience retention curves |
| `TotalVideoAdBreakEarnings / AdImpressions / AdCPM` | Ad monetization stats |
| `TotalVideoConsumptionRate` | Watch-through rate |
| `TotalVideoStoriesByActionType` | Story actions |
| `TotalVideoReactionsByTypeTotal` | Video reactions by type |

---

## Instagram

### Account-Level Metrics
| Metric | Description |
|---|---|
| `FollowersCount` | Total followers |
| `FollowsCount` | Accounts followed |
| `MediaCount` | Total posts published |
| `AccountsEngaged` | Unique accounts engaged |
| `Engagement` | Total engagement count |
| `Likes / Comments / Shares / Saves` | Engagement type breakdown |
| `Impressions` | Total impressions |
| `Reach` | Unique accounts reached |
| `Views` | Total views |
| `ProfileViews` | Profile visits |
| `AudienceAge / Gender / GenderAge / Locale` | Audience demographics |
| `AudienceCity / AudienceCountry` | Audience location |
| `AudienceAgeByEngagement / GenderByEngagement / GenderAgeByEngagement` | Demographics by engagement |
| `AudienceCityByEngagement / CountryByEngagement` | Location by engagement |
| `AudienceAgeByReach / GenderByReach / GenderAgeByReach` | Demographics by reach |
| `AudienceCityByReach / CountryByReach` | Location by reach |
| `OnlineFollowers` | Hourly online followers heatmap |

### Post-Level Metrics
| Metric | Description |
|---|---|
| `LikeCount` | Likes |
| `CommentsCount` | Comments |
| `Saved` | Saves |
| `Shares` | Shares |
| `Replies` | Story/reel replies |
| `Engagement` | Likes + comments + saved |
| `Impressions` | Total impressions |
| `Reach` | Unique reach |
| `Views` | Total views |
| `VideoViews` | Video view count |
| `ReelsAvgWatchTime` | Average reel watch duration |
| `ReelsTotalWatchTime` | Total reel watch time |
| `Exits` | Story exits |
| `TapsForward / TapsBack` | Story navigation taps |
| `MediaType` | IMAGE / VIDEO / CAROUSEL_ALBUM / REELS |
| `EntityType` | FEED / STORIES / REELS |
| `DayOfWeek / HourOfDay` | Publish timing |

---

## Twitter / X

### Account-Level Metrics
| Metric | Description |
|---|---|
| `FollowersCount` | Total followers |
| `FollowingCount` | Accounts following |
| `TweetCount` | Total tweets published |
| `ListedCount` | Times added to lists |
| `LikeCount` | Total likes given by account |

### Tweet-Level Metrics
| Metric | Description |
|---|---|
| `ImpressionCount` | Total tweet impressions |
| `RetweetCount` | Retweets |
| `ReplyCount` | Replies |
| `LikeCount` | Likes |
| `BookmarkCount` | Bookmarks |
| `QuoteCount` | Quote tweets |
| `TotalEngagement` | Retweets + replies + likes + bookmarks + quotes |
| `TweetType` | tweet / retweet / quote / reply |

---

## LinkedIn

### Page-Level Metrics
| Metric | Description |
|---|---|
| `OrganicFollowerCount` | Organic followers |
| `PaidFollowerCount` | Paid/ad-driven followers |
| `TotalFollowerCount` | Total followers |
| `DailyFollowerCount` | Daily follower snapshot |
| `Engagement` | Total engagement |
| `Comments` | Comments |
| `Reactions` | Reactions |
| `Repost` | Reshares |
| `PostClicks` | Post link clicks |
| `ImpressionCount` | Total impressions |
| `Reach` | Unique reach |
| `PageViews / UniqueVisitors` | Total & unique page views |
| `DesktopPageViews / MobilePageViews` | Device breakdown |
| `OverviewPageViews / AboutPageViews / JobsPageViews / PeoplePageViews` | Page section views |
| `CareersPageViews / LifeAtPageViews / InsightsPageViews / ProductsPageViews` | Page section views (continued) |
| `PageViewsByCountry / Region / Industry / Seniority / Function / StaffCount` | Views by audience segment |
| `FollowersBySeniority / Industry / Country / City` | Follower demographics |

### Post-Level Metrics
| Metric | Description |
|---|---|
| `Comments` | Comments |
| `Favorites` | Reactions (likes) |
| `Repost` | Reshares |
| `TotalEngagement` | Sum of all engagement |
| `Impressions` | Impressions |
| `Reach` | Unique reach |
| `PostClicks` | Link clicks |
| `DayOfWeek / HourOfDay` | Publish timing |

---

## YouTube

### Channel-Level Metrics
| Metric | Description |
|---|---|
| `SubscriberCount` | Total subscribers |
| `VideoCount` | Total videos published |
| `ViewCount` | All-time channel views |

### Video-Level Metrics
| Metric | Description |
|---|---|
| `Views` | Total views |
| `RedViews` | YouTube Premium views |
| `Likes` | Likes |
| `Dislikes` | Dislikes |
| `Comments` | Comments |
| `Shares` | Shares |
| `Saved` | Saves |
| `MinutesWatched` | Total watch time (minutes) |
| `RedMinutesWatched` | Premium watch time |
| `AvgViewDuration` | Average view duration |
| `AvgViewPercentage` | Average % of video watched |
| `SubscribersGained` | New subscribers from this video |
| `Impressions` | YouTube search/browse impressions |
| `ImpressionsClickThroughRate` | Thumbnail CTR |
| `MediaType` | video / short |

### Daily Activity Insights
| Metric | Description |
|---|---|
| `Views / RedViews` | Daily view counts |
| `Likes / Dislikes / Comments / Shares` | Daily engagement |
| `EstimatedMinutesWatched / EstimatedRedMinutesWatched` | Daily watch time |
| `AvgViewDuration / AvgViewPercentage` | Daily average watch |
| `SubscribersGained` | Daily subscriber gain |

### Traffic Source Insights
| Metric | Description |
|---|---|
| `YTSearchViews` | Views from YouTube Search |
| `RelatedVideoViews` | Views from suggested videos |
| `YTChannelViews` | Views from channel page |
| `YTOtherPageViews` | Other YouTube pages |
| `ExtURLViews` | External URL referrals |
| `PlaylistViews` | Views from playlists |
| `NotificationViews` | Views from notifications |
| `ShortsViews` | Views via Shorts feed |
| `SubscriberViews` | Views from subscribers |
| `PaidViews` | Views from paid ads |
| `AnnotationViews / EndScreenViews / CampaignCardView` | Interactive element views |
| `SubscriberWatchTime / NonSubscriberWatchTime` | Watch time by audience type |

### Sharing Insights (44 platforms tracked)
`WhatsApp`, `Facebook`, `Twitter`, `LinkedIn`, `Reddit`, `Pinterest`, `Telegram`, `Discord`, `Skype`, `Viber`, `WeChat`, `Weibo`, `VKontakte`, `Email (Mail)`, `CopyPaste`, `Embed`, `Blogger`, `Tumblr`, `Myspace`, `Digg`, `Dropbox`, and more.

---

## TikTok

### Account-Level Metrics
| Metric | Description |
|---|---|
| `TotalFollowerCount` | Total followers |
| `TotalFollowingCount` | Accounts following |
| `TotalVideoCount` | Total videos published |
| `TotalLikeCount` | Total likes received |
| `TotalVideoLikes` | Likes across all videos |
| `TotalVideoComments` | Comments across all videos |
| `TotalVideoShares` | Shares across all videos |
| `TotalVideoViews` | Views across all videos |

### Video-Level Metrics
| Metric | Description |
|---|---|
| `ViewCount` | Video views |
| `LikeCount` | Likes |
| `CommentCount` | Comments |
| `ShareCount` | Shares |
| `EngagementCount` | Total engagements |
| `EngagementRate` | Engagement as % of views |
| `Duration` | Video length (seconds) |

---

## Pinterest

### Account-Level Metrics
| Metric | Description |
|---|---|
| `FollowerCount` | Total followers |
| `FollowingCount` | Accounts following |
| `PinCount` | Total pins published |
| `BoardCount` | Total boards |
| `MonthlyViews` | Monthly unique viewers |

### Account Performance Insights
| Metric | Description |
|---|---|
| `Impression` | Total impressions |
| `PinClicks / PinClickRate` | Pin click count & rate |
| `OutboundClicks / ClickthroughRate` | External link clicks & rate |
| `Saves / SaveRate` | Saves & save rate |
| `Engagement / EngagementRate` | Total engagement & rate |
| `ProfileVisit` | Profile visits |
| `Closeup` | Closeup views |
| `VideoMRCView / VideoStart / Video10sView` | Video milestone views |
| `VideoAvgWatchTime / VideoV50WatchTime` | Video watch time metrics |
| `FullScreenPlay / FullScreenPlaytime` | Full-screen video events |
| `Quartile95sPercent` | 95% watch-through rate |

### Board-Level Metrics
| Metric | Description |
|---|---|
| `PinCount` | Pins in this board |
| `FollowerCount` | Board followers |
| `CollaboratorCount` | Board collaborators |

### Pin-Level Metrics
| Metric | Description |
|---|---|
| `Impression` | Impressions |
| `PinClicks / OutboundClicks` | Pin & outbound clicks |
| `Saves / SaveRate` | Saves & rate |
| `Clickthrough / ClickthroughRate` | External CTR |
| `Engagement / EngagementRate` | Engagement & rate |
| `VideoMRCView / VideoStart / Video10sView` | Video view milestones |
| `VideoAvgWatchTime / VideoV50WatchTime` | Watch time |
| `FullScreenPlay / FullScreenPlaytime` | Full-screen events |
| `ProfileVisit / Closeup` | User interaction signals |
| `Quartile95sPercent / UserFollow` | Watch-through & follows |
| `DayOfWeek / HourOfDay` | Publish timing |

---

## Google Business Profile (GMB)

### Daily Performance Metrics
| Metric | Description |
|---|---|
| `BusinessImpressionsDesktopMaps` | Impressions on Google Maps (desktop) |
| `BusinessImpressionsDesktopSearch` | Impressions on Google Search (desktop) |
| `BusinessImpressionsMobileMaps` | Impressions on Google Maps (mobile) |
| `BusinessImpressionsMobileSearch` | Impressions on Google Search (mobile) |
| `CallClicks` | Phone call button clicks |
| `WebsiteClicks` | Website link clicks |
| `BusinessDirectionRequests` | Get directions requests |
| `BusinessConversations` | Conversations initiated |
| `BusinessBookings` | Booking actions |
| `BusinessFoodOrders` | Food order actions |
| `BusinessFoodMenuClicks` | Food menu clicks |

### Search Keywords
| Metric | Description |
|---|---|
| `ImpressionsValue` | Search impressions for keyword |
| `ImpressionsThreshold` | Threshold for suppressed data |
| `Keyword` | Search query |
| `KeywordMonth` | Month of data (YYYY-MM) |

### Local Posts
Tracks post state, topic type, media assets, and timestamps.

### Reviews
| Metric | Description |
|---|---|
| `StarRating` | 1–5 star rating |
| `Comment` | Review text |
| `ReplyComment` | Business reply text |

---

## Meta Ads (Facebook & Instagram Ads)

### Campaign / Ad Set / Ad Status
`Status`, `EffectiveStatus`, `Objective`, `DailyBudget`, `LifetimeBudget`, `BudgetRemaining`

### Performance Insights (all hierarchy levels)
| Metric | Description |
|---|---|
| `Spend` | Total amount spent |
| `Impressions` | Ad impressions |
| `Reach` | Unique people reached |
| `Clicks / UniqueClicks` | Total & unique clicks |
| `CTR / UniqueCTR` | Click-through rate |
| `CPC` | Cost per click |
| `CPM` | Cost per 1,000 impressions |
| `CPP` | Cost per person reached |
| `Frequency` | Average times each person saw the ad |
| `Actions` | Conversion/action events (type + value pairs) |

### Demographic Breakdowns
- **Age / Gender** — impressions, reach, clicks, spend, CTR, CPM, CPC by age bucket and gender
- **Device / Platform / Position** — by impression device, publisher platform, and placement
- **Region / Country** — by geographic location

---

## Social Listening

### Mention Metrics
| Metric | Description |
|---|---|
| `TotalEngagement` | Total engagement on mention |
| `LikesCount / CommentsCount / SharesCount` | Engagement breakdown |
| `AuthorFollowers` | Mention author's audience size |
| `SentimentLabel` | positive / negative / neutral |
| `SentimentScore` | Confidence score |
| `MatchedKeywords` | Keywords that triggered the mention |
| `AITags` | AI-generated topic/category tags |
| `Platform` | Source platform |
| `Language` | Content language |

---

## Metric Coverage Summary

| Platform | Account Metrics | Post Metrics | Demographic Breakdowns | Notes |
|---|---|---|---|---|
| Facebook | 30+ | 20+ | Age, gender, city, country, locale | Extensive video insights, ad breaks |
| Instagram | 20+ | 15+ | Age, gender, city, country by engagement & reach | Reels, Stories, Carousels |
| Twitter / X | 5 | 7 | — | Impression-based model |
| LinkedIn | 20+ | 7 | Seniority, industry, country, city, function | B2B demographics |
| YouTube | 3 (channel) | 10+ | Traffic sources, sharing platforms | 44 sharing platforms tracked |
| TikTok | 8 | 7 | — | Engagement rate per video |
| Pinterest | 10+ | 12+ | — | Board + pin + account levels |
| GMB | 11 daily | — | Desktop vs mobile, Maps vs Search | Reviews + keywords |
| Meta Ads | Budget + status | Spend, CTR, CPC, CPM | Age/gender, device, region | Campaign → Ad Set → Ad hierarchy |
| Social Listening | Sentiment + tags | 6 | Platform, language | Cross-platform mention tracking |
