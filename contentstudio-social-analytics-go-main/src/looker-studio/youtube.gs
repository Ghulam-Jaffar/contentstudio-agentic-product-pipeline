// ============================================================
// youtube.gs — YouTube Analytics
//
// Base URL : p.analytics_go  (Go analytics service, set by main.gs → getData())
// Prefix   : analytics/overview/youtube/{endpoint}
// Account  : account_id passed as &youtube_id=... query param
//
// ── Endpoints & Looker Studio field groups ───────────────────
//
// IMPORTANT: The YouTube API uses singular field names for engagement metrics
// (like, dislike, comment, share) both in overview and per-video responses.
// Looker Studio GS fields use plural/descriptive names. Aliases noted below.
//
//  overviewSummary              → yt_fetchSummary()
//    JSON path: j.overview.current / j.overview.previous
//    Key aliases (API → GS):
//      c.videos         → total_posts
//      c.like           → likes       (singular in API)
//      c.dislike        → dislikes
//      c.comment        → comments
//      c.share          → shares
//    Fields : period, subscribers, total_posts, views, watch_time,
//             avg_view_duration, likes, dislikes, comments, shares, engagement,
//             prev_* (same metrics for previous period)
//
//  overviewSubscriberTrend      → yt_fetchSubscriberTrend()
//    JSON path: j.buckets[], j.subscribers_gained_daily[], j.subscribers_total[]
//
//  overviewEngagementTrend      → yt_fetchEngagementTrend()
//    JSON path: j.buckets[], j.like_daily[], j.dislike_daily[], j.share_daily[],
//               j.comment_daily[], j.engagement_daily[],
//               j.like_total[], j.dislike_total[], j.share_total[],
//               j.comment_total[], j.engagement_total[]
//
//  overviewViewsTrend           → yt_fetchViewsTrend()
//    JSON path: j.buckets[], j.subscriber_views_daily[], j.non_subscriber_views_daily[],
//               j.video_views_daily[], j.subscriber_views_total[],
//               j.non_subscriber_views_total[], j.video_views_total[]
//
//  overviewWatchTimeTrend       → yt_fetchWatchTimeTrend()
//    JSON path: j.buckets[], j.subscriber_watch_time_daily[],
//               j.non_subscriber_watch_time_daily[], j.average_watch_time[],
//               j.subscriber_watch_time_total[], j.non_subscriber_watch_time_total[]
//
//  overviewFindVideo            → yt_fetchFindVideo()
//    JSON path: j.data[] — [{name, value, perc_value}]
//    traffic_source → item.name
//    source_value   → item.value
//    source_perc    → item.perc_value
//
//  overviewVideoSharing         → yt_fetchVideoSharing()
//    JSON path: j.data[] — [{name, value, perc_value}]  (same shape as findVideo)
//    sharing_platform → item.name  (unique field, prevents routing collision with findVideo)
//
//  getSortedTopPosts?limit=15   → yt_fetchTopPosts()
//    JSON path: j.top_posts[] — video objects
//    Key aliases (API → GS):
//      v.title                → video_title
//      v.share_url            → permalink
//      v.like                 → like_count
//      v.dislike              → dislike_count
//      v.comment              → comments_count
//      v.views                → video_views
//      v.average_view_duration → video_avg_duration
//      v.iframe_embed_url     → video_embed_url
//      v.share                → video_shares
//      v.average_view_percentage → video_avg_view_percentage
//
//  overviewLeastPosts           → yt_fetchLeastPosts()
//    JSON path: j.least_posts_ordered_by_views[] + j.least_posts_ordered_by_engagement[]
//    Same VideoItem shape as top_posts. Metric fields use lp_ prefix.
//    lp_sort_by indicates which sort order produced this row ('views' or 'engagement').
//    Shared dimensions (video_id, video_title, permalink, video_description, etc.) have no prefix.
//
//  overviewPerformanceAndVideoPostingSchedule → yt_fetchPerformanceSchedule()
//    JSON path: j.engagement.buckets[], j.engagement.*, j.video_views.*
//    ps_ prefix on all fields prevents routing collision with other date-based groups.
//
// ── bestMatch routing ────────────────────────────────────────
// Discriminating fields per group:
//   videoSharing → sharing_platform (unique; findVideo uses traffic_source)
//   findVideo    → traffic_source
//   topPosts     → video_id, video_title, like_count, video_engagement
//   leastPosts   → lp_sort_by, lp_like_count, lp_video_views
//   perfSchedule → ps_count, ps_likes, ps_engagement, ps_sub_views
//   watchTime    → subscriber_watch_time_daily, non_subscriber_watch_time_daily
//   views        → subscriber_views_daily, video_views_daily
//   engTrend     → like_daily, dislike_daily, like_total
//   subscribers  → subscribers_gained_daily, subscribers_total
//   summary      → period, subscribers, watch_time, avg_view_duration
// ============================================================

