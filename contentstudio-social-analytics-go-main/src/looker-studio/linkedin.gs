// ============================================================
// linkedin.gs ? LinkedIn Analytics
// ============================================================

function getFields_linkedin(fields, types) {
  // Dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);
  fields.newDimension().setId('hashtag').setName('Hashtag').setType(types.TEXT);
  fields.newDimension().setId('day_of_week').setName('Day of Week').setType(types.TEXT);

  fields.newDimension().setId('follower_demographics_category').setName('Follower Demographics Category').setType(types.TEXT);
  fields.newDimension().setId('follower_demographics_label').setName('Follower Demographics Label').setType(types.TEXT);
  fields.newDimension().setId('follower_demographics_country').setName('Follower Demographics Country').setType(types.COUNTRY);
  fields.newDimension().setId('follower_demographics_city').setName('Follower Demographics City').setType(types.CITY);
  fields.newDimension().setId('follower_demographics_industry').setName('Follower Demographics Industry').setType(types.TEXT);
  fields.newDimension().setId('follower_demographics_seniority').setName('Follower Demographics Seniority').setType(types.TEXT);

  // Summary ? current period (root: overview)
  fields.newMetric().setId('sum_curr_followers').setName('Sum Curr Followers').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_total_posts').setName('Sum Curr Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_likes').setName('Sum Curr Post Likes').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_comments').setName('Sum Curr Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_shares').setName('Sum Curr Post Shares').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_clicks').setName('Sum Curr Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_total_engagement').setName('Sum Curr Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_engagement_rate').setName('Sum Curr Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('sum_curr_post_engagement_rate').setName('Sum Curr Post Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('sum_curr_page_impressions').setName('Sum Curr Page Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_reach').setName('Sum Curr Page Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_views').setName('Sum Curr Page Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_comments').setName('Sum Curr Page Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_reactions').setName('Sum Curr Page Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_shares').setName('Sum Curr Page Shares').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_unique_visitors').setName('Sum Curr Page Unique Visitors').setType(types.NUMBER);
  // Summary ? previous period
  fields.newMetric().setId('sum_prev_followers').setName('Sum Prev Followers').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_total_posts').setName('Sum Prev Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_likes').setName('Sum Prev Post Likes').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_comments').setName('Sum Prev Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_shares').setName('Sum Prev Post Shares').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_clicks').setName('Sum Prev Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_total_engagement').setName('Sum Prev Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_engagement_rate').setName('Sum Prev Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('sum_prev_post_engagement_rate').setName('Sum Prev Post Engagement Rate').setType(types.PERCENT);
  fields.newMetric().setId('sum_prev_page_impressions').setName('Sum Prev Page Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_reach').setName('Sum Prev Page Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_views').setName('Sum Prev Page Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_comments').setName('Sum Prev Page Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_reactions').setName('Sum Prev Page Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_shares').setName('Sum Prev Page Shares').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_unique_visitors').setName('Sum Prev Page Unique Visitors').setType(types.NUMBER);

  // Audience growth ? daily (root: audience_growth)
  fields.newMetric().setId('audience_growth_total_follower_count').setName('Audience Growth Total Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_total_followers_daily').setName('Audience Growth Total Followers Daily').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_organic_follower_count').setName('Audience Growth Organic Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_organic_followers_daily').setName('Audience Growth Organic Followers Daily').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_paid_follower_count').setName('Audience Growth Paid Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_paid_followers_daily').setName('Audience Growth Paid Followers Daily').setType(types.NUMBER);
  // Audience growth rollup ? current (root: audience_growth_rollup)
  fields.newMetric().setId('audience_growth_rollup_curr_total_follower_count').setName('Audience Growth Rollup Curr Total Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_organic_follower_count').setName('Audience Growth Rollup Curr Organic Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_paid_follower_count').setName('Audience Growth Rollup Curr Paid Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_avg_follower_count').setName('Audience Growth Rollup Curr Avg Follower Count').setType(types.NUMBER);
  // Audience growth rollup ? previous
  fields.newMetric().setId('audience_growth_rollup_prev_total_follower_count').setName('Audience Growth Rollup Prev Total Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_organic_follower_count').setName('Audience Growth Rollup Prev Organic Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_paid_follower_count').setName('Audience Growth Rollup Prev Paid Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_avg_follower_count').setName('Audience Growth Rollup Prev Avg Follower Count').setType(types.NUMBER);

  // Publishing behaviour ? daily (root: publishing_behaviour)
  fields.newMetric().setId('publishing_behaviour_total_posts').setName('Publishing Behaviour Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_impressions').setName('Publishing Behaviour Impressions').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_engagement').setName('Publishing Behaviour Engagement').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_clicks').setName('Publishing Behaviour Clicks').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_likes').setName('Publishing Behaviour Likes').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_comments').setName('Publishing Behaviour Comments').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_shares').setName('Publishing Behaviour Shares').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_reach').setName('Publishing Behaviour Reach').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_engagement_rate').setName('Publishing Behaviour Engagement Rate').setType(types.PERCENT);
  // Publishing behaviour rollup ? current (root: publishing_behaviour_rollup)
  fields.newDimension().setId('pub_beh_rlu_media_type').setName('Pub Beh Rlu Media Type').setType(types.TEXT);
  fields.newMetric().setId('pub_beh_rlu_curr_total_posts').setName('Pub Beh Rlu Curr Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_likes').setName('Pub Beh Rlu Curr Likes').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_comments').setName('Pub Beh Rlu Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_shares').setName('Pub Beh Rlu Curr Shares').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_clicks').setName('Pub Beh Rlu Curr Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_engagements').setName('Pub Beh Rlu Curr Engagements').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_impressions').setName('Pub Beh Rlu Curr Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_reach').setName('Pub Beh Rlu Curr Reach').setType(types.NUMBER);
  // Publishing behaviour rollup ? previous
  fields.newMetric().setId('pub_beh_rlu_prev_total_posts').setName('Pub Beh Rlu Prev Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_likes').setName('Pub Beh Rlu Prev Likes').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_comments').setName('Pub Beh Rlu Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_shares').setName('Pub Beh Rlu Prev Shares').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_clicks').setName('Pub Beh Rlu Prev Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_engagements').setName('Pub Beh Rlu Prev Engagements').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_impressions').setName('Pub Beh Rlu Prev Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_reach').setName('Pub Beh Rlu Prev Reach').setType(types.NUMBER);

  // Page views ? daily (root: page_views)
  fields.newMetric().setId('page_views_total_page_views_daily').setName('Page Views Total Page Views Daily').setType(types.NUMBER);
  fields.newMetric().setId('page_views_desktop_page_views_daily').setName('Page Views Desktop Page Views Daily').setType(types.NUMBER);
  fields.newMetric().setId('page_views_mobile_page_views_daily').setName('Page Views Mobile Page Views Daily').setType(types.NUMBER);
  // Page views rollup ? current (root: page_views_rollup)
  fields.newMetric().setId('page_views_rollup_curr_total_page_views').setName('Page Views Rollup Curr Total Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_curr_desktop_page_views').setName('Page Views Rollup Curr Desktop Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_curr_mobile_page_views').setName('Page Views Rollup Curr Mobile Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_curr_avg_page_views').setName('Page Views Rollup Curr Avg Page Views').setType(types.NUMBER);
  // Page views rollup ? previous
  fields.newMetric().setId('page_views_rollup_prev_total_page_views').setName('Page Views Rollup Prev Total Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_prev_desktop_page_views').setName('Page Views Rollup Prev Desktop Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_prev_mobile_page_views').setName('Page Views Rollup Prev Mobile Page Views').setType(types.NUMBER);
  fields.newMetric().setId('page_views_rollup_prev_avg_page_views').setName('Page Views Rollup Prev Avg Page Views').setType(types.NUMBER);

  // Top hashtags ? list (root: top_hashtags)
  fields.newMetric().setId('top_hashtags_posts').setName('Top Hashtags Posts').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_engagements').setName('Top Hashtags Engagements').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_likes').setName('Top Hashtags Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_comments').setName('Top Hashtags Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_shares').setName('Top Hashtags Shares').setType(types.NUMBER);
  // Top hashtags rollup ? current (root: top_hashtags_rollup)
  fields.newMetric().setId('top_hashtags_rollup_curr_total_hashtags').setName('Top Hashtags Rollup Curr Total Hashtags').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_times_used').setName('Top Hashtags Rollup Curr Total Times Used').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_likes').setName('Top Hashtags Rollup Curr Total Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_comments').setName('Top Hashtags Rollup Curr Total Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_shares').setName('Top Hashtags Rollup Curr Total Shares').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_engagement').setName('Top Hashtags Rollup Curr Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_impressions').setName('Top Hashtags Rollup Curr Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_reach').setName('Top Hashtags Rollup Curr Total Reach').setType(types.NUMBER);
  // Top hashtags rollup ? previous
  fields.newMetric().setId('top_hashtags_rollup_prev_total_hashtags').setName('Top Hashtags Rollup Prev Total Hashtags').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_times_used').setName('Top Hashtags Rollup Prev Total Times Used').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_likes').setName('Top Hashtags Rollup Prev Total Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_comments').setName('Top Hashtags Rollup Prev Total Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_shares').setName('Top Hashtags Rollup Prev Total Shares').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_engagement').setName('Top Hashtags Rollup Prev Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_impressions').setName('Top Hashtags Rollup Prev Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_reach').setName('Top Hashtags Rollup Prev Total Reach').setType(types.NUMBER);

  // Top posts ? dimensions (root: top_posts)
  fields.newDimension().setId('top_posts_published_at').setName('Top Posts Published At').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_created_at').setName('Top Posts Created At').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_saving_time').setName('Top Posts Saving Time').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_linkedin_id').setName('Top Posts LinkedIn ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_post_id').setName('Top Posts Post ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_activity').setName('Top Posts Activity').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media_type').setName('Top Posts Media Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_article_url').setName('Top Posts Article URL').setType(types.URL);
  fields.newDimension().setId('top_posts_article_title').setName('Top Posts Article Title').setType(types.TEXT);
  fields.newDimension().setId('top_posts_image').setName('Top Posts Image').setType(types.URL);
  fields.newDimension().setId('top_posts_title').setName('Top Posts Title').setType(types.TEXT);
  fields.newDimension().setId('top_posts_type').setName('Top Posts Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_day_of_week').setName('Top Posts Day Of Week').setType(types.TEXT);
  fields.newDimension().setId('top_posts_poll_data').setName('Top Posts Poll Data').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media').setName('Top Posts Media').setType(types.TEXT);
  fields.newDimension().setId('top_posts_hashtags').setName('Top Posts Hashtags').setType(types.TEXT);
  // Top posts ? metrics
  fields.newMetric().setId('top_posts_hour_of_day').setName('Top Posts Hour Of Day').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_favorites').setName('Top Posts Favorites').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_comments').setName('Top Posts Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_total_engagement').setName('Top Posts Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_reach').setName('Top Posts Reach').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_repost').setName('Top Posts Repost').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_clicks').setName('Top Posts Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_impressions').setName('Top Posts Impressions').setType(types.NUMBER);

  // Posts per days (root: posts_per_days)
  fields.newMetric().setId('posts_per_days_posts').setName('Posts Per Days Posts').setType(types.NUMBER);

  // Follower demographics ? metric (root: follower_demographics)
  fields.newMetric().setId('follower_demographics_count').setName('Follower Demographics Count').setType(types.NUMBER);

  return fields;
}

