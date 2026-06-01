// ============================================================
// pinterest.gs — Pinterest Analytics
//
// Base URL : p.analytics_go  (Go analytics service, set by main.gs → getData())
// Prefix   : analytics/overview/pinterest/{endpoint}
// Account  : account_id passed as &pinterest_id=... query param
//
// ── Endpoints & Looker Studio field groups ───────────────────
//
//  overviewSummary              → pin_fetchSummary()
//    JSON path: j.overview.current / j.overview.previous
//    Fields : period, follower_count, impressions, pin_clicks, outbound_clicks,
//             saves, total_engagement, prev_* (same metrics for previous period)
//
//  overviewFollowers            → pin_fetchFollowerTrend()  [daily]
//    JSON path: j.buckets[], j.followers_daily[], j.followers_gained[]
//
//  overviewImpressions          → pin_fetchImpressionsTrend()  [daily]
//    JSON path: j.buckets[], j.impressions_daily[], j.impressions_total[]
//
//  overviewEngagement           → pin_fetchEngagementTrend()  [daily]
//    JSON path: j.buckets[], j.saves_daily[], j.pin_clicks_daily[],
//               j.outbound_clicks_daily[], j.engagement_daily[],
//               j.saves_total[], j.pin_clicks_total[],
//               j.outbound_clicks_total[], j.engagement_total[]
//
//  overviewPinPostingRollup     → pin_fetchPinRollup()
//    JSON path: j.overview.current / j.overview.previous
//    Fields : period, total_pins, impressions, pin_clicks, outbound_clicks,
//             saves, video_views, avg_watch_time,
//             quartile_95s_percent_view, video_10s_view, prev_*
//
//  overviewTopPins?limit=10     → pin_fetchTopPins()
//    JSON path: j.top[] — array of pin objects
//    Key aliases (API → GS):
//      pin.pin_clicks      → pin_clicks_top   (prefixed to avoid routing collision with summary)
//      pin.total_engagement → pin_engagement
//      pin.product_tags[]  → pin_product_tags (joined with ',')
//
//  overviewPinPostingPerformance → pin_fetchPinPerformance()  [daily]
//    JSON path: j.buckets[], j.pins_count[], j.pin_clicks[], j.outbound_clicks[],
//               j.saves[], j.engagements[], j.impressions[]
//    Discriminating fields: pin_performance_engagement, pin_performance_impressions
//    (prefixed to avoid routing collision with summary impressions)
//
//  overviewPinPostingPerDay     → pin_fetchPinPosting()  [daily]
//    JSON path: j.buckets[], j.pins_count[]
//    pins_posting_count is the only field — unique discriminator for this group
//
// ── bestMatch routing ────────────────────────────────────────
// Discriminating fields per group:
//   topPins        → pin_id, pin_board_name, pin_impressions, pin_saves, pin_clicks_top
//   rollup         → total_pins, video_views, avg_watch_time, quartile_95s_percent_view
//   followers      → followers_daily, followers_gained
//   impTrend       → impressions_daily, impressions_total
//   engTrend       → saves_daily, pin_clicks_daily, outbound_clicks_daily, saves_total
//   performance    → pins_count, pin_performance_engagement, pin_performance_impressions
//   pinPosting     → pins_posting_count
//   summary        → period, follower_count, total_engagement, prev_follower_count
// ============================================================

