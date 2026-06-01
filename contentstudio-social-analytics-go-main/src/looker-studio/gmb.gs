// ============================================================
// gmb.gs ? Google Business Profile (GMB) Analytics
//
// Base URL : p.analytics_go (Go service, passed from main.gs)
// Endpoints: analytics/overview/gmb/{section}
// Account  : gmb_id (mapped from config's account_id)
//
// Data sections:
//   summary        ? period, search_views, map_views, total_views,
//                    direction_requests, phone_calls, website_clicks,
//                    total_actions, photo_views, total_posts
//   impressions    ? date, search_views, map_views, direction_requests,
//                    phone_calls, website_clicks, total_actions
//   reviews        ? period, review_count, avg_rating, new_reviews
//   searchKeywords ? keyword, keyword_impressions
// ============================================================

function getFields_gmb(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('keyword').setName('Search Keyword').setType(types.TEXT);
  fields.newDimension().setId('post_name').setName('Post Name').setType(types.TEXT);
  fields.newDimension().setId('topic_type').setName('Post Type').setType(types.TEXT);

  // Business impressions (search_views/map_views/total_views map to Go's search_impressions/maps_impressions/total_impressions)
  fields.newMetric().setId('search_views').setName('Search Impressions').setType(types.NUMBER);
  fields.newMetric().setId('map_views').setName('Maps Impressions').setType(types.NUMBER);
  fields.newMetric().setId('total_views').setName('Total Impressions').setType(types.NUMBER);

  // Business actions (phone_calls maps to call_clicks)
  fields.newMetric().setId('direction_requests').setName('Direction Requests').setType(types.NUMBER);
  fields.newMetric().setId('phone_calls').setName('Call Clicks').setType(types.NUMBER);
  fields.newMetric().setId('website_clicks').setName('Website Clicks').setType(types.NUMBER);
  fields.newMetric().setId('total_actions').setName('Total Actions').setType(types.NUMBER);

  // Media activity (photo_views ? photo_count_daily)
  fields.newMetric().setId('photo_views').setName('Photo Count Daily').setType(types.NUMBER);
  fields.newMetric().setId('video_count').setName('Video Count Daily').setType(types.NUMBER);
  fields.newMetric().setId('total_posts').setName('Total Posts').setType(types.NUMBER);

  // Reviews
  fields.newMetric().setId('review_count').setName('Review Count').setType(types.NUMBER);
  fields.newMetric().setId('avg_rating').setName('Avg Rating').setType(types.NUMBER);
  fields.newMetric().setId('new_reviews').setName('New Reviews').setType(types.NUMBER);

  // Search keywords
  fields.newMetric().setId('keyword_impressions').setName('Keyword Impressions').setType(types.NUMBER);

  return fields;
}

// p.analytics_go is set by main.gs ? getData() from Script Properties
function buildUrl_gmb(endpoint, p, extra) {
  return p.analytics_go + 'gmb/' + endpoint
    + buildBaseParams(p,
        '&gmb_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

function getData_gmb(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    reviews:   ['review_count','avg_rating','new_reviews'],
    keywords:  ['keyword','keyword_impressions'],
    topPosts:  ['post_name','topic_type'],
    media:     ['photo_views','video_count'],
    actions:   ['direction_requests','phone_calls','website_clicks','total_actions'],
    daily:     ['search_views','map_views','total_views'],
    // summary before pubBehav so 'period' wins over 'total_posts' on ties
    summary:   ['period','avg_rating','review_count'],
    pubBehav:  ['total_posts']
  });

  switch(best) {
    case 'reviews':  return gmb_fetchReviews(p);
    case 'keywords': return gmb_fetchSearchKeywords(p);
    case 'topPosts': return gmb_fetchTopPosts(p);
    case 'media':    return gmb_fetchMediaActivity(p);
    case 'actions':  return gmb_fetchActions(p);
    case 'daily':    return gmb_fetchImpressions(p);
    case 'pubBehav': return gmb_fetchPublishingBehavior(p);
    default:         return gmb_fetchSummary(p);
  }
}

