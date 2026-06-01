// ============================================================
// tiktok.gs — TikTok Analytics
//
// Base URL : p.analytics_go  (Go analytics service, set by main.gs → getData())
// Prefix   : analytics/overview/tiktok/{endpoint}
// Account  : account_id passed as &tiktok_id=... query param
//
// ── Endpoints & Looker Studio field groups ───────────────────
//
//  getPageAndPostsInsights      → tt_fetchSummary()
//    JSON path: j.data (flat object)
//    Key aliases (API field → GS field):
//      total_followers   → followers
//      total_video_views → video_views
//      total_likes       → likes
//      total_comments    → comments
//      total_shares      → shares
//      total_engagements → engagement
//    Fields : period, followers, total_followings, total_posts, video_views,
//             likes, comments, shares, engagement,
//             total_*_diff (absolute vs prev), total_*_growth (% vs prev)
//
//  getPageFollowersAndViews     → tt_fetchFollowersAndViews()
//    JSON path: j.data[0] — object with parallel arrays
//      day_bucket[]           — dates
//      followers_count[]      → followers
//      followers_count_diff[] → followers_daily
//      views_per_day[]        → profile_views
//      views_per_day_diff[]   → profile_views_diff
//    Skips days where followers_count is 0.
//
//  getDailyEngagementsData      → tt_fetchDailyEngagements()
//    JSON path: j.data[0] — object with parallel arrays
//      days_bucket[]          — dates
//      daily_video_likes[]    → likes
//      daily_video_comments[] → comments
//      daily_video_shares[]   → shares
//      daily_engagement[]     → engagement
//      total_video_likes[]    → total_video_likes  (cumulative)
//      total_video_comments[] → total_video_comments
//      total_video_shares[]   → total_video_shares
//      total_engagement[]     → total_engagement
//
//  getPostsAndEngagements       → tt_fetchPostsAndEngagements()
//    JSON path: j.data[0] — object with parallel arrays
//      days_bucket[]          — dates
//      sum_view_count[]       → video_views
//      sum_like_count[]       → likes
//      sum_comments_count[]   → comments
//      sum_share_count[]      → shares
//      sum_engagement_count[] → engagement
//      post_count[]           → total_posts
//
//  getTopAndLeastPerformingPosts → tt_fetchTopPosts() / tt_fetchLeastPosts()
//    JSON path: j.data.top_posts[] / j.data.least_posts[]
//    Same endpoint, different JSON path — post objects have identical shape.
//    Key aliases (API → GS):
//      post.created_time  → date
//      post.category      → media_type
//      post.share_url     → permalink
//      post.likes_count   → like_count
//      post.comments_count → comments_count
//      post.views_count   → post_views
//      post.shares_count  → post_shares
//      post.engagements_count → post_engagement
//      post.hashtags[]    → hashtags (joined with ',')
//    leastPosts uses lp_ prefix on metrics to prevent routing collision with topPosts.
//
// ── bestMatch routing ────────────────────────────────────────
// Discriminating fields per group:
//   topPosts         → post_id, embed_link, hashtags, engagement_rate
//   leastPosts       → lp_like_count, lp_engagement_rate (all lp_* fields)
//   summary          → period, eng_rate, total_followings, *_diff, *_growth
//   postsEngagements → video_views + total_posts (daily posting volume)
//   audience         → followers_daily, profile_views, profile_views_diff
//   daily            → total_video_likes, total_video_comments, total_video_shares
// ============================================================