function buildUrl_li(endpoint, p, extra) {
  return p.analytics_go + 'linkedin/' + endpoint
    + buildBaseParams(p,
        '&linkedin_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

function getData_linkedin(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    topPosts:              ['top_posts_published_at','top_posts_created_at','top_posts_saving_time','top_posts_linkedin_id','top_posts_post_id','top_posts_activity',
                            'top_posts_media_type','top_posts_article_url','top_posts_article_title','top_posts_image','top_posts_title','top_posts_type',
                            'top_posts_day_of_week','top_posts_hour_of_day','top_posts_poll_data','top_posts_media','top_posts_hashtags',
                            'top_posts_favorites','top_posts_comments','top_posts_total_engagement','top_posts_reach','top_posts_repost','top_posts_post_clicks','top_posts_impressions'],
    hashtags:              ['hashtag','top_hashtags_posts','top_hashtags_engagements','top_hashtags_likes','top_hashtags_comments','top_hashtags_shares'],
    hashtagsRollup:        ['top_hashtags_rollup_curr_total_hashtags','top_hashtags_rollup_curr_total_times_used','top_hashtags_rollup_curr_total_likes','top_hashtags_rollup_curr_total_comments','top_hashtags_rollup_curr_total_shares','top_hashtags_rollup_curr_total_engagement','top_hashtags_rollup_curr_total_impressions','top_hashtags_rollup_curr_total_reach',
                            'top_hashtags_rollup_prev_total_hashtags','top_hashtags_rollup_prev_total_times_used','top_hashtags_rollup_prev_total_likes','top_hashtags_rollup_prev_total_comments','top_hashtags_rollup_prev_total_shares','top_hashtags_rollup_prev_total_engagement','top_hashtags_rollup_prev_total_impressions','top_hashtags_rollup_prev_total_reach'],
    pageViews:             ['date','page_views_total_page_views_daily','page_views_desktop_page_views_daily','page_views_mobile_page_views_daily'],
    pageViewsRollup:       ['page_views_rollup_curr_total_page_views','page_views_rollup_curr_desktop_page_views','page_views_rollup_curr_mobile_page_views','page_views_rollup_curr_avg_page_views',
                            'page_views_rollup_prev_total_page_views','page_views_rollup_prev_desktop_page_views','page_views_rollup_prev_mobile_page_views','page_views_rollup_prev_avg_page_views'],
    pubRollup:             ['pub_beh_rlu_media_type','pub_beh_rlu_curr_total_posts','pub_beh_rlu_curr_likes','pub_beh_rlu_curr_comments','pub_beh_rlu_curr_shares','pub_beh_rlu_curr_clicks','pub_beh_rlu_curr_engagements','pub_beh_rlu_curr_impressions','pub_beh_rlu_curr_reach',
                            'pub_beh_rlu_prev_total_posts','pub_beh_rlu_prev_likes','pub_beh_rlu_prev_comments','pub_beh_rlu_prev_shares','pub_beh_rlu_prev_clicks','pub_beh_rlu_prev_engagements','pub_beh_rlu_prev_impressions','pub_beh_rlu_prev_reach'],
    postsPerDay:           ['day_of_week','posts_per_days_posts'],
    demographicsIndustry:  ['follower_demographics_industry','follower_demographics_count'],
    demographicsCountry:   ['follower_demographics_country','follower_demographics_count'],
    demographicsCity:      ['follower_demographics_city','follower_demographics_count'],
    demographicsSeniority: ['follower_demographics_seniority','follower_demographics_count'],
    demographics:          ['follower_demographics_category','follower_demographics_label','follower_demographics_country','follower_demographics_city','follower_demographics_industry','follower_demographics_seniority','follower_demographics_count'],
    audienceRollup:        ['audience_growth_rollup_curr_total_follower_count','audience_growth_rollup_curr_organic_follower_count','audience_growth_rollup_curr_paid_follower_count','audience_growth_rollup_curr_avg_follower_count',
                            'audience_growth_rollup_prev_total_follower_count','audience_growth_rollup_prev_organic_follower_count','audience_growth_rollup_prev_paid_follower_count','audience_growth_rollup_prev_avg_follower_count'],
    audience:              ['date','audience_growth_total_follower_count','audience_growth_total_followers_daily','audience_growth_organic_follower_count','audience_growth_organic_followers_daily','audience_growth_paid_follower_count','audience_growth_paid_followers_daily'],
    pub:                   ['date','publishing_behaviour_total_posts','publishing_behaviour_impressions','publishing_behaviour_engagement','publishing_behaviour_likes','publishing_behaviour_comments','publishing_behaviour_shares','publishing_behaviour_clicks','publishing_behaviour_reach','publishing_behaviour_engagement_rate'],
    summary:               ['period',
                            'sum_curr_followers','sum_curr_total_posts','sum_curr_post_likes','sum_curr_post_comments','sum_curr_post_shares','sum_curr_post_clicks',
                            'sum_curr_total_engagement','sum_curr_engagement_rate','sum_curr_post_engagement_rate',
                            'sum_curr_page_impressions','sum_curr_page_reach','sum_curr_page_views','sum_curr_page_comments','sum_curr_page_reactions','sum_curr_page_shares','sum_curr_page_unique_visitors',
                            'sum_prev_followers','sum_prev_total_posts','sum_prev_post_likes','sum_prev_post_comments','sum_prev_post_shares','sum_prev_post_clicks',
                            'sum_prev_total_engagement','sum_prev_engagement_rate','sum_prev_post_engagement_rate',
                            'sum_prev_page_impressions','sum_prev_page_reach','sum_prev_page_views','sum_prev_page_comments','sum_prev_page_reactions','sum_prev_page_shares','sum_prev_page_unique_visitors']
  });

  switch(best) {
    case 'topPosts':        return li_fetchTopPosts(p);
    case 'hashtags':        return li_fetchHashtags(p);
    case 'hashtagsRollup':  return li_fetchHashtagsRollup(p);
    case 'pageViews':       return li_fetchPageViews(p);
    case 'pageViewsRollup': return li_fetchPageViewsRollup(p);
    case 'pubRollup':       return li_fetchPubRollup(p);
    case 'postsPerDay':     return li_fetchPostsPerDays(p);
    case 'demographicsIndustry':  return li_fetchDemographics(p).filter(function(r){ return r.follower_demographics_industry  !== ''; });
    case 'demographicsCountry':   return li_fetchDemographics(p).filter(function(r){ return r.follower_demographics_country   !== ''; });
    case 'demographicsCity':      return li_fetchDemographics(p).filter(function(r){ return r.follower_demographics_city       !== ''; });
    case 'demographicsSeniority': return li_fetchDemographics(p).filter(function(r){ return r.follower_demographics_seniority  !== ''; });
    case 'demographics':          return li_fetchDemographics(p);
    case 'audienceRollup':  return li_fetchAudienceGrowthRollup(p);
    case 'audience':        return li_fetchAudienceGrowth(p);
    case 'pub':             return li_fetchPublishingBehaviour(p);
    default:                return li_fetchSummary(p);
  }
}