function gmb_fetchSummary(p) {
  var j = analyticsGet(buildUrl_gmb('summary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  function mapMetrics(m) {
    return {
      search_views:       m.search_impressions || 0,
      map_views:          m.maps_impressions   || 0,
      total_views:        m.total_impressions  || 0,
      direction_requests: m.direction_requests || 0,
      phone_calls:        m.call_clicks        || 0,
      website_clicks:     m.website_clicks     || 0,
      total_actions:      (m.call_clicks || 0) + (m.website_clicks || 0) + (m.direction_requests || 0) + (m.other_actions || 0),
      total_posts:        m.total_posts        || 0,
      review_count:       m.total_reviews      || 0,
      avg_rating:         m.average_rating     || 0,
      photo_views:        0
    };
  }
  return [
    Object.assign({ period: 'Current'  }, mapMetrics(c)),
    Object.assign({ period: 'Previous' }, mapMetrics(v))
  ];
}

function gmb_fetchImpressions(p) {
  var j  = analyticsGet(buildUrl_gmb('impressions', p), p.access_token);
  var im = j.impressions || {};
  return (im.buckets || []).map(function(date, i) {
    return {
      date:        date.replace(/-/g, ''),
      search_views: ((im.desktop_search_daily || [])[i] || 0) + ((im.mobile_search_daily || [])[i] || 0),
      map_views:    ((im.desktop_maps_daily   || [])[i] || 0) + ((im.mobile_maps_daily   || [])[i] || 0),
      total_views:  (im.total_impressions_daily || [])[i] || 0
    };
  });
}

function gmb_fetchActions(p) {
  var j  = analyticsGet(buildUrl_gmb('actions', p), p.access_token);
  var ac = j.actions || {};
  return (ac.buckets || []).map(function(date, i) {
    var calls = (ac.call_clicks         || [])[i] || 0;
    var web   = (ac.website_clicks      || [])[i] || 0;
    var dir   = (ac.direction_requests  || [])[i] || 0;
    var other = (ac.other_actions       || [])[i] || 0;
    return {
      date:               date.replace(/-/g, ''),
      phone_calls:        calls,
      website_clicks:     web,
      direction_requests: dir,
      total_actions:      calls + web + dir + other
    };
  });
}

function gmb_fetchReviews(p) {
  var j  = analyticsGet(buildUrl_gmb('reviews', p), p.access_token);
  var rc = (j.reviews_rollup || {}).current  || {};
  var rv = (j.reviews_rollup || {}).previous || {};
  return [
    { period: 'Current',  review_count: Number(rc.total_reviews || 0), avg_rating: Number(rc.avg_rating || 0), new_reviews: 0 },
    { period: 'Previous', review_count: Number(rv.total_reviews || 0), avg_rating: Number(rv.avg_rating || 0), new_reviews: 0 }
  ];
}

function gmb_fetchTopPosts(p) {
  var j = analyticsGet(buildUrl_gmb('topPosts', p), p.access_token);
  return (j.posts || []).map(function(post) {
    return {
      date:       toDateStr(post.created_at || ''),
      post_name:  post.post_name  || post.summary || '',
      topic_type: post.topic_type || ''
    };
  });
}

function gmb_fetchPublishingBehavior(p) {
  var j  = analyticsGet(buildUrl_gmb('publishingBehavior', p), p.access_token);
  var pb = j.publishing_behaviour || {};
  return (pb.buckets || []).map(function(date, i) {
    return {
      date:        date.replace(/-/g, ''),
      total_posts: (pb.post_count || [])[i] || 0
    };
  });
}

function gmb_fetchMediaActivity(p) {
  var j  = analyticsGet(buildUrl_gmb('mediaActivity', p), p.access_token);
  var ma = j.media_activity || {};
  return (ma.buckets || []).map(function(date, i) {
    return {
      date:        date.replace(/-/g, ''),
      photo_views: (ma.photo_count_daily || [])[i] || 0,
      video_count: (ma.video_count_daily || [])[i] || 0
    };
  });
}

function gmb_fetchSearchKeywords(p) {
  // SearchKeywordsResponse: {keywords: [{keyword, impressions_value, impressions_threshold, keyword_month}]}
  var j = analyticsGet(buildUrl_gmb('searchKeywords', p), p.access_token);
  return (j.keywords || []).map(function(kw) {
    return {
      keyword:             kw.keyword          || '',
      keyword_impressions: Number(kw.impressions_value || 0)
    };
  });
}