function getFields_tiktok(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('media_type').setName('Media Type').setType(types.TEXT);
  fields.newDimension().setId('permalink').setName('Video URL').setType(types.URL);

  // Page overview
  fields.newMetric().setId('followers').setName('Followers').setType(types.NUMBER);
  fields.newMetric().setId('followers_daily').setName('Followers Daily Change').setType(types.NUMBER);
  fields.newMetric().setId('total_followings').setName('Followings').setType(types.NUMBER);
  fields.newMetric().setId('total_posts').setName('Total Videos').setType(types.NUMBER);
  fields.newMetric().setId('video_views').setName('Video Views').setType(types.NUMBER);
  fields.newMetric().setId('profile_views').setName('Profile Views').setType(types.NUMBER);
  fields.newMetric().setId('profile_views_diff').setName('Profile Views Daily Change').setType(types.NUMBER);
  fields.newMetric().setId('likes').setName('Likes').setType(types.NUMBER);
  fields.newMetric().setId('comments').setName('Comments').setType(types.NUMBER);
  fields.newMetric().setId('shares').setName('Shares').setType(types.NUMBER);
  fields.newMetric().setId('engagement').setName('Engagement').setType(types.NUMBER);
  fields.newMetric().setId('eng_rate').setName('Engagement Rate').setType(types.PERCENT);

  // Top posts — metrics
  fields.newMetric().setId('like_count').setName('Post Likes').setType(types.NUMBER);
  fields.newMetric().setId('comments_count').setName('Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('post_views').setName('Post Views').setType(types.NUMBER);
  fields.newMetric().setId('post_shares').setName('Post Shares').setType(types.NUMBER);
  fields.newMetric().setId('post_engagement').setName('Post Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rate').setName('Post Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('total_follower_count').setName('Post Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('duration').setName('Post Duration').setType(types.NUMBER);
  fields.newMetric().setId('height').setName('Post Height').setType(types.NUMBER);
  fields.newMetric().setId('width').setName('Post Width').setType(types.NUMBER);
  // Top posts — additional dimensions
  fields.newDimension().setId('post_id').setName('Post ID').setType(types.TEXT);
  fields.newDimension().setId('post_title').setName('Post Title').setType(types.TEXT);
  fields.newDimension().setId('post_description').setName('Post Description').setType(types.TEXT);
  fields.newDimension().setId('cover_image_url').setName('Cover Image URL').setType(types.URL);
  fields.newDimension().setId('embed_link').setName('Embed Link').setType(types.URL);
  fields.newDimension().setId('profile_link').setName('Profile Link').setType(types.URL);
  fields.newDimension().setId('hashtags').setName('Hashtags').setType(types.TEXT);

  // Period comparison — absolute diff (current minus previous)
  fields.newMetric().setId('total_likes_diff').setName('Likes Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_comments_diff').setName('Comments Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_shares_diff').setName('Shares Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_engagements_diff').setName('Engagements Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_posts_diff').setName('Posts Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_followers_diff').setName('Followers Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_video_views_diff').setName('Video Views Diff').setType(types.NUMBER);

  // Period comparison — percentage growth
  fields.newMetric().setId('total_likes_growth').setName('Likes Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_comments_growth').setName('Comments Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_shares_growth').setName('Shares Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_engagements_growth').setName('Engagements Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_posts_growth').setName('Posts Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_followers_growth').setName('Followers Growth %').setType(types.PERCENT);
  fields.newMetric().setId('total_video_views_growth').setName('Video Views Growth %').setType(types.PERCENT);

  // Least posts (lp_ prefix avoids routing collision with topPosts)
  fields.newMetric().setId('lp_like_count').setName('Least Post Likes').setType(types.NUMBER);
  fields.newMetric().setId('lp_comments_count').setName('Least Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('lp_post_views').setName('Least Post Views').setType(types.NUMBER);
  fields.newMetric().setId('lp_post_shares').setName('Least Post Shares').setType(types.NUMBER);
  fields.newMetric().setId('lp_post_engagement').setName('Least Post Engagement').setType(types.NUMBER);
  fields.newMetric().setId('lp_engagement_rate').setName('Least Post Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('lp_total_follower_count').setName('Least Post Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('lp_duration').setName('Least Post Duration').setType(types.NUMBER);
  fields.newMetric().setId('lp_height').setName('Least Post Height').setType(types.NUMBER);
  fields.newMetric().setId('lp_width').setName('Least Post Width').setType(types.NUMBER);

  // Daily engagements — cumulative totals
  fields.newMetric().setId('total_video_likes').setName('Total Video Likes').setType(types.NUMBER);
  fields.newMetric().setId('total_video_comments').setName('Total Video Comments').setType(types.NUMBER);
  fields.newMetric().setId('total_video_shares').setName('Total Video Shares').setType(types.NUMBER);
  fields.newMetric().setId('total_engagement').setName('Total Engagement').setType(types.NUMBER);

  // Period comparison — followings
  fields.newMetric().setId('total_followings_diff').setName('Followings Diff').setType(types.NUMBER);
  fields.newMetric().setId('total_followings_growth').setName('Followings Growth %').setType(types.PERCENT);

  return fields;
}

// Builds the full API URL for a TikTok endpoint.
// p.analytics_go is injected by main.gs → getData() from Script/UserProperties.
// &tiktok_id maps config's account_id to the TikTok platform identifier.
function buildUrl_tt(endpoint, p, extra) {
  return p.analytics_go + 'tiktok/' + endpoint
    + buildBaseParams(p,
        '&tiktok_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

// Routes each Looker Studio chart to the correct TikTok API endpoint.
// p._reqIds holds the list of field IDs the chart is requesting.
// bestMatch() scores each group by overlap and returns the winning key.
function getData_tiktok(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    topPosts:         ['permalink','like_count','comments_count','post_views','post_shares','post_engagement',
                       'engagement_rate','total_follower_count','duration','height','width',
                       'post_id','post_title','post_description','cover_image_url','embed_link','profile_link','hashtags'],
    leastPosts:       ['lp_like_count','lp_comments_count','lp_post_views','lp_post_shares','lp_post_engagement',
                       'lp_engagement_rate','lp_total_follower_count','lp_duration','lp_height','lp_width'],
    summary:          ['period','eng_rate','followers','total_posts','video_views','likes','comments','shares','engagement','profile_views',
                       'total_followings','total_followings_diff','total_followings_growth',
                       'total_likes_diff','total_comments_diff','total_shares_diff','total_engagements_diff',
                       'total_posts_diff','total_followers_diff','total_video_views_diff',
                       'total_likes_growth','total_comments_growth','total_shares_growth','total_engagements_growth',
                       'total_posts_growth','total_followers_growth','total_video_views_growth'],
    postsEngagements: ['date','video_views','total_posts'],
    audience:         ['date','followers','followers_daily','profile_views','profile_views_diff'],
    daily:            ['date','likes','comments','shares','engagement',
                       'total_video_likes','total_video_comments','total_video_shares','total_engagement']
  });

  switch(best) {
    case 'topPosts':         return tt_fetchTopPosts(p);
    case 'leastPosts':       return tt_fetchLeastPosts(p);
    case 'postsEngagements': return tt_fetchPostsAndEngagements(p);
    case 'audience':         return tt_fetchFollowersAndViews(p);
    case 'daily':            return tt_fetchDailyEngagements(p);
    default:                 return tt_fetchSummary(p);
  }
}

// Endpoint : GET analytics/overview/tiktok/getPageAndPostsInsights
// Response : j.data — flat object with account-level totals
// Key aliases (API → GS):
//   d.total_followers   → followers
//   d.total_video_views → video_views
//   d.total_likes       → likes
//   d.total_comments    → comments
//   d.total_shares      → shares
//   d.total_engagements → engagement
//   d.total_*_diff      → *_diff  (absolute change vs previous period)
//   d.total_*_growth    → *_growth (percentage change vs previous period)
function tt_fetchSummary(p) {
  var j = analyticsGet(buildUrl_tt('getPageAndPostsInsights', p), p.access_token);
  var d = j.data || {};
  return [{
    period:                    'Current',
    followers:                 d.total_followers           || 0,
    total_followings:          d.total_followings          || 0,
    total_posts:               d.total_posts               || 0,
    video_views:               d.total_video_views         || 0,
    likes:                     d.total_likes               || 0,
    comments:                  d.total_comments            || 0,
    shares:                    d.total_shares              || 0,
    engagement:                d.total_engagements         || 0,
    // Absolute diff vs previous period
    total_likes_diff:          d.total_likes_diff          || 0,
    total_comments_diff:       d.total_comments_diff       || 0,
    total_shares_diff:         d.total_shares_diff         || 0,
    total_engagements_diff:    d.total_engagements_diff    || 0,
    total_posts_diff:          d.total_posts_diff          || 0,
    total_followers_diff:      d.total_followers_diff      || 0,
    total_video_views_diff:    d.total_video_views_diff    || 0,
    total_followings_diff:     d.total_followings_diff     || 0,
    // Percentage growth vs previous period
    total_likes_growth:        d.total_likes_growth        || 0,
    total_comments_growth:     d.total_comments_growth     || 0,
    total_shares_growth:       d.total_shares_growth       || 0,
    total_engagements_growth:  d.total_engagements_growth  || 0,
    total_posts_growth:        d.total_posts_growth        || 0,
    total_followers_growth:    d.total_followers_growth    || 0,
    total_video_views_growth:  d.total_video_views_growth  || 0,
    total_followings_growth:   d.total_followings_growth   || 0
  }];
}

// Endpoint : GET analytics/overview/tiktok/getPageFollowersAndViews
// Response : j.data[0] — object with parallel arrays (one element per day)
//   d.day_bucket[]           — ISO date strings
//   d.followers_count[]      → followers
//   d.followers_count_diff[] → followers_daily (daily change)
//   d.views_per_day[]        → profile_views
//   d.views_per_day_diff[]   → profile_views_diff
// Skips days where followers_count is 0 (API returns 0 for days with no data).
function tt_fetchFollowersAndViews(p) {
  var j = analyticsGet(buildUrl_tt('getPageFollowersAndViews', p), p.access_token);
  var d = ((j.data || [])[0]) || {};
  return (d.day_bucket || []).reduce(function(acc, date, i) {
    var f = (d.followers_count || [])[i] || 0;
    if (f > 0) acc.push({
      date:              date.replace(/-/g, ''),
      followers:         f,
      followers_daily:   (d.followers_count_diff || [])[i] || 0,
      profile_views:     (d.views_per_day        || [])[i] || 0,
      profile_views_diff:(d.views_per_day_diff   || [])[i] || 0
    });
    return acc;
  }, []);
}

// Endpoint : GET analytics/overview/tiktok/getDailyEngagementsData
// Response : j.data[0] — object with parallel arrays (one element per day)
//   d.days_bucket[]          — ISO date strings
//   d.daily_video_likes[]    → likes
//   d.daily_video_comments[] → comments
//   d.daily_video_shares[]   → shares
//   d.daily_engagement[]     → engagement
//   d.total_video_likes[]    → total_video_likes  (cumulative daily total)
//   d.total_video_comments[] → total_video_comments
//   d.total_video_shares[]   → total_video_shares
//   d.total_engagement[]     → total_engagement
function tt_fetchDailyEngagements(p) {
  var j = analyticsGet(buildUrl_tt('getDailyEngagementsData', p), p.access_token);
  var d = ((j.data || [])[0]) || {};
  return (d.days_bucket || []).map(function(date, i) {
    return {
      date:                 date.replace(/-/g, ''),
      likes:                (d.daily_video_likes    || [])[i] || 0,
      comments:             (d.daily_video_comments || [])[i] || 0,
      shares:               (d.daily_video_shares   || [])[i] || 0,
      engagement:           (d.daily_engagement     || [])[i] || 0,
      total_video_likes:    (d.total_video_likes    || [])[i] || 0,
      total_video_comments: (d.total_video_comments || [])[i] || 0,
      total_video_shares:   (d.total_video_shares   || [])[i] || 0,
      total_engagement:     (d.total_engagement     || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/tiktok/getPostsAndEngagements
// Response : j.data[0] — object with parallel arrays (one element per day)
//   d.days_bucket[]          — ISO date strings
//   d.sum_view_count[]       → video_views
//   d.sum_like_count[]       → likes
//   d.sum_comments_count[]   → comments
//   d.sum_share_count[]      → shares
//   d.sum_engagement_count[] → engagement
//   d.post_count[]           → total_posts
function tt_fetchPostsAndEngagements(p) {
  var j = analyticsGet(buildUrl_tt('getPostsAndEngagements', p), p.access_token);
  var d = ((j.data || [])[0]) || {};
  return (d.days_bucket || []).map(function(date, i) {
    return {
      date:        date.replace(/-/g, ''),
      video_views: (d.sum_view_count       || [])[i] || 0,
      likes:       (d.sum_like_count       || [])[i] || 0,
      comments:    (d.sum_comments_count   || [])[i] || 0,
      shares:      (d.sum_share_count      || [])[i] || 0,
      engagement:  (d.sum_engagement_count || [])[i] || 0,
      total_posts: (d.post_count           || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/tiktok/getTopAndLeastPerformingPosts
// Response : j.data.top_posts[] — top performing post objects
// Key aliases (API → GS):
//   post.created_time    → date
//   post.category        → media_type  (VIDEO, IMAGE, etc.)
//   post.share_url       → permalink
//   post.likes_count     → like_count
//   post.comments_count  → comments_count
//   post.views_count     → post_views
//   post.shares_count    → post_shares
//   post.engagements_count → post_engagement
//   post.hashtags[]      → hashtags (joined with ',')
function tt_fetchTopPosts(p) {
  var j     = analyticsGet(buildUrl_tt('getTopAndLeastPerformingPosts', p), p.access_token);
  var posts = ((j.data || {}).top_posts) || [];
  return posts.map(function(post) {
    return {
      date:                 toDateStr(post.created_time    || ''),
      media_type:           post.category                  || '',
      permalink:            post.share_url                 || '',
      post_id:              post.post_id                   || '',
      post_title:           post.title                     || '',
      post_description:     post.post_description          || '',
      cover_image_url:      post.cover_image_url           || '',
      embed_link:           post.embed_link                || '',
      profile_link:         post.profile_link              || '',
      hashtags:             (post.hashtags || []).join(','),
      like_count:           Number(post.likes_count        || 0),
      comments_count:       Number(post.comments_count     || 0),
      post_views:           Number(post.views_count        || 0),
      post_shares:          Number(post.shares_count       || 0),
      post_engagement:      Number(post.engagements_count  || 0),
      engagement_rate:      Number(post.engagement_rate    || 0),
      total_follower_count: Number(post.total_follower_count|| 0),
      duration:             Number(post.duration           || 0),
      height:               Number(post.height             || 0),
      width:                Number(post.width              || 0)
    };
  });
}

// Endpoint : GET analytics/overview/tiktok/getTopAndLeastPerformingPosts
// Response : j.data.least_posts[] — least performing post objects (same shape as top_posts)
// All metric fields use the lp_ prefix (e.g. lp_like_count) to prevent bestMatch()
// from routing to tt_fetchTopPosts when leastPosts fields are requested.
// Shared dimensions (date, media_type, permalink, post_id, etc.) have no prefix.
function tt_fetchLeastPosts(p) {
  var j     = analyticsGet(buildUrl_tt('getTopAndLeastPerformingPosts', p), p.access_token);
  var posts = ((j.data || {}).least_posts) || [];
  return posts.map(function(post) {
    return {
      date:                    toDateStr(post.created_time     || ''),
      media_type:              post.category                   || '',
      permalink:               post.share_url                  || '',
      post_id:                 post.post_id                    || '',
      post_title:              post.title                      || '',
      post_description:        post.post_description           || '',
      cover_image_url:         post.cover_image_url            || '',
      embed_link:              post.embed_link                 || '',
      profile_link:            post.profile_link               || '',
      hashtags:                (post.hashtags || []).join(','),
      lp_like_count:           Number(post.likes_count         || 0),
      lp_comments_count:       Number(post.comments_count      || 0),
      lp_post_views:           Number(post.views_count         || 0),
      lp_post_shares:          Number(post.shares_count        || 0),
      lp_post_engagement:      Number(post.engagements_count   || 0),
      lp_engagement_rate:      Number(post.engagement_rate     || 0),
      lp_total_follower_count: Number(post.total_follower_count|| 0),
      lp_duration:             Number(post.duration            || 0),
      lp_height:               Number(post.height              || 0),
      lp_width:                Number(post.width               || 0)
    };
  });
}