function li_fetchSummary(p) {
  var j = analyticsGet(buildUrl_li('summary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                          'Current',
    sum_curr_followers:              c.followers              || 0,
    sum_curr_total_posts:            c.total_posts            || 0,
    sum_curr_post_likes:             c.post_likes             || 0,
    sum_curr_post_comments:          c.post_comments          || 0,
    sum_curr_post_shares:            c.post_shares            || 0,
    sum_curr_post_clicks:            c.post_clicks            || 0,
    sum_curr_total_engagement:       c.total_engagement       || 0,
    sum_curr_engagement_rate:        c.engagement_rate        || 0,
    sum_curr_post_engagement_rate:   c.post_engagement_rate   || 0,
    sum_curr_page_impressions:       c.page_impressions       || 0,
    sum_curr_page_reach:             c.page_reach             || 0,
    sum_curr_page_views:             c.page_views             || 0,
    sum_curr_page_comments:          c.page_comments          || 0,
    sum_curr_page_reactions:         c.page_reactions         || 0,
    sum_curr_page_shares:            c.page_shares            || 0,
    sum_curr_page_unique_visitors:   c.page_unique_visitors   || 0,
    sum_prev_followers:              v.followers              || 0,
    sum_prev_total_posts:            v.total_posts            || 0,
    sum_prev_post_likes:             v.post_likes             || 0,
    sum_prev_post_comments:          v.post_comments          || 0,
    sum_prev_post_shares:            v.post_shares            || 0,
    sum_prev_post_clicks:            v.post_clicks            || 0,
    sum_prev_total_engagement:       v.total_engagement       || 0,
    sum_prev_engagement_rate:        v.engagement_rate        || 0,
    sum_prev_post_engagement_rate:   v.post_engagement_rate   || 0,
    sum_prev_page_impressions:       v.page_impressions       || 0,
    sum_prev_page_reach:             v.page_reach             || 0,
    sum_prev_page_views:             v.page_views             || 0,
    sum_prev_page_comments:          v.page_comments          || 0,
    sum_prev_page_reactions:         v.page_reactions         || 0,
    sum_prev_page_shares:            v.page_shares            || 0,
    sum_prev_page_unique_visitors:   v.page_unique_visitors   || 0
  }];
}

