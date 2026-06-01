// ============================================================
// twitter.gs — X (Twitter) Analytics
//
// Base URL : p.analytics  (Go analytics service, set by main.gs → getData())
// Prefix   : analytics/overview/twitter/{endpoint}
// Account  : account_id passed as &twitter_id=... query param
//
// ── Endpoints & Looker Studio field groups ───────────────────
//
//  getPageAndPostsInsights   → tw_fetchSummary()
//    JSON path: j.data | j.overview
//    Fields   : period, followers_count, following_count, tweet_count,
//               total_engagement, eng_rate, like_count, retweet_count,
//               reply_count, quote_count, bookmark_count,
//               *_diff (absolute vs prev), *_growth (% vs prev)
//
//  getFollowersTrendData     → tw_fetchFollowersTrend()
//    JSON path: j.buckets[], j.follower_count[], j.follower_count_daily[],
//               j.following_count_daily[]
//    Fields   : date, followers, followers_daily, following_count_daily
//
//  getEngagementImpressionData → tw_fetchEngagements()
//    JSON path: j.tweeted_at_date[], j.total_engagement[], j.impression_count[],
//               j.like_count[], j.retweet_count[], j.reply_count[],
//               j.tweet_count[], j.bookmark_count[], j.quote_count[]
//    Fields   : date, engagement, impressions, reach(=0), like_count,
//               retweet_count, reply_count, tweets_daily, bookmarks_daily, quotes_daily
//    Note     : impression_count → impressions alias
//               tweet_count → tweets_daily alias
//               bookmark_count → bookmarks_daily alias
//               quote_count → quotes_daily alias
//
//  getTopTweets              → tw_fetchTopTweets()
//    JSON path: j.top_tweets[]
//    Fields   : date, permalink, tweet_id, tweet_text, tweet_type,
//               tweet_media_url (media_url[].join(',')),
//               like_count, retweet_count, reply_count, quote_count,
//               bookmark_count, impressions (← impression_count), listed_count
//
//  getLeastTweets            → tw_fetchLeastTweets()
//    JSON path: j.least_tweets[]
//    Fields   : same as topTweets but prefixed lt_* to avoid routing collision
//
//  getCreditsUsedCount       → tw_fetchCredits()
//    JSON path: j.data.credits_used
//    Fields   : credits_used
//
// ── bestMatch routing ────────────────────────────────────────
// Looker Studio calls getData() with a list of requested field IDs.
// bestMatch() picks the endpoint group with the most overlapping IDs.
// Each group uses at least one unique discriminating field:
//   credits     → credits_used
//   summary     → followers_count, eng_rate, *_diff, *_growth
//   topTweets   → tweet_id (unique to top tweets)
//   leastTweets → lt_permalink (unique to least tweets)
//   audience    → followers, followers_daily
//   engagement  → tweets_daily, bookmarks_daily, quotes_daily
// ============================================================