function getFields_pinterest(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('media_type').setName('Pin Type').setType(types.TEXT);
  fields.newDimension().setId('permalink').setName('Pin URL').setType(types.URL);

  // Account overview (summary) — current period
  fields.newMetric().setId('follower_count').setName('Followers').setType(types.NUMBER);
  fields.newMetric().setId('impressions').setName('Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pin_clicks').setName('Pin Clicks').setType(types.NUMBER);
  fields.newMetric().setId('outbound_clicks').setName('Outbound Clicks').setType(types.NUMBER);
  fields.newMetric().setId('saves').setName('Saves').setType(types.NUMBER);
  fields.newMetric().setId('total_engagement').setName('Total Engagement').setType(types.NUMBER);

  // Account overview — previous period
  fields.newMetric().setId('prev_follower_count').setName('Prev Followers').setType(types.NUMBER);
  fields.newMetric().setId('prev_impressions').setName('Prev Impressions').setType(types.NUMBER);
  fields.newMetric().setId('prev_pin_clicks').setName('Prev Pin Clicks').setType(types.NUMBER);
  fields.newMetric().setId('prev_outbound_clicks').setName('Prev Outbound Clicks').setType(types.NUMBER);
  fields.newMetric().setId('prev_saves').setName('Prev Saves').setType(types.NUMBER);
  fields.newMetric().setId('prev_total_engagement').setName('Prev Total Engagement').setType(types.NUMBER);

  // Follower trend
  fields.newMetric().setId('followers_daily').setName('Followers Daily').setType(types.NUMBER);
  fields.newMetric().setId('followers_gained').setName('Followers Gained').setType(types.NUMBER);

  // Impressions trend
  fields.newMetric().setId('impressions_daily').setName('Impressions Daily').setType(types.NUMBER);
  fields.newMetric().setId('impressions_total').setName('Impressions Total').setType(types.NUMBER);

  // Engagement trend — per-day deltas
  fields.newMetric().setId('saves_daily').setName('Saves Daily').setType(types.NUMBER);
  fields.newMetric().setId('pin_clicks_daily').setName('Pin Clicks Daily').setType(types.NUMBER);
  fields.newMetric().setId('outbound_clicks_daily').setName('Outbound Clicks Daily').setType(types.NUMBER);
  fields.newMetric().setId('engagement_daily').setName('Engagement Daily').setType(types.NUMBER);

  // Engagement trend — cumulative totals
  fields.newMetric().setId('saves_total').setName('Saves Total').setType(types.NUMBER);
  fields.newMetric().setId('pin_clicks_total').setName('Pin Clicks Total').setType(types.NUMBER);
  fields.newMetric().setId('outbound_clicks_total').setName('Outbound Clicks Total').setType(types.NUMBER);
  fields.newMetric().setId('engagement_total').setName('Engagement Total').setType(types.NUMBER);

  // Pin rollup — current period
  fields.newMetric().setId('total_pins').setName('Total Pins').setType(types.NUMBER);
  fields.newMetric().setId('video_views').setName('Video Views').setType(types.NUMBER);
  fields.newMetric().setId('avg_watch_time').setName('Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('quartile_95s_percent_view').setName('Quartile 95% View').setType(types.PERCENT);
  fields.newMetric().setId('video_10s_view').setName('Video 10s View').setType(types.NUMBER);

  // Pin rollup — previous period
  fields.newMetric().setId('prev_total_pins').setName('Prev Total Pins').setType(types.NUMBER);
  fields.newMetric().setId('prev_video_views').setName('Prev Video Views').setType(types.NUMBER);
  fields.newMetric().setId('prev_avg_watch_time').setName('Prev Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('prev_quartile_95s_percent_view').setName('Prev Quartile 95% View').setType(types.PERCENT);
  fields.newMetric().setId('prev_video_10s_view').setName('Prev Video 10s View').setType(types.NUMBER);

  // Top/least pins — metrics
  fields.newMetric().setId('pin_impressions').setName('Pin Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pin_saves').setName('Pin Saves').setType(types.NUMBER);
  fields.newMetric().setId('pin_clicks_top').setName('Pin Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pin_outbound_clicks').setName('Pin Outbound Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pin_engagement').setName('Pin Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pin_engagement_rate').setName('Pin Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('pin_height').setName('Pin Height').setType(types.NUMBER);
  fields.newMetric().setId('pin_width').setName('Pin Width').setType(types.NUMBER);

  // Top/least pins — dimensions
  fields.newDimension().setId('pin_id').setName('Pin ID').setType(types.TEXT);
  fields.newDimension().setId('pin_board_name').setName('Board Name').setType(types.TEXT);
  fields.newDimension().setId('pin_embed_link').setName('Pin Embed Link').setType(types.URL);
  fields.newDimension().setId('pin_title').setName('Pin Title').setType(types.TEXT);
  fields.newDimension().setId('pin_description').setName('Pin Description').setType(types.TEXT);
  fields.newDimension().setId('pin_board_owner').setName('Board Owner').setType(types.TEXT);
  fields.newDimension().setId('pin_cover_image_url').setName('Pin Cover Image URL').setType(types.URL);
  fields.newDimension().setId('pin_dominant_color').setName('Pin Dominant Color').setType(types.TEXT);
  fields.newDimension().setId('pin_creative_type').setName('Pin Creative Type').setType(types.TEXT);
  fields.newDimension().setId('pin_product_tags').setName('Pin Product Tags').setType(types.TEXT);

  // Pin posting performance (distinct IDs avoid routing conflicts with account-level impressions)
  fields.newMetric().setId('pins_count').setName('Pins Count').setType(types.NUMBER);
  fields.newMetric().setId('pin_performance_engagement').setName('Perf Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pin_performance_impressions').setName('Perf Impressions').setType(types.NUMBER);

  // Pin posting per day (unique metric avoids routing collision with pins_count in performance group)
  fields.newMetric().setId('pins_posting_count').setName('Pins Posted Per Day').setType(types.NUMBER);

  return fields;
}

// Builds the full API URL for a Pinterest endpoint.
// p.analytics_go is injected by main.gs → getData() from Script/UserProperties.
// &pinterest_id maps config's account_id to the Pinterest account identifier.
function buildUrl_pin(endpoint, p, extra) {
  return p.analytics_go + 'pinterest/' + endpoint
    + buildBaseParams(p,
        '&pinterest_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

// Routes each Looker Studio chart to the correct Pinterest API endpoint.
// p._reqIds holds the list of field IDs the chart is requesting.
// bestMatch() scores each group by overlap and returns the winning key.
function getData_pinterest(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    topPins:     ['permalink', 'media_type', 'pin_impressions', 'pin_saves', 'pin_clicks_top',
                  'pin_outbound_clicks', 'pin_engagement', 'pin_engagement_rate',
                  'pin_id', 'pin_board_name', 'pin_embed_link', 'pin_title', 'pin_description',
                  'pin_board_owner', 'pin_cover_image_url', 'pin_dominant_color', 'pin_creative_type',
                  'pin_product_tags', 'pin_height', 'pin_width'],
    rollup:      ['total_pins', 'video_views', 'avg_watch_time',
                  'prev_total_pins', 'prev_video_views', 'prev_avg_watch_time',
                  'quartile_95s_percent_view', 'video_10s_view',
                  'prev_quartile_95s_percent_view', 'prev_video_10s_view'],
    engTrend:    ['date', 'saves_daily', 'pin_clicks_daily', 'outbound_clicks_daily', 'engagement_daily',
                  'saves_total', 'pin_clicks_total', 'outbound_clicks_total', 'engagement_total'],
    impTrend:    ['date', 'impressions_daily', 'impressions_total'],
    followers:   ['date', 'followers_daily', 'followers_gained'],
    pinPosting:  ['date', 'pins_posting_count'],
    summary:     ['period', 'follower_count', 'impressions', 'pin_clicks',
                  'outbound_clicks', 'saves', 'total_engagement',
                  'prev_follower_count', 'prev_impressions', 'prev_pin_clicks',
                  'prev_outbound_clicks', 'prev_saves', 'prev_total_engagement'],
    performance: ['date', 'pin_performance_engagement', 'pin_performance_impressions',
                  'pins_count', 'pin_clicks', 'outbound_clicks', 'saves']
  });

  switch(best) {
    case 'topPins':     return pin_fetchTopPins(p);
    case 'performance': return pin_fetchPinPerformance(p);
    case 'rollup':      return pin_fetchPinRollup(p);
    case 'engTrend':    return pin_fetchEngagementTrend(p);
    case 'impTrend':    return pin_fetchImpressionsTrend(p);
    case 'followers':   return pin_fetchFollowerTrend(p);
    case 'pinPosting':  return pin_fetchPinPosting(p);
    default:            return pin_fetchSummary(p);
  }
}

// Endpoint : GET analytics/overview/pinterest/overviewSummary
// Response : j.overview.current / j.overview.previous
// Returns one row with current and previous period totals side by side.
function pin_fetchSummary(p) {
  var j = analyticsGet(buildUrl_pin('overviewSummary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                'Current',
    follower_count:        c.follower_count   || 0,
    impressions:           c.impressions      || 0,
    pin_clicks:            c.pin_clicks       || 0,
    outbound_clicks:       c.outbound_clicks  || 0,
    saves:                 c.saves            || 0,
    total_engagement:      c.total_engagement || 0,
    prev_follower_count:   v.follower_count   || 0,
    prev_impressions:      v.impressions      || 0,
    prev_pin_clicks:       v.pin_clicks       || 0,
    prev_outbound_clicks:  v.outbound_clicks  || 0,
    prev_saves:            v.saves            || 0,
    prev_total_engagement: v.total_engagement || 0
  }];
}

// Endpoint : GET analytics/overview/pinterest/overviewFollowers
// Response : flat parallel arrays — j.buckets[], j.followers_daily[], j.followers_gained[]
function pin_fetchFollowerTrend(p) {
  var j = analyticsGet(buildUrl_pin('overviewFollowers', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:             date.replace(/-/g, ''),
      followers_daily:  (j.followers_daily  || [])[i] || 0,
      followers_gained: (j.followers_gained || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/pinterest/overviewImpressions
// Response : flat parallel arrays — j.buckets[], j.impressions_daily[], j.impressions_total[]
function pin_fetchImpressionsTrend(p) {
  var j = analyticsGet(buildUrl_pin('overviewImpressions', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:              date.replace(/-/g, ''),
      impressions_daily: (j.impressions_daily || [])[i] || 0,
      impressions_total: (j.impressions_total || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/pinterest/overviewEngagement
// Response : flat parallel arrays — j.buckets[], j.saves_daily[], j.pin_clicks_daily[],
//   j.outbound_clicks_daily[], j.engagement_daily[], j.saves_total[],
//   j.pin_clicks_total[], j.outbound_clicks_total[], j.engagement_total[]
// _daily = per-day delta; _total = cumulative running total.
function pin_fetchEngagementTrend(p) {
  var j = analyticsGet(buildUrl_pin('overviewEngagement', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:                  date.replace(/-/g, ''),
      saves_daily:           (j.saves_daily            || [])[i] || 0,
      saves_total:           (j.saves_total            || [])[i] || 0,
      pin_clicks_daily:      (j.pin_clicks_daily       || [])[i] || 0,
      pin_clicks_total:      (j.pin_clicks_total       || [])[i] || 0,
      outbound_clicks_daily: (j.outbound_clicks_daily  || [])[i] || 0,
      outbound_clicks_total: (j.outbound_clicks_total  || [])[i] || 0,
      engagement_daily:      (j.engagement_daily       || [])[i] || 0,
      engagement_total:      (j.engagement_total       || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/pinterest/overviewPinPostingRollup
// Response : j.overview.current / j.overview.previous
// Returns one row with current/previous pin publishing totals including
// video-specific metrics (quartile_95s_percent_view, video_10s_view).
function pin_fetchPinRollup(p) {
  var j = analyticsGet(buildUrl_pin('overviewPinPostingRollup', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                         'Current',
    total_pins:                     c.total_pins               || 0,
    impressions:                    c.impressions              || 0,
    pin_clicks:                     c.pin_clicks               || 0,
    outbound_clicks:                c.outbound_clicks          || 0,
    saves:                          c.saves                    || 0,
    video_views:                    c.video_views              || 0,
    avg_watch_time:                 c.avg_watch_time           || 0,
    quartile_95s_percent_view:      c.quartile_95s_percent_view|| 0,
    video_10s_view:                 c.video_10s_view           || 0,
    prev_total_pins:                v.total_pins               || 0,
    prev_video_views:               v.video_views              || 0,
    prev_avg_watch_time:            v.avg_watch_time           || 0,
    prev_quartile_95s_percent_view: v.quartile_95s_percent_view|| 0,
    prev_video_10s_view:            v.video_10s_view           || 0
  }];
}

// Endpoint : GET analytics/overview/pinterest/overviewTopPins?limit=10&order_by=impressions
// Response : j.top[] — array of pin objects
// Key aliases (API → GS):
//   pin.pin_clicks       → pin_clicks_top    (prefixed to avoid routing collision with summary pin_clicks)
//   pin.total_engagement → pin_engagement
//   pin.product_tags[]   → pin_product_tags  (joined with ',')
function pin_fetchTopPins(p) {
  var j = analyticsGet(
    buildUrl_pin('overviewTopPins', p, '&limit=10&order_by=impressions'),
    p.access_token
  );
  return (j.top || []).map(function(pin) {
    return {
      date:                  toDateStr(pin.created_at     || ''),
      media_type:            pin.media_type               || '',
      permalink:             pin.permalink                || '',
      pin_impressions:       Number(pin.impressions       || 0),
      pin_saves:             Number(pin.saves             || 0),
      pin_clicks_top:        Number(pin.pin_clicks        || 0),
      pin_outbound_clicks:   Number(pin.outbound_clicks   || 0),
      pin_engagement:        Number(pin.total_engagement  || 0),
      pin_engagement_rate:   Number(pin.engagement_rate   || 0),
      pin_height:            Number(pin.height            || 0),
      pin_width:             Number(pin.width             || 0),
      pin_id:                pin.pin_id                   || '',
      pin_board_name:        pin.board_name               || '',
      pin_embed_link:        pin.embed_link               || '',
      pin_title:             pin.title                    || '',
      pin_description:       pin.description              || '',
      pin_board_owner:       pin.board_owner              || '',
      pin_cover_image_url:   pin.cover_image_url          || '',
      pin_dominant_color:    pin.dominant_color           || '',
      pin_creative_type:     pin.creative_type            || '',
      pin_product_tags:      (pin.product_tags || []).join(',')
    };
  });
}

// Endpoint : GET analytics/overview/pinterest/overviewPinPostingPerformance
// Response : flat parallel arrays — j.buckets[], j.pins_count[], j.pin_clicks[],
//   j.outbound_clicks[], j.saves[], j.engagements[], j.impressions[]
// pin_performance_engagement / pin_performance_impressions prefixes prevent routing
// collision with the account-level impressions/engagement in the summary group.
function pin_fetchPinPerformance(p) {
  var j = analyticsGet(buildUrl_pin('overviewPinPostingPerformance', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:                        date.replace(/-/g, ''),
      pins_count:                  (j.pins_count        || [])[i] || 0,
      pin_clicks:                  (j.pin_clicks        || [])[i] || 0,
      outbound_clicks:             (j.outbound_clicks   || [])[i] || 0,
      saves:                       (j.saves             || [])[i] || 0,
      pin_performance_engagement:  (j.engagements       || [])[i] || 0,
      pin_performance_impressions: (j.impressions       || [])[i] || 0
    };
  });
}

// Endpoint : GET analytics/overview/pinterest/overviewPinPostingPerDay
// Response : flat parallel arrays — j.buckets[], j.pins_count[]
// pins_posting_count is the only field — it is the unique discriminator that
// separates this group from pinPerformance (which also has pins_count).
function pin_fetchPinPosting(p) {
  var j = analyticsGet(buildUrl_pin('overviewPinPostingPerDay', p), p.access_token);
  return (j.buckets || []).map(function(date, i) {
    return {
      date:               date.replace(/-/g, ''),
      pins_posting_count: (j.pins_count || [])[i] || 0
    };
  });
}