function li_fetchAudienceGrowth(p) {
  var j  = analyticsGet(buildUrl_li('audienceGrowth', p), p.access_token);
  var ag = j.audience_growth || {};
  return (ag.buckets || []).map(function(date, i) {
    return {
      date:                                    date.replace(/-/g, ''),
      audience_growth_total_follower_count:    (ag.total_follower_count    || [])[i]   || 0,
      audience_growth_total_followers_daily:   (ag.total_followers_daily   || [])[i+1] || 0,
      audience_growth_organic_follower_count:  (ag.organic_follower_count  || [])[i]   || 0,
      audience_growth_organic_followers_daily: (ag.organic_followers_daily || [])[i+1] || 0,
      audience_growth_paid_follower_count:     (ag.paid_follower_count     || [])[i]   || 0,
      audience_growth_paid_followers_daily:    (ag.paid_followers_daily    || [])[i+1] || 0
    };
  });
}

function li_fetchAudienceGrowthRollup(p) {
  var j = analyticsGet(buildUrl_li('audienceGrowth', p), p.access_token);
  var c = (j.audience_growth_rollup || {}).current  || {};
  var v = (j.audience_growth_rollup || {}).previous || {};
  return [{
    period:                                          'Current',
    audience_growth_rollup_curr_total_follower_count:   Number(c.total_follower_count   || 0),
    audience_growth_rollup_curr_organic_follower_count: Number(c.organic_follower_count || 0),
    audience_growth_rollup_curr_paid_follower_count:    Number(c.paid_follower_count    || 0),
    audience_growth_rollup_curr_avg_follower_count:     Number(c.avg_follower_count     || 0),
    audience_growth_rollup_prev_total_follower_count:   Number(v.total_follower_count   || 0),
    audience_growth_rollup_prev_organic_follower_count: Number(v.organic_follower_count || 0),
    audience_growth_rollup_prev_paid_follower_count:    Number(v.paid_follower_count    || 0),
    audience_growth_rollup_prev_avg_follower_count:     Number(v.avg_follower_count     || 0)
  }];
}