function getFields_twitter(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('permalink').setName('Tweet URL').setType(types.URL);

  // Account overview
  fields.newMetric().setId('followers_count').setName('Followers').setType(types.NUMBER);
  fields.newMetric().setId('following_count').setName('Following').setType(types.NUMBER);
  fields.newMetric().setId('tweet_count').setName('Total Tweets').setType(types.NUMBER);
  fields.newMetric().setId('total_engagement').setName('Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('impressions').setName('Impressions').setType(types.NUMBER);
  fields.newMetric().setId('eng_rate').setName('Engagement Rate').setType(types.PERCENT);

  // Engagement metrics
  fields.newMetric().setId('like_count').setName('Likes').setType(types.NUMBER);
  fields.newMetric().setId('retweet_count').setName('Retweets').setType(types.NUMBER);
  fields.newMetric().setId('reply_count').setName('Replies').setType(types.NUMBER);
  fields.newMetric().setId('quote_count').setName('Quotes').setType(types.NUMBER);
  fields.newMetric().setId('bookmark_count').setName('Bookmarks').setType(types.NUMBER);

  // Daily / trend
  fields.newMetric().setId('followers').setName('Followers (Daily)').setType(types.NUMBER);
  fields.newMetric().setId('followers_daily').setName('Followers Daily Change').setType(types.NUMBER);
  fields.newMetric().setId('following_count_daily').setName('Following Daily Change').setType(types.NUMBER);
  fields.newMetric().setId('engagement').setName('Daily Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reach').setName('Daily Reach').setType(types.NUMBER);
  fields.newMetric().setId('tweets_daily').setName('Tweets Daily').setType(types.NUMBER);
  fields.newMetric().setId('bookmarks_daily').setName('Bookmarks Daily').setType(types.NUMBER);
  fields.newMetric().setId('quotes_daily').setName('Quotes Daily').setType(types.NUMBER);
  fields.newMetric().setId('credits_used').setName('Credits Used').setType(types.NUMBER);

  // Period comparison — absolute diff (current minus previous)
  fields.newMetric().setId('followers_count_diff').setName('Followers Diff').setType(types.NUMBER);
  fields.newMetric().setId('following_count_diff').setName('Following Diff').setType(types.NUMBER);
  fields.newMetric().setId('tweet_count_diff').setName('Tweets Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_engagement_diff').setName('Engagement Diff').setType(types.NUMBER);
  fields.newMetric().setId('like_count_diff').setName('Likes Diff').setType(types.NUMBER);
  fields.newMetric().setId('retweet_count_diff').setName('Retweets Diff').setType(types.NUMBER);
  fields.newMetric().setId('reply_count_diff').setName('Replies Diff').setType(types.NUMBER);
  fields.newMetric().setId('quote_count_diff').setName('Quotes Diff').setType(types.NUMBER);
  fields.newMetric().setId('bookmark_count_diff').setName('Bookmarks Diff').setType(types.NUMBER);

  // Period comparison — percentage growth
  fields.newMetric().setId('followers_count_growth').setName('Followers Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_engagement_growth').setName('Engagement Growth %').setType(types.PERCENT);
  fields.newMetric().setId('like_count_growth').setName('Likes Growth %').setType(types.PERCENT);
  fields.newMetric().setId('retweet_count_growth').setName('Retweets Growth %').setType(types.PERCENT);
  fields.newMetric().setId('reply_count_growth').setName('Replies Growth %').setType(types.PERCENT);
  fields.newMetric().setId('tweet_count_growth').setName('Tweets Growth %').setType(types.PERCENT);

  // Top tweets — additional dimensions
  fields.newDimension().setId('tweet_id').setName('Tweet ID').setType(types.TEXT);
  fields.newDimension().setId('tweet_text').setName('Tweet Text').setType(types.TEXT);
  fields.newDimension().setId('tweet_type').setName('Tweet Type').setType(types.TEXT);
  fields.newDimension().setId('tweet_media_url').setName('Tweet Media URLs').setType(types.TEXT);
  fields.newMetric().setId('listed_count').setName('Listed Count').setType(types.NUMBER);

  // Least tweets (lt_ prefix avoids routing collision with topTweets)
  fields.newDimension().setId('lt_permalink').setName('Least Tweet URL').setType(types.URL);
  fields.newDimension().setId('lt_tweet_id').setName('Least Tweet ID').setType(types.TEXT);
  fields.newDimension().setId('lt_tweet_text').setName('Least Tweet Text').setType(types.TEXT);
  fields.newDimension().setId('lt_tweet_type').setName('Least Tweet Type').setType(types.TEXT);
  fields.newDimension().setId('lt_tweet_media_url').setName('Least Tweet Media URLs').setType(types.TEXT);
  fields.newMetric().setId('lt_like_count').setName('Least Likes').setType(types.NUMBER);
  fields.newMetric().setId('lt_retweet_count').setName('Least Retweets').setType(types.NUMBER);
  fields.newMetric().setId('lt_reply_count').setName('Least Replies').setType(types.NUMBER);
  fields.newMetric().setId('lt_quote_count').setName('Least Quotes').setType(types.NUMBER);
  fields.newMetric().setId('lt_bookmark_count').setName('Least Bookmarks').setType(types.NUMBER);
  fields.newMetric().setId('lt_impressions').setName('Least Impressions').setType(types.NUMBER);
  fields.newMetric().setId('lt_listed_count').setName('Least Listed Count').setType(types.NUMBER);

  return fields;
}

// Builds the full API URL for a Twitter endpoint.
// p.analytics is injected by main.gs → getData() from Script/UserProperties.
// &twitter_id maps config's account_id to the Twitter platform identifier.
function buildUrl_tw(endpoint, p, extra) {
  return p.analytics + 'twitter/' + endpoint
    + buildBaseParams(p,
        '&twitter_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

// Routes each Looker Studio chart to the correct Twitter API endpoint.
// p._reqIds holds the list of field IDs the chart is requesting.
// bestMatch() scores each group by overlap and returns the winning key.
function getData_twitter(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    credits:     ['credits_used'],
    summary:     ['period','followers_count','following_count','tweet_count','total_engagement','eng_rate',
                  'like_count','retweet_count','reply_count','quote_count','bookmark_count',
                  'followers_count_diff','following_count_diff','tweet_count_diff','total_engagement_diff',
                  'like_count_diff','retweet_count_diff','reply_count_diff','quote_count_diff','bookmark_count_diff',
                  'followers_count_growth','total_engagement_growth','like_count_growth',
                  'retweet_count_growth','reply_count_growth','tweet_count_growth'],
    topTweets:   ['permalink','like_count','retweet_count','reply_count','quote_count','bookmark_count',
                  'tweet_id','tweet_text','tweet_type','tweet_media_url','listed_count'],
    leastTweets: ['lt_permalink','lt_like_count','lt_retweet_count','lt_reply_count','lt_quote_count','lt_bookmark_count','lt_impressions',
                  'lt_tweet_id','lt_tweet_text','lt_tweet_type','lt_tweet_media_url','lt_listed_count'],
    audience:    ['date','followers','followers_daily','following_count_daily'],
    engagement:  ['date','engagement','impressions','reach','tweets_daily','bookmarks_daily','quotes_daily']
  });

  switch(best) {
    case 'topTweets':   return tw_fetchTopTweets(p);
    case 'leastTweets': return tw_fetchLeastTweets(p);
    case 'credits':     return tw_fetchCredits(p);
    case 'audience':    return tw_fetchFollowersTrend(p);
    case 'engagement':  return tw_fetchEngagements(p);
    default:            return tw_fetchSummary(p);
  }
}

// Endpoint : GET analytics/overview/twitter/getPageAndPostsInsights
// Response : flat object — j.data (or j.overview as fallback)
// Returns  : one summary row with current-period totals and period-over-period
//            diff/growth values. Twitter does not split current/previous into
//            sub-objects; all fields are at the root of j.data.
function tw_fetchSummary(p) {
  var j = analyticsGet(buildUrl_tw('getPageAndPostsInsights', p), p.access_token);
  // Twitter returns a flat object in j.data rather than current/previous sub-objects
  var d = j.data || j.overview || {};
  return [Object.assign({ period: 'Current' }, d)];
}

// Endpoint : GET analytics/overview/twitter/getFollowersTrendData
// Response : flat object with parallel arrays keyed by date bucket
//   j.buckets[]               — ISO date strings (YYYY-MM-DD)
//   j.follower_count[]        — cumulative follower count per day → followers
//   j.follower_count_daily[]  — daily change in followers → followers_daily
//   j.following_count_daily[] — daily change in following → following_count_daily
// Skips days where follower_count is 0 (no data collected).
function tw_fetchFollowersTrend(p) {
  // FollowersTrendResponse is a flat object (not wrapped): parallel arrays + buckets[]
  var j = analyticsGet(buildUrl_tw('getFollowersTrendData', p), p.access_token);
  return (j.buckets || []).reduce(function(acc, date, i) {
    var f = (j.follower_count || [])[i] || 0;
    if (f > 0) acc.push({
      date:                 date.replace(/-/g, ''),
      followers:            f,
      followers_daily:      (j.follower_count_daily  || [])[i] || 0,
      following_count_daily:(j.following_count_daily || [])[i] || 0
    });
    return acc;
  }, []);
}

// Endpoint : GET analytics/overview/twitter/getEngagementImpressionData
// Response : flat object with parallel arrays — one entry per day
//   j.tweeted_at_date[]  — date strings (may include time; toDateStr strips it)
//   j.total_engagement[] → engagement
//   j.impression_count[] → impressions  (API uses impression_count, GS uses impressions)
//   j.like_count[]       → like_count
//   j.retweet_count[]    → retweet_count
//   j.reply_count[]      → reply_count
//   j.tweet_count[]      → tweets_daily  (tweets posted that day)
//   j.bookmark_count[]   → bookmarks_daily
//   j.quote_count[]      → quotes_daily
// reach is always 0 (Twitter API does not provide reach per day).
function tw_fetchEngagements(p) {
  // EngagementImpressionResponse: flat object with parallel arrays, one entry per tweet date
  var j = analyticsGet(buildUrl_tw('getEngagementImpressionData', p), p.access_token);
  return (j.tweeted_at_date || []).map(function(date, i) {
    return {
      date:            toDateStr(date) || date.replace(/-/g, ''),
      engagement:      (j.total_engagement  || [])[i] || 0,
      impressions:     (j.impression_count  || [])[i] || 0,
      reach:           0,
      like_count:      (j.like_count        || [])[i] || 0,
      retweet_count:   (j.retweet_count     || [])[i] || 0,
      reply_count:     (j.reply_count       || [])[i] || 0,
      tweets_daily:    (j.tweet_count       || [])[i] || 0,
      bookmarks_daily: (j.bookmark_count    || [])[i] || 0,
      quotes_daily:    (j.quote_count       || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/twitter/getTopTweets?limit=10&order_by=total_engagement
// Response : j.top_tweets[] — array of tweet objects
// Returns  : top 10 tweets sorted by total engagement.
//            tw_mapTweets() handles the field-level mapping for both top and least tweets.
function tw_fetchTopTweets(p) {
  var j = analyticsGet(
    buildUrl_tw('getTopTweets', p, '&limit=10&order_by=total_engagement'),
    p.access_token
  );
  return tw_mapTweets(j.top_tweets || []);
}

// Shared mapper for tweet objects — used by both tw_fetchTopTweets and tw_fetchLeastTweets.
// Key API→GS field aliases:
//   post.id              → tweet_id
//   post.impression_count → impressions
//   post.media_url[]     → tweet_media_url (joined with ',')
function tw_mapTweets(tweets) {
  return tweets.map(function(post) {
    return {
      date:            toDateStr(post.tweeted_at    || ''),
      permalink:       post.permalink               || '',
      tweet_id:        post.id                      || '',
      tweet_text:      post.tweet_text              || '',
      tweet_type:      post.tweet_type              || '',
      tweet_media_url: (post.media_url || []).join(','),
      like_count:      Number(post.like_count       || 0),
      retweet_count:   Number(post.retweet_count    || 0),
      reply_count:     Number(post.reply_count      || 0),
      quote_count:     Number(post.quote_count      || 0),
      bookmark_count:  Number(post.bookmark_count   || 0),
      impressions:     Number(post.impression_count || 0),
      listed_count:    Number(post.listed_count     || 0)
    };
  });
}

// Endpoint : GET analytics/overview/twitter/getLeastTweets?limit=10&order_by=total_engagement
// Response : j.least_tweets[] — array of tweet objects (same shape as top_tweets)
// All fields use the lt_ prefix (e.g. lt_like_count) to prevent bestMatch() routing
// from confusing this group with topTweets when both sets of fields are in the schema.
function tw_fetchLeastTweets(p) {
  var j = analyticsGet(
    buildUrl_tw('getLeastTweets', p, '&limit=10&order_by=total_engagement'),
    p.access_token
  );
  return (j.least_tweets || []).map(function(post) {
    return {
      date:              toDateStr(post.tweeted_at    || ''),
      lt_permalink:      post.permalink               || '',
      lt_tweet_id:       post.id                     || '',
      lt_tweet_text:     post.tweet_text              || '',
      lt_tweet_type:     post.tweet_type              || '',
      lt_tweet_media_url:(post.media_url || []).join(','),
      lt_like_count:     Number(post.like_count       || 0),
      lt_retweet_count:  Number(post.retweet_count    || 0),
      lt_reply_count:    Number(post.reply_count      || 0),
      lt_quote_count:    Number(post.quote_count      || 0),
      lt_bookmark_count: Number(post.bookmark_count   || 0),
      lt_impressions:    Number(post.impression_count || 0),
      lt_listed_count:   Number(post.listed_count     || 0)
    };
  });
}

// Endpoint : GET analytics/overview/twitter/getCreditsUsedCount
// Response : j.data.credits_used — total Twitter API credits consumed for this account
// Returns  : one row with credits_used. Shown in account settings / usage charts.
function tw_fetchCredits(p) {
  var j = analyticsGet(buildUrl_tw('getCreditsUsedCount', p), p.access_token);
  var d = j.data || {};
  return [{ period: 'Current', credits_used: Number(d.credits_used || 0) }];
}