function getFields_youtube(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('video_id').setName('Video ID').setType(types.TEXT);
  fields.newDimension().setId('video_title').setName('Video Title').setType(types.TEXT);
  fields.newDimension().setId('permalink').setName('Video URL').setType(types.URL);
  fields.newDimension().setId('traffic_source').setName('Traffic Source').setType(types.TEXT);
  fields.newDimension().setId('sharing_platform').setName('Sharing Platform').setType(types.TEXT);

  // Channel overview (summary)
  fields.newMetric().setId('subscribers').setName('Subscribers').setType(types.NUMBER);
  fields.newMetric().setId('total_posts').setName('Total Videos').setType(types.NUMBER);
  fields.newMetric().setId('views').setName('Views').setType(types.NUMBER);
  fields.newMetric().setId('watch_time').setName('Watch Time (min)').setType(types.NUMBER);
  fields.newMetric().setId('avg_view_duration').setName('Avg View Duration (s)').setType(types.NUMBER);
  fields.newMetric().setId('likes').setName('Likes').setType(types.NUMBER);
  fields.newMetric().setId('dislikes').setName('Dislikes').setType(types.NUMBER);
  fields.newMetric().setId('comments').setName('Comments').setType(types.NUMBER);
  fields.newMetric().setId('shares').setName('Shares').setType(types.NUMBER);
  fields.newMetric().setId('engagement').setName('Engagement').setType(types.NUMBER);

  // Subscriber trend
  fields.newMetric().setId('subscribers_gained_daily').setName('Subscribers Gained Daily').setType(types.NUMBER);
  fields.newMetric().setId('subscribers_total').setName('Subscribers Total').setType(types.NUMBER);

  // Engagement trend
  fields.newMetric().setId('like_daily').setName('Likes Daily').setType(types.NUMBER);
  fields.newMetric().setId('dislike_daily').setName('Dislikes Daily').setType(types.NUMBER);
  fields.newMetric().setId('share_daily').setName('Shares Daily').setType(types.NUMBER);
  fields.newMetric().setId('comment_daily').setName('Comments Daily').setType(types.NUMBER);
  fields.newMetric().setId('engagement_daily').setName('Engagement Daily').setType(types.NUMBER);

  // Views trend
  fields.newMetric().setId('subscriber_views_daily').setName('Subscriber Views Daily').setType(types.NUMBER);
  fields.newMetric().setId('non_subscriber_views_daily').setName('Non-Subscriber Views Daily').setType(types.NUMBER);
  fields.newMetric().setId('video_views_daily').setName('Video Views Daily').setType(types.NUMBER);

  // Watch time trend
  fields.newMetric().setId('subscriber_watch_time_daily').setName('Subscriber Watch Time Daily').setType(types.NUMBER);
  fields.newMetric().setId('non_subscriber_watch_time_daily').setName('Non-Subscriber Watch Time Daily').setType(types.NUMBER);
  fields.newMetric().setId('average_watch_time').setName('Average Watch Time').setType(types.NUMBER);

  // Traffic sources / sharing breakdown
  fields.newMetric().setId('source_value').setName('Source Value').setType(types.NUMBER);
  fields.newMetric().setId('source_perc').setName('Source %').setType(types.PERCENT);

  // Per-video metrics (top posts)
  fields.newMetric().setId('like_count').setName('Video Likes').setType(types.NUMBER);
  fields.newMetric().setId('dislike_count').setName('Video Dislikes').setType(types.NUMBER);
  fields.newMetric().setId('comments_count').setName('Video Comments').setType(types.NUMBER);
  fields.newMetric().setId('video_views').setName('Video Views').setType(types.NUMBER);
  fields.newMetric().setId('minutes_watched').setName('Minutes Watched').setType(types.NUMBER);
  fields.newMetric().setId('video_avg_duration').setName('Avg View Duration (s)').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rate').setName('Engagement Rate').setType(types.PERCENT);
  // Per-video — additional fields
  fields.newDimension().setId('video_description').setName('Video Description').setType(types.TEXT);
  fields.newDimension().setId('video_thumbnail_url').setName('Video Thumbnail URL').setType(types.URL);
  fields.newDimension().setId('video_media_type').setName('Video Media Type').setType(types.TEXT);
  fields.newDimension().setId('video_embed_url').setName('Video Embed URL').setType(types.URL);
  fields.newMetric().setId('video_duration').setName('Video Duration (s)').setType(types.NUMBER);
  fields.newMetric().setId('video_engagement').setName('Video Engagement').setType(types.NUMBER);
  fields.newMetric().setId('video_red_views').setName('Video Red Views').setType(types.NUMBER);
  fields.newMetric().setId('video_favorites').setName('Video Favorites').setType(types.NUMBER);
  fields.newMetric().setId('video_subscribers_gained').setName('Video Subscribers Gained').setType(types.NUMBER);
  fields.newMetric().setId('video_shares').setName('Video Shares').setType(types.NUMBER);
  fields.newMetric().setId('video_red_minutes_watched').setName('Video Red Minutes Watched').setType(types.NUMBER);
  fields.newMetric().setId('video_avg_view_percentage').setName('Video Avg View Percentage').setType(types.PERCENT);

  // Channel overview — previous period
  fields.newMetric().setId('prev_subscribers').setName('Prev Subscribers').setType(types.NUMBER);
  fields.newMetric().setId('prev_total_posts').setName('Prev Total Videos').setType(types.NUMBER);
  fields.newMetric().setId('prev_views').setName('Prev Views').setType(types.NUMBER);
  fields.newMetric().setId('prev_watch_time').setName('Prev Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('prev_avg_view_duration').setName('Prev Avg View Duration').setType(types.NUMBER);
  fields.newMetric().setId('prev_likes').setName('Prev Likes').setType(types.NUMBER);
  fields.newMetric().setId('prev_dislikes').setName('Prev Dislikes').setType(types.NUMBER);
  fields.newMetric().setId('prev_comments').setName('Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('prev_shares').setName('Prev Shares').setType(types.NUMBER);
  fields.newMetric().setId('prev_engagement').setName('Prev Engagement').setType(types.NUMBER);

  // Least posts (lp_ prefix avoids routing collision with topPosts)
  fields.newDimension().setId('lp_sort_by').setName('Least Posts Sort By').setType(types.TEXT);
  fields.newMetric().setId('lp_like_count').setName('Least Post Likes').setType(types.NUMBER);
  fields.newMetric().setId('lp_dislike_count').setName('Least Post Dislikes').setType(types.NUMBER);
  fields.newMetric().setId('lp_comments_count').setName('Least Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('lp_video_views').setName('Least Post Views').setType(types.NUMBER);
  fields.newMetric().setId('lp_minutes_watched').setName('Least Post Min Watched').setType(types.NUMBER);
  fields.newMetric().setId('lp_avg_duration').setName('Least Post Avg Duration').setType(types.NUMBER);
  fields.newMetric().setId('lp_engagement_rate').setName('Least Post Eng Rate').setType(types.PERCENT);

  // Engagement trend — cumulative totals
  fields.newMetric().setId('like_total').setName('Likes Total').setType(types.NUMBER);
  fields.newMetric().setId('dislike_total').setName('Dislikes Total').setType(types.NUMBER);
  fields.newMetric().setId('share_total').setName('Shares Total').setType(types.NUMBER);
  fields.newMetric().setId('comment_total').setName('Comments Total').setType(types.NUMBER);
  fields.newMetric().setId('engagement_total').setName('Engagement Total').setType(types.NUMBER);

  // Views trend — cumulative totals
  fields.newMetric().setId('subscriber_views_total').setName('Subscriber Views Total').setType(types.NUMBER);
  fields.newMetric().setId('non_subscriber_views_total').setName('Non-Subscriber Views Total').setType(types.NUMBER);
  fields.newMetric().setId('video_views_total').setName('Video Views Total').setType(types.NUMBER);

  // Watch time trend — cumulative totals
  fields.newMetric().setId('subscriber_watch_time_total').setName('Subscriber Watch Time Total').setType(types.NUMBER);
  fields.newMetric().setId('non_subscriber_watch_time_total').setName('Non-Subscriber Watch Time Total').setType(types.NUMBER);

  // Performance & schedule (ps_ prefix to avoid collision with daily trend fields)
  fields.newMetric().setId('ps_count').setName('PS Post Count').setType(types.NUMBER);
  fields.newMetric().setId('ps_likes').setName('PS Likes').setType(types.NUMBER);
  fields.newMetric().setId('ps_dislikes').setName('PS Dislikes').setType(types.NUMBER);
  fields.newMetric().setId('ps_shares').setName('PS Shares').setType(types.NUMBER);
  fields.newMetric().setId('ps_comments').setName('PS Comments').setType(types.NUMBER);
  fields.newMetric().setId('ps_engagement').setName('PS Engagement').setType(types.NUMBER);
  fields.newMetric().setId('ps_sub_views').setName('PS Subscriber Views').setType(types.NUMBER);
  fields.newMetric().setId('ps_non_sub_views').setName('PS Non-Subscriber Views').setType(types.NUMBER);

  return fields;
}

// p.analytics_go is set by main.gs → getData() from Script Properties
// Builds the full API URL for a YouTube endpoint.
// p.analytics_go is injected by main.gs → getData() from Script/UserProperties.
// &youtube_id maps config's account_id to the YouTube channel identifier.
function buildUrl_yt(endpoint, p, extra) {
  return p.analytics_go + 'youtube/' + endpoint
    + buildBaseParams(p,
        '&youtube_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

// Routes each Looker Studio chart to the correct YouTube API endpoint.
// p._reqIds holds the list of field IDs the chart is requesting.
// bestMatch() scores each group by overlap and returns the winning key.
// Note: videoSharing must be listed before findVideo in the groups object
// because both share source_value and source_perc; sharing_platform is the
// discriminating field that separates them.
function getData_youtube(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    // sharing_platform is unique to videoSharing; must precede findVideo since both have source_value/source_perc
    videoSharing:  ['sharing_platform', 'source_value', 'source_perc'],
    findVideo:     ['traffic_source',   'source_value', 'source_perc'],
    topPosts:      ['video_id', 'video_title', 'permalink', 'like_count', 'dislike_count',
                    'comments_count', 'video_views', 'minutes_watched', 'video_avg_duration', 'engagement_rate',
                    'video_description', 'video_thumbnail_url', 'video_media_type', 'video_embed_url',
                    'video_duration', 'video_engagement', 'video_red_views', 'video_favorites',
                    'video_subscribers_gained', 'video_shares', 'video_red_minutes_watched', 'video_avg_view_percentage'],
    leastPosts:    ['lp_sort_by', 'lp_like_count', 'lp_video_views', 'lp_engagement_rate',
                    'lp_comments_count', 'lp_minutes_watched', 'lp_dislike_count', 'lp_avg_duration',
                    'video_description', 'video_thumbnail_url', 'video_media_type', 'video_embed_url',
                    'video_duration', 'video_engagement', 'video_red_views', 'video_favorites',
                    'video_subscribers_gained', 'video_shares', 'video_red_minutes_watched', 'video_avg_view_percentage'],
    perfSchedule:  ['date', 'ps_count', 'ps_likes', 'ps_engagement', 'ps_sub_views', 'ps_non_sub_views',
                    'ps_dislikes', 'ps_shares', 'ps_comments'],
    watchTime:     ['date', 'subscriber_watch_time_daily', 'non_subscriber_watch_time_daily', 'average_watch_time',
                    'subscriber_watch_time_total', 'non_subscriber_watch_time_total'],
    views:         ['date', 'subscriber_views_daily', 'non_subscriber_views_daily', 'video_views_daily',
                    'subscriber_views_total', 'non_subscriber_views_total', 'video_views_total'],
    engTrend:      ['date', 'like_daily', 'dislike_daily', 'share_daily', 'comment_daily', 'engagement_daily',
                    'like_total', 'dislike_total', 'share_total', 'comment_total', 'engagement_total'],
    subscribers:   ['date', 'subscribers_gained_daily', 'subscribers_total'],
    summary:       ['period', 'subscribers', 'total_posts', 'views', 'watch_time',
                    'avg_view_duration', 'likes', 'dislikes', 'comments', 'shares', 'engagement',
                    'prev_subscribers', 'prev_total_posts', 'prev_views', 'prev_watch_time',
                    'prev_avg_view_duration', 'prev_likes', 'prev_dislikes', 'prev_comments', 'prev_shares', 'prev_engagement']
  });

  switch(best) {
    case 'videoSharing': return yt_fetchVideoSharing(p);
    case 'findVideo':    return yt_fetchFindVideo(p);
    case 'topPosts':     return yt_fetchTopPosts(p);
    case 'leastPosts':   return yt_fetchLeastPosts(p);
    case 'perfSchedule': return yt_fetchPerformanceSchedule(p);
    case 'watchTime':    return yt_fetchWatchTimeTrend(p);
    case 'views':        return yt_fetchViewsTrend(p);
    case 'engTrend':     return yt_fetchEngagementTrend(p);
    case 'subscribers':  return yt_fetchSubscriberTrend(p);
    default:             return yt_fetchSummary(p);
  }
}

// Endpoint : GET analytics/overview/youtube/overviewSummary
// Response : j.overview.current / j.overview.previous
// Key aliases (API uses singular engagement field names):
//   c.videos  → total_posts
//   c.like    → likes
//   c.dislike → dislikes
//   c.comment → comments
//   c.share   → shares
// Returns one row with current and previous period metrics side by side.
function yt_fetchSummary(p) {
  var j = analyticsGet(buildUrl_yt('overviewSummary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                 'Current',
    subscribers:            c.subscribers       || 0,
    total_posts:            c.videos            || 0,
    views:                  c.views             || 0,
    watch_time:             c.watch_time        || 0,
    avg_view_duration:      c.avg_view_duration || 0,
    likes:                  c.like              || 0,
    dislikes:               c.dislike           || 0,
    comments:               c.comment           || 0,
    shares:                 c.share             || 0,
    engagement:             c.engagement        || 0,
    // Previous period
    prev_subscribers:       v.subscribers       || 0,
    prev_total_posts:       v.videos            || 0,
    prev_views:             v.views             || 0,
    prev_watch_time:        v.watch_time        || 0,
    prev_avg_view_duration: v.avg_view_duration || 0,
    prev_likes:             v.like              || 0,
    prev_dislikes:          v.dislike           || 0,
    prev_comments:          v.comment           || 0,
    prev_shares:            v.share             || 0,
    prev_engagement:        v.engagement        || 0
  }];
}

// Endpoint : GET analytics/overview/youtube/overviewSubscriberTrend
// Response : flat parallel arrays — j.buckets[], j.subscribers_gained_daily[], j.subscribers_total[]
function yt_fetchSubscriberTrend(p) {
  var j = analyticsGet(buildUrl_yt('overviewSubscriberTrend', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:                     date.replace(/-/g, ''),
      subscribers_gained_daily: (j.subscribers_gained_daily || [])[i] || 0,
      subscribers_total:        (j.subscribers_total        || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewEngagementTrend
// Response : flat parallel arrays — j.buckets[], j.like_daily[], j.dislike_daily[],
//   j.share_daily[], j.comment_daily[], j.engagement_daily[],
//   j.like_total[], j.dislike_total[], j.share_total[], j.comment_total[], j.engagement_total[]
// _daily fields = per-day delta; _total fields = cumulative running total.
function yt_fetchEngagementTrend(p) {
  var j = analyticsGet(buildUrl_yt('overviewEngagementTrend', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:             date.replace(/-/g, ''),
      like_daily:       (j.like_daily       || [])[i] || 0,
      like_total:       (j.like_total       || [])[i] || 0,
      dislike_daily:    (j.dislike_daily    || [])[i] || 0,
      dislike_total:    (j.dislike_total    || [])[i] || 0,
      share_daily:      (j.share_daily      || [])[i] || 0,
      share_total:      (j.share_total      || [])[i] || 0,
      comment_daily:    (j.comment_daily    || [])[i] || 0,
      comment_total:    (j.comment_total    || [])[i] || 0,
      engagement_daily: (j.engagement_daily || [])[i] || 0,
      engagement_total: (j.engagement_total || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewViewsTrend
// Response : flat parallel arrays — j.buckets[], j.subscriber_views_daily[],
//   j.non_subscriber_views_daily[], j.video_views_daily[],
//   j.subscriber_views_total[], j.non_subscriber_views_total[], j.video_views_total[]
function yt_fetchViewsTrend(p) {
  var j = analyticsGet(buildUrl_yt('overviewViewsTrend', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:                         date.replace(/-/g, ''),
      subscriber_views_daily:       (j.subscriber_views_daily     || [])[i] || 0,
      subscriber_views_total:       (j.subscriber_views_total     || [])[i] || 0,
      non_subscriber_views_daily:   (j.non_subscriber_views_daily || [])[i] || 0,
      non_subscriber_views_total:   (j.non_subscriber_views_total || [])[i] || 0,
      video_views_daily:            (j.video_views_daily          || [])[i] || 0,
      video_views_total:            (j.video_views_total          || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewWatchTimeTrend
// Response : flat parallel arrays — j.buckets[], j.subscriber_watch_time_daily[],
//   j.non_subscriber_watch_time_daily[], j.average_watch_time[],
//   j.subscriber_watch_time_total[], j.non_subscriber_watch_time_total[]
function yt_fetchWatchTimeTrend(p) {
  var j = analyticsGet(buildUrl_yt('overviewWatchTimeTrend', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:                             date.replace(/-/g, ''),
      subscriber_watch_time_daily:      (j.subscriber_watch_time_daily     || [])[i] || 0,
      subscriber_watch_time_total:      (j.subscriber_watch_time_total     || [])[i] || 0,
      non_subscriber_watch_time_daily:  (j.non_subscriber_watch_time_daily || [])[i] || 0,
      non_subscriber_watch_time_total:  (j.non_subscriber_watch_time_total || [])[i] || 0,
      average_watch_time:               (j.average_watch_time              || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewFindVideo
// Response : j.data[] — [{name, value, perc_value}]
//   item.name       → traffic_source  (e.g. "YouTube Search", "External", "Suggested")
//   item.value      → source_value    (view count from this source)
//   item.perc_value → source_perc     (percentage of total views)
function yt_fetchFindVideo(p) {
  var j = analyticsGet(buildUrl_yt('overviewFindVideo', p), p.access_token);
  return (j.data || []).map(function(item) {
    return {
      traffic_source: item.name       || '',
      source_value:   Number(item.value      || 0),
      source_perc:    Number(item.perc_value || 0)
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewVideoSharing
// Response : j.data[] — [{name, value, perc_value}]  (same shape as findVideo)
//   item.name       → sharing_platform  (e.g. "Twitter", "WhatsApp", "Facebook")
//   item.value      → source_value
//   item.perc_value → source_perc
// sharing_platform is the discriminating field that separates this group
// from findVideo in bestMatch() routing.
function yt_fetchVideoSharing(p) {
  var j = analyticsGet(buildUrl_yt('overviewVideoSharing', p), p.access_token);
  return (j.data || []).map(function(item) {
    return {
      sharing_platform: item.name       || '',
      source_value:     Number(item.value      || 0),
      source_perc:      Number(item.perc_value || 0)
    };
  });
}

// Endpoint : GET analytics/overview/youtube/getSortedTopPosts?limit=15&order_by=views
// Response : j.top_posts[] — array of VideoItem objects
// Key aliases (API uses singular form; GS uses descriptive names):
//   v.title                  → video_title
//   v.share_url              → permalink
//   v.like                   → like_count      (singular in API)
//   v.dislike                → dislike_count
//   v.comment                → comments_count
//   v.views                  → video_views
//   v.average_view_duration  → video_avg_duration
//   v.iframe_embed_url       → video_embed_url
//   v.share                  → video_shares    (singular in API)
//   v.average_view_percentage → video_avg_view_percentage
function yt_fetchTopPosts(p) {
  var j = analyticsGet(
    buildUrl_yt('getSortedTopPosts', p, '&limit=15&order_by=views'),
    p.access_token
  );
  return (j.top_posts || []).map(function(v) {
    return {
      date:                       toDateStr(v.published_at          || ''),
      video_id:                   v.video_id                        || '',
      video_title:                v.title                           || '',
      permalink:                  v.share_url                       || '',
      video_description:          v.description                     || '',
      video_thumbnail_url:        v.thumbnail_url                   || '',
      video_media_type:           v.media_type                      || '',
      video_embed_url:            v.iframe_embed_url                || '',
      like_count:                 Number(v.like                     || 0),
      dislike_count:              Number(v.dislike                  || 0),
      comments_count:             Number(v.comment                  || 0),
      video_views:                Number(v.views                    || 0),
      minutes_watched:            Number(v.minutes_watched          || 0),
      video_avg_duration:         Number(v.average_view_duration    || 0),
      engagement_rate:            Number(v.engagement_rate          || 0),
      video_duration:             Number(v.duration                 || 0),
      video_engagement:           Number(v.engagement               || 0),
      video_red_views:            Number(v.red_views                || 0),
      video_favorites:            Number(v.favorites                || 0),
      video_subscribers_gained:   Number(v.subscribers_gained       || 0),
      video_shares:               Number(v.share                    || 0),
      video_red_minutes_watched:  Number(v.red_minutes_watched      || 0),
      video_avg_view_percentage:  Number(v.average_view_percentage  || 0)
    };
  });
}

// Endpoint : GET analytics/overview/youtube/overviewLeastPosts
// Response : j.least_posts_ordered_by_views[] + j.least_posts_ordered_by_engagement[]
//   Two arrays of VideoItem (same shape as top_posts). Both are concatenated into one result.
//   lp_sort_by marks which sort produced each row ('views' or 'engagement').
// Metric fields use lp_ prefix to prevent routing collision with topPosts.
// Shared dimensions (video_id, video_title, permalink, video_description,
//   video_thumbnail_url, video_media_type, video_embed_url, video_duration,
//   video_engagement, video_red_views, video_favorites, video_subscribers_gained,
//   video_shares, video_red_minutes_watched, video_avg_view_percentage) have no prefix.
function yt_fetchLeastPosts(p) {
  var j = analyticsGet(buildUrl_yt('overviewLeastPosts', p), p.access_token);
  function mapRow(v, sortBy) {
    return {
      lp_sort_by:                 sortBy,
      date:                       toDateStr(v.published_at          || ''),
      video_id:                   v.video_id                        || '',
      video_title:                v.title                           || '',
      permalink:                  v.share_url                       || '',
      video_description:          v.description                     || '',
      video_thumbnail_url:        v.thumbnail_url                   || '',
      video_media_type:           v.media_type                      || '',
      video_embed_url:            v.iframe_embed_url                || '',
      lp_like_count:              Number(v.like                     || 0),
      lp_dislike_count:           Number(v.dislike                  || 0),
      lp_comments_count:          Number(v.comment                  || 0),
      lp_video_views:             Number(v.views                    || 0),
      lp_minutes_watched:         Number(v.minutes_watched          || 0),
      lp_avg_duration:            Number(v.average_view_duration    || 0),
      lp_engagement_rate:         Number(v.engagement_rate          || 0),
      video_duration:             Number(v.duration                 || 0),
      video_engagement:           Number(v.engagement               || 0),
      video_red_views:            Number(v.red_views                || 0),
      video_favorites:            Number(v.favorites                || 0),
      video_subscribers_gained:   Number(v.subscribers_gained       || 0),
      video_shares:               Number(v.share                    || 0),
      video_red_minutes_watched:  Number(v.red_minutes_watched      || 0),
      video_avg_view_percentage:  Number(v.average_view_percentage  || 0)
    };
  }
  var byViews = (j.least_posts_ordered_by_views      || []).map(function(v) { return mapRow(v, 'views'); });
  var byEng   = (j.least_posts_ordered_by_engagement || []).map(function(v) { return mapRow(v, 'engagement'); });
  return byViews.concat(byEng);
}

// Endpoint : GET analytics/overview/youtube/overviewPerformanceAndVideoPostingSchedule
// Response : j.engagement (parallel arrays) + j.video_views (parallel arrays)
//   j.engagement.buckets[]          — dates
//   j.engagement.count[]            → ps_count   (videos posted)
//   j.engagement.likes[]            → ps_likes
//   j.engagement.dislikes[]         → ps_dislikes
//   j.engagement.shares[]           → ps_shares
//   j.engagement.comments[]         → ps_comments
//   j.engagement.engagement[]       → ps_engagement
//   j.video_views.subscriber_views[]     → ps_sub_views
//   j.video_views.non_subscriber_views[] → ps_non_sub_views
// All fields use ps_ prefix to prevent routing collision with other date-based groups.
function yt_fetchPerformanceSchedule(p) {
  var j   = analyticsGet(buildUrl_yt('overviewPerformanceAndVideoPostingSchedule', p), p.access_token);
  var eng = j.engagement  || {};
  var vv  = j.video_views || {};
  return (eng.buckets || []).map(function(date, i) {
    return {
      date:             date.replace(/-/g, ''),
      ps_count:         (eng.count                  || [])[i] || 0,
      ps_likes:         (eng.likes                  || [])[i] || 0,
      ps_dislikes:      (eng.dislikes               || [])[i] || 0,
      ps_shares:        (eng.shares                 || [])[i] || 0,
      ps_comments:      (eng.comments               || [])[i] || 0,
      ps_engagement:    (eng.engagement             || [])[i] || 0,
      ps_sub_views:     (vv.subscriber_views        || [])[i] || 0,
      ps_non_sub_views: (vv.non_subscriber_views    || [])[i] || 0
    };
  });
}