function li_fetchPublishingBehaviour(p) {
  var j  = analyticsGet(buildUrl_li('publishingBehaviour', p), p.access_token);
  var pb = j.publishing_behaviour || {};
  return (pb.buckets || []).map(function(date, i) {
    return {
      date:                                 date.replace(/-/g, ''),
      publishing_behaviour_total_posts:     (pb.total_posts     || [])[i] || 0,
      publishing_behaviour_impressions:     (pb.impressions     || [])[i] || 0,
      publishing_behaviour_engagement:      (pb.engagement      || [])[i] || 0,
      publishing_behaviour_likes:           (pb.likes           || [])[i] || 0,
      publishing_behaviour_comments:        (pb.comments        || [])[i] || 0,
      publishing_behaviour_shares:          (pb.shares          || [])[i] || 0,
      publishing_behaviour_clicks:          (pb.clicks          || [])[i] || 0,
      publishing_behaviour_reach:           (pb.reach           || [])[i] || 0,
      publishing_behaviour_engagement_rate: (pb.engagement_rate || [])[i] || 0
    };
  });
}

function li_fetchPubRollup(p) {
  var j    = analyticsGet(buildUrl_li('publishingBehaviour', p), p.access_token);
  var cur  = (j.publishing_behaviour_rollup || {}).current  || [];
  var prev = (j.publishing_behaviour_rollup || {}).previous || [];
  var prevMap = {};
  prev.forEach(function(item) { prevMap[item.media_type] = item; });
  return cur.map(function(item) {
    var v = prevMap[item.media_type] || {};
    return {
      period:                  'Current',
      pub_beh_rlu_media_type: item.media_type || '',
      pub_beh_rlu_curr_total_posts: Number(item.total_posts  || 0),
      pub_beh_rlu_curr_likes:       Number(item.likes        || 0),
      pub_beh_rlu_curr_comments:    Number(item.comments     || 0),
      pub_beh_rlu_curr_shares:      Number(item.shares       || 0),
      pub_beh_rlu_curr_clicks:      Number(item.clicks       || 0),
      pub_beh_rlu_curr_engagements: Number(item.engagements  || 0),
      pub_beh_rlu_curr_impressions: Number(item.impressions  || 0),
      pub_beh_rlu_curr_reach:       Number(item.reach        || 0),
      pub_beh_rlu_prev_total_posts: Number(v.total_posts  || 0),
      pub_beh_rlu_prev_likes:       Number(v.likes        || 0),
      pub_beh_rlu_prev_comments:    Number(v.comments     || 0),
      pub_beh_rlu_prev_shares:      Number(v.shares       || 0),
      pub_beh_rlu_prev_clicks:      Number(v.clicks       || 0),
      pub_beh_rlu_prev_engagements: Number(v.engagements  || 0),
      pub_beh_rlu_prev_impressions: Number(v.impressions  || 0),
      pub_beh_rlu_prev_reach:       Number(v.reach        || 0)
    };
  });
}

function li_fetchPageViews(p) {
  var j  = analyticsGet(buildUrl_li('pageViews', p), p.access_token);
  var pv = j.page_views || {};
  return (pv.buckets || []).map(function(date, i) {
    return {
      date:                                  date.replace(/-/g, ''),
      page_views_total_page_views_daily:     (pv.total_page_views_daily   || [])[i] || 0,
      page_views_desktop_page_views_daily:   (pv.desktop_page_views_daily || [])[i] || 0,
      page_views_mobile_page_views_daily:    (pv.mobile_page_views_daily  || [])[i] || 0
    };
  });
}

function li_fetchPageViewsRollup(p) {
  var j = analyticsGet(buildUrl_li('pageViews', p), p.access_token);
  var c = (j.page_views_rollup || {}).current  || {};
  var v = (j.page_views_rollup || {}).previous || {};
  return [{
    period:                                    'Current',
    page_views_rollup_curr_total_page_views:   Number(c.total_page_views   || 0),
    page_views_rollup_curr_desktop_page_views: Number(c.desktop_page_views || 0),
    page_views_rollup_curr_mobile_page_views:  Number(c.mobile_page_views  || 0),
    page_views_rollup_curr_avg_page_views:     Number(c.avg_page_views     || 0),
    page_views_rollup_prev_total_page_views:   Number(v.total_page_views   || 0),
    page_views_rollup_prev_desktop_page_views: Number(v.desktop_page_views || 0),
    page_views_rollup_prev_mobile_page_views:  Number(v.mobile_page_views  || 0),
    page_views_rollup_prev_avg_page_views:     Number(v.avg_page_views     || 0)
  }];
}

function li_fetchHashtags(p) {
  var j  = analyticsGet(buildUrl_li('hashtags', p), p.access_token);
  var th = j.top_hashtags || {};
  return (th.name || []).map(function(name, i) {
    return {
      hashtag:                  '#' + name,
      top_hashtags_posts:       (th.posts       || [])[i] || 0,
      top_hashtags_engagements: (th.engagements || [])[i] || 0,
      top_hashtags_likes:       (th.likes       || [])[i] || 0,
      top_hashtags_comments:    (th.comments    || [])[i] || 0,
      top_hashtags_shares:      (th.shares      || [])[i] || 0
    };
  });
}

function li_fetchHashtagsRollup(p) {
  var j = analyticsGet(buildUrl_li('hashtags', p), p.access_token);
  var c = (j.top_hashtags_rollup || {}).current  || {};
  var v = (j.top_hashtags_rollup || {}).previous || {};
  return [{
    period:                                        'Current',
    top_hashtags_rollup_curr_total_hashtags:       Number(c.total_hashtags    || 0),
    top_hashtags_rollup_curr_total_times_used:     Number(c.total_times_used  || 0),
    top_hashtags_rollup_curr_total_likes:          Number(c.total_likes       || 0),
    top_hashtags_rollup_curr_total_comments:       Number(c.total_comments    || 0),
    top_hashtags_rollup_curr_total_shares:         Number(c.total_shares      || 0),
    top_hashtags_rollup_curr_total_engagement:     Number(c.total_engagement  || 0),
    top_hashtags_rollup_curr_total_impressions:    Number(c.total_impressions || 0),
    top_hashtags_rollup_curr_total_reach:          Number(c.total_reach       || 0),
    top_hashtags_rollup_prev_total_hashtags:       Number(v.total_hashtags    || 0),
    top_hashtags_rollup_prev_total_times_used:     Number(v.total_times_used  || 0),
    top_hashtags_rollup_prev_total_likes:          Number(v.total_likes       || 0),
    top_hashtags_rollup_prev_total_comments:       Number(v.total_comments    || 0),
    top_hashtags_rollup_prev_total_shares:         Number(v.total_shares      || 0),
    top_hashtags_rollup_prev_total_engagement:     Number(v.total_engagement  || 0),
    top_hashtags_rollup_prev_total_impressions:    Number(v.total_impressions || 0),
    top_hashtags_rollup_prev_total_reach:          Number(v.total_reach       || 0)
  }];
}

// UPDATED FUNCTION WITH DEDUPLICATION
function li_fetchTopPosts(p) {
  var j = analyticsGet(
    buildUrl_li('topPosts', p, '&limit=50&order_by=total_engagement'),
    p.access_token
  );

  // Remove duplicates based on post_id
  var posts = j.top_posts || [];
  var seen = {};
  var uniquePosts = [];

  for (var i = 0; i < posts.length; i++) {
    var post = posts[i];
    var postId = post.post_id || '';

    if (!seen[postId]) {
      seen[postId] = true;
      uniquePosts.push(post);
    }
  }

  console.log('Total posts from API: ' + posts.length + ', Unique posts: ' + uniquePosts.length);

  return uniquePosts.map(function(post) {
    return {
      top_posts_published_at:     toDateStr(post.published_at || ''),
      top_posts_created_at:       toDateStr(post.created_at   || ''),
      top_posts_saving_time:      toDateStr(post.saving_time  || ''),
      top_posts_linkedin_id:      post.linkedin_id    || '',
      top_posts_post_id:          post.post_id        || '',
      top_posts_activity:         post.activity       || '',
      top_posts_media_type:       post.media_type     || '',
      top_posts_article_url:      post.article_url    || '',
      top_posts_article_title:    post.article_title  || '',
      top_posts_image:            post.image          || '',
      top_posts_title:            post.title          || '',
      top_posts_type:             post.type           || '',
      top_posts_day_of_week:      post.day_of_week    || '',
      top_posts_hour_of_day:      Number(post.hour_of_day     || 0),
      top_posts_poll_data:        post.poll_data      || '',
      top_posts_media:            (post.media    || []).join(','),
      top_posts_hashtags:         (post.hashtags || []).join(','),
      top_posts_favorites:        Number(post.favorites        || 0),
      top_posts_comments:         Number(post.comments         || 0),
      top_posts_total_engagement: Number(post.total_engagement || 0),
      top_posts_reach:            Number(post.reach            || 0),
      top_posts_repost:           Number(post.repost           || 0),
      top_posts_post_clicks:      Number(post.post_clicks      || 0),
      top_posts_impressions:      Number(post.impressions      || 0)
    };
  });
}

function li_fetchPostsPerDays(p) {
  var j       = analyticsGet(buildUrl_li('postsPerDays', p), p.access_token);
  var daysMap = ((j.posts_per_days || {}).data || {}).days || {};
  return Object.keys(daysMap).map(function(day) {
    return { day_of_week: day, posts_per_days_posts: Number(daysMap[day] || 0) };
  });
}

function li_fetchDemographics(p) {
  var j     = analyticsGet(buildUrl_li('followersDemographics', p), p.access_token);
  var demos = j.follower_demographics || {};
  var rows  = [];
  Object.keys(demos).forEach(function(category) {
    var cat = demos[category] || {};
    (cat.buckets || []).forEach(function(label, i) {
      rows.push({
        follower_demographics_category:  category,
        follower_demographics_label:     label,
        follower_demographics_country:   category === 'country'   ? label : '',
        follower_demographics_city:      category === 'city'      ? label : '',
        follower_demographics_industry:  category === 'industry'  ? label : '',
        follower_demographics_seniority: category === 'seniority' ? label : '',
        follower_demographics_count:     Number((cat.values || [])[i] || 0)
      });
    });
  });
  return rows;
}