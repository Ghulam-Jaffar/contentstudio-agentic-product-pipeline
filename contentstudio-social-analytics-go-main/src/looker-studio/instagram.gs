// ============================================================
// instagram.gs ? Instagram Analytics (UPDATED WITH COMPREHENSIVE DATA)
// ============================================================

function getFields_instagram(fields, types) {
  // Shared dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);

  // Summary ? current period (root: overview)
  fields.newMetric().setId('sum_curr_total_posts').setName('Sum Curr Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_engagement').setName('Sum Curr Post Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_reactions').setName('Sum Curr Post Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_comments').setName('Sum Curr Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_saves').setName('Sum Curr Post Saves').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_reach').setName('Sum Curr Post Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_profile_impressions').setName('Sum Curr Profile Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_post_views').setName('Sum Curr Post Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_total_stories').setName('Sum Curr Total Stories').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_profile_views').setName('Sum Curr Profile Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_followers_count').setName('Sum Curr Followers Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_follows_count').setName('Sum Curr Follows Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_accounts_engaged').setName('Sum Curr Accounts Engaged').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_profile_engagement').setName('Sum Curr Profile Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_profile_reach').setName('Sum Curr Profile Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_doc_count').setName('Sum Curr Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_eng_rate').setName('Sum Curr Eng Rate').setType(types.PERCENT);
  // Summary ? previous period
  fields.newMetric().setId('sum_prev_total_posts').setName('Sum Prev Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_engagement').setName('Sum Prev Post Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_reactions').setName('Sum Prev Post Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_comments').setName('Sum Prev Post Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_saves').setName('Sum Prev Post Saves').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_reach').setName('Sum Prev Post Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_profile_impressions').setName('Sum Prev Profile Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_post_views').setName('Sum Prev Post Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_total_stories').setName('Sum Prev Total Stories').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_profile_views').setName('Sum Prev Profile Views').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_followers_count').setName('Sum Prev Followers Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_follows_count').setName('Sum Prev Follows Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_accounts_engaged').setName('Sum Prev Accounts Engaged').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_profile_engagement').setName('Sum Prev Profile Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_profile_reach').setName('Sum Prev Profile Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_doc_count').setName('Sum Prev Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_eng_rate').setName('Sum Prev Eng Rate').setType(types.PERCENT);

  // Audience growth ? daily (root: audience_growth)
  fields.newMetric().setId('audience_growth_followers').setName('Audience Growth Followers').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_followers_daily').setName('Audience Growth Followers Daily').setType(types.NUMBER);
  // Audience growth rollup (root: audience_growth_rollup)
  fields.newMetric().setId('audience_growth_rollup_curr_follower_count').setName('Audience Growth Rollup Curr Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_follower_gained').setName('Audience Growth Rollup Curr Follower Gained').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_follower_count').setName('Audience Growth Rollup Prev Follower Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_follower_gained').setName('Audience Growth Rollup Prev Follower Gained').setType(types.NUMBER);

  // Publishing behaviour ? daily (root: publishing_behaviour)
  fields.newMetric().setId('publishing_behaviour_total_posts').setName('Publishing Behaviour Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_likes').setName('Publishing Behaviour Likes').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_comments').setName('Publishing Behaviour Comments').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_saved').setName('Publishing Behaviour Saved').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_engagement').setName('Publishing Behaviour Engagement').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_reach').setName('Publishing Behaviour Reach').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_impressions').setName('Publishing Behaviour Impressions').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_views').setName('Publishing Behaviour Views').setType(types.NUMBER);
  // Publishing behaviour rollup (root: publishing_behaviour_rollup)
  fields.newDimension().setId('pub_beh_rlu_media_type').setName('Pub Beh Rlu Media Type').setType(types.TEXT);
  fields.newMetric().setId('pub_beh_rlu_curr_total_posts').setName('Pub Beh Rlu Curr Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_likes').setName('Pub Beh Rlu Curr Likes').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_comments').setName('Pub Beh Rlu Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_saved').setName('Pub Beh Rlu Curr Saved').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_engagement').setName('Pub Beh Rlu Curr Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_reach').setName('Pub Beh Rlu Curr Reach').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_views').setName('Pub Beh Rlu Curr Views').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_total_posts').setName('Pub Beh Rlu Prev Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_likes').setName('Pub Beh Rlu Prev Likes').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_comments').setName('Pub Beh Rlu Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_saved').setName('Pub Beh Rlu Prev Saved').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_engagement').setName('Pub Beh Rlu Prev Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_reach').setName('Pub Beh Rlu Prev Reach').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_views').setName('Pub Beh Rlu Prev Views').setType(types.NUMBER);

  // Stories performance ? daily (root: stories_performance)
  fields.newMetric().setId('stories_performance_published_stories').setName('Stories Performance Published Stories').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_impressions').setName('Stories Performance Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_avg_story_impressions').setName('Stories Performance Avg Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_reach').setName('Stories Performance Story Reach').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_reply').setName('Stories Performance Story Reply').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_exits').setName('Stories Performance Story Exits').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_taps_forward').setName('Stories Performance Story Taps Forward').setType(types.NUMBER);
  fields.newMetric().setId('stories_performance_story_taps_back').setName('Stories Performance Story Taps Back').setType(types.NUMBER);
  // Stories rollup (root: stories_rollup)
  fields.newMetric().setId('stories_rollup_curr_published_stories').setName('Stories Rollup Curr Published Stories').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_impressions').setName('Stories Rollup Curr Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_avg_story_impressions').setName('Stories Rollup Curr Avg Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_reach').setName('Stories Rollup Curr Story Reach').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_reply').setName('Stories Rollup Curr Story Reply').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_exits').setName('Stories Rollup Curr Story Exits').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_taps_forward').setName('Stories Rollup Curr Story Taps Forward').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_curr_story_taps_back').setName('Stories Rollup Curr Story Taps Back').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_published_stories').setName('Stories Rollup Prev Published Stories').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_impressions').setName('Stories Rollup Prev Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_avg_story_impressions').setName('Stories Rollup Prev Avg Story Impressions').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_reach').setName('Stories Rollup Prev Story Reach').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_reply').setName('Stories Rollup Prev Story Reply').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_exits').setName('Stories Rollup Prev Story Exits').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_taps_forward').setName('Stories Rollup Prev Story Taps Forward').setType(types.NUMBER);
  fields.newMetric().setId('stories_rollup_prev_story_taps_back').setName('Stories Rollup Prev Story Taps Back').setType(types.NUMBER);

  // Reels ? daily (root: reels)
  fields.newMetric().setId('reels_total_posts').setName('Reels Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('reels_engagement').setName('Reels Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_likes').setName('Reels Likes').setType(types.NUMBER);
  fields.newMetric().setId('reels_comments').setName('Reels Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_saves').setName('Reels Saves').setType(types.NUMBER);
  fields.newMetric().setId('reels_shares').setName('Reels Shares').setType(types.NUMBER);
  fields.newMetric().setId('reels_avg_watch_time').setName('Reels Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('reels_total_watch_time').setName('Reels Total Watch Time').setType(types.NUMBER);
  // Reels rollup (root: reels_rollup)
  fields.newMetric().setId('reels_rollup_curr_engagement').setName('Reels Rollup Curr Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_likes').setName('Reels Rollup Curr Likes').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_comments').setName('Reels Rollup Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_saves').setName('Reels Rollup Curr Saves').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_total_posts').setName('Reels Rollup Curr Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_shares').setName('Reels Rollup Curr Shares').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_avg_watch_time').setName('Reels Rollup Curr Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_total_watch_time').setName('Reels Rollup Curr Total Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_engagement').setName('Reels Rollup Prev Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_likes').setName('Reels Rollup Prev Likes').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_comments').setName('Reels Rollup Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_saves').setName('Reels Rollup Prev Saves').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_total_posts').setName('Reels Rollup Prev Total Posts').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_shares').setName('Reels Rollup Prev Shares').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_avg_watch_time').setName('Reels Rollup Prev Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_total_watch_time').setName('Reels Rollup Prev Total Watch Time').setType(types.NUMBER);

  // Top hashtags ? list (root: top_hashtags)
  fields.newDimension().setId('top_hashtags_name').setName('Top Hashtags Name').setType(types.TEXT);
  fields.newMetric().setId('top_hashtags_posts').setName('Top Hashtags Posts').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_engagement').setName('Top Hashtags Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_likes').setName('Top Hashtags Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_comments').setName('Top Hashtags Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_saved').setName('Top Hashtags Saved').setType(types.NUMBER);
  // Top hashtags rollup (root: top_hashtags_rollup)
  fields.newMetric().setId('top_hashtags_rollup_curr_total_engagement').setName('Top Hashtags Rollup Curr Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_likes').setName('Top Hashtags Rollup Curr Total Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_comments').setName('Top Hashtags Rollup Curr Total Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_saves').setName('Top Hashtags Rollup Curr Total Saves').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_unique_hashtags').setName('Top Hashtags Rollup Curr Total Unique Hashtags').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_curr_total_hashtag_uses').setName('Top Hashtags Rollup Curr Total Hashtag Uses').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_engagement').setName('Top Hashtags Rollup Prev Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_likes').setName('Top Hashtags Rollup Prev Total Likes').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_comments').setName('Top Hashtags Rollup Prev Total Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_saves').setName('Top Hashtags Rollup Prev Total Saves').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_unique_hashtags').setName('Top Hashtags Rollup Prev Total Unique Hashtags').setType(types.NUMBER);
  fields.newMetric().setId('top_hashtags_rollup_prev_total_hashtag_uses').setName('Top Hashtags Rollup Prev Total Hashtag Uses').setType(types.NUMBER);

  // Impressions ? daily (root: impressions)
  fields.newMetric().setId('impressions_impressions').setName('Impressions Impressions').setType(types.NUMBER);
  // Impressions rollup (root: impressions_rollup)
  fields.newMetric().setId('impressions_rollup_curr_total_impressions').setName('Impressions Rollup Curr Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_curr_avg_impressions').setName('Impressions Rollup Curr Avg Impressions').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_prev_total_impressions').setName('Impressions Rollup Prev Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_prev_avg_impressions').setName('Impressions Rollup Prev Avg Impressions').setType(types.NUMBER);

  // Engagements ? daily (root: engagements)
  fields.newMetric().setId('engagements_engagement').setName('Engagements Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagements_comments').setName('Engagements Comments').setType(types.NUMBER);
  fields.newMetric().setId('engagements_reactions').setName('Engagements Reactions').setType(types.NUMBER);
  // Engagements rollup (root: engagements_rollup)
  fields.newMetric().setId('engagements_rollup_curr_engagement').setName('Engagements Rollup Curr Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_curr_avg_engagement').setName('Engagements Rollup Curr Avg Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_curr_comments').setName('Engagements Rollup Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_curr_reactions').setName('Engagements Rollup Curr Reactions').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_curr_saved').setName('Engagements Rollup Curr Saved').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_curr_count').setName('Engagements Rollup Curr Count').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_engagement').setName('Engagements Rollup Prev Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_avg_engagement').setName('Engagements Rollup Prev Avg Engagement').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_comments').setName('Engagements Rollup Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_reactions').setName('Engagements Rollup Prev Reactions').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_saved').setName('Engagements Rollup Prev Saved').setType(types.NUMBER);
  fields.newMetric().setId('engagements_rollup_prev_count').setName('Engagements Rollup Prev Count').setType(types.NUMBER);

  // Active users ? days (root: active_users_days)
  fields.newDimension().setId('active_users_days_buckets').setName('Active Users Days Buckets').setType(types.TEXT);
  fields.newMetric().setId('active_users_days_values').setName('Active Users Days Values').setType(types.NUMBER);
  fields.newMetric().setId('active_users_days_highest_value').setName('Active Users Days Highest Value').setType(types.NUMBER);
  fields.newDimension().setId('active_users_days_highest_day').setName('Active Users Days Highest Day').setType(types.TEXT);
  // Active users ? hours (root: active_users_hours)
  fields.newMetric().setId('active_users_hours_buckets').setName('Active Users Hours Buckets').setType(types.NUMBER);
  fields.newMetric().setId('active_users_hours_values').setName('Active Users Hours Values').setType(types.NUMBER);
  fields.newMetric().setId('active_users_hours_highest_value').setName('Active Users Hours Highest Value').setType(types.NUMBER);
  fields.newMetric().setId('active_users_hours_highest_hour').setName('Active Users Hours Highest Hour').setType(types.NUMBER);

  // Top posts ? dimensions (root: top_posts)
  fields.newDimension().setId('top_posts_post_created_at').setName('Top Posts Post Created At').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_stored_event_at').setName('Top Posts Stored Event At').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_instagram_id').setName('Top Posts Instagram ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media_id').setName('Top Posts Media ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_caption').setName('Top Posts Caption').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media_type').setName('Top Posts Media Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_entity_type').setName('Top Posts Entity Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media_url').setName('Top Posts Media URL').setType(types.TEXT);
  fields.newDimension().setId('top_posts_video_url').setName('Top Posts Video URL').setType(types.TEXT);
  fields.newDimension().setId('top_posts_permalink').setName('Top Posts Permalink').setType(types.URL);
  fields.newDimension().setId('top_posts_hashtags').setName('Top Posts Hashtags').setType(types.TEXT);
  fields.newDimension().setId('top_posts_day_of_week').setName('Top Posts Day Of Week').setType(types.TEXT);
  // Top posts ? metrics
  fields.newMetric().setId('top_posts_hour_of_day').setName('Top Posts Hour Of Day').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_like_count').setName('Top Posts Like Count').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_comments_count').setName('Top Posts Comments Count').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_saved').setName('Top Posts Saved').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_engagement').setName('Top Posts Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_reach').setName('Top Posts Reach').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_impressions').setName('Top Posts Impressions').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_views').setName('Top Posts Views').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_shares').setName('Top Posts Shares').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_reels_avg_watch_time').setName('Top Posts Reels Avg Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_reels_total_watch_time').setName('Top Posts Reels Total Watch Time').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_exits').setName('Top Posts Exits').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_replies').setName('Top Posts Replies').setType(types.NUMBER);

  // Demographics ? age (root: audience_age)
  fields.newDimension().setId('audience_age_bracket').setName('Audience Age Bracket').setType(types.TEXT);
  fields.newMetric().setId('audience_age_count').setName('Audience Age Count').setType(types.NUMBER);

  // Demographics ? gender (root: audience_gender)
  fields.newDimension().setId('audience_gender_gender').setName('Audience Gender Gender').setType(types.TEXT);
  fields.newMetric().setId('audience_gender_count').setName('Audience Gender Count').setType(types.NUMBER);

  // Audience location ? city (root: audience_city)
  fields.newDimension().setId('audience_city_city').setName('Audience City City').setType(types.CITY);
  fields.newMetric().setId('audience_city_count').setName('Audience City Count').setType(types.NUMBER);

  // Audience location ? country (root: audience_country)
  fields.newDimension().setId('audience_country_country').setName('Audience Country Country').setType(types.COUNTRY);
  fields.newMetric().setId('audience_country_count').setName('Audience Country Count').setType(types.NUMBER);

  return fields;
}

function buildUrl_ig(endpoint, p, extra) {
  return p.analytics_go + 'instagram/' + endpoint
    + buildBaseParams(p,
        '&instagram_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

function getData_instagram(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    igLocCity:       ['audience_city_city','audience_city_count'],
    igLocCountry:    ['audience_country_country','audience_country_count'],
    ageDemo:         ['audience_age_bracket','audience_age_count'],
    genderDemo:      ['audience_gender_gender','audience_gender_count'],
    activeHours:     ['active_users_hours_buckets','active_users_hours_values','active_users_hours_highest_value','active_users_hours_highest_hour'],
    activeDays:      ['active_users_days_buckets','active_users_days_values','active_users_days_highest_value','active_users_days_highest_day'],
    topPosts:        ['top_posts_post_created_at','top_posts_stored_event_at','top_posts_instagram_id','top_posts_media_id','top_posts_caption',
                      'top_posts_media_type','top_posts_entity_type','top_posts_media_url','top_posts_video_url','top_posts_permalink',
                      'top_posts_hashtags','top_posts_day_of_week','top_posts_hour_of_day',
                      'top_posts_like_count','top_posts_comments_count','top_posts_saved','top_posts_engagement','top_posts_reach',
                      'top_posts_impressions','top_posts_views','top_posts_shares',
                      'top_posts_reels_avg_watch_time','top_posts_reels_total_watch_time','top_posts_exits','top_posts_replies'],
    hashtagsRollup:  ['top_hashtags_rollup_curr_total_engagement','top_hashtags_rollup_curr_total_likes','top_hashtags_rollup_curr_total_comments','top_hashtags_rollup_curr_total_saves','top_hashtags_rollup_curr_total_unique_hashtags','top_hashtags_rollup_curr_total_hashtag_uses',
                      'top_hashtags_rollup_prev_total_engagement','top_hashtags_rollup_prev_total_likes','top_hashtags_rollup_prev_total_comments','top_hashtags_rollup_prev_total_saves','top_hashtags_rollup_prev_total_unique_hashtags','top_hashtags_rollup_prev_total_hashtag_uses'],
    hashtags:        ['top_hashtags_name','top_hashtags_posts','top_hashtags_engagement','top_hashtags_likes','top_hashtags_comments','top_hashtags_saved'],
    storiesRollup:   ['stories_rollup_curr_published_stories','stories_rollup_curr_story_impressions','stories_rollup_curr_avg_story_impressions','stories_rollup_curr_story_reach','stories_rollup_curr_story_reply','stories_rollup_curr_story_exits','stories_rollup_curr_story_taps_forward','stories_rollup_curr_story_taps_back',
                      'stories_rollup_prev_published_stories','stories_rollup_prev_story_impressions','stories_rollup_prev_avg_story_impressions','stories_rollup_prev_story_reach','stories_rollup_prev_story_reply','stories_rollup_prev_story_exits','stories_rollup_prev_story_taps_forward','stories_rollup_prev_story_taps_back'],
    reelsRollup:     ['reels_rollup_curr_engagement','reels_rollup_curr_likes','reels_rollup_curr_comments','reels_rollup_curr_saves','reels_rollup_curr_total_posts','reels_rollup_curr_shares','reels_rollup_curr_avg_watch_time','reels_rollup_curr_total_watch_time',
                      'reels_rollup_prev_engagement','reels_rollup_prev_likes','reels_rollup_prev_comments','reels_rollup_prev_saves','reels_rollup_prev_total_posts','reels_rollup_prev_shares','reels_rollup_prev_avg_watch_time','reels_rollup_prev_total_watch_time'],
    audienceRollup:  ['audience_growth_rollup_curr_follower_count','audience_growth_rollup_curr_follower_gained','audience_growth_rollup_prev_follower_count','audience_growth_rollup_prev_follower_gained'],
    impRollup:       ['impressions_rollup_curr_total_impressions','impressions_rollup_curr_avg_impressions','impressions_rollup_prev_total_impressions','impressions_rollup_prev_avg_impressions'],
    engRollup:       ['engagements_rollup_curr_engagement','engagements_rollup_curr_avg_engagement','engagements_rollup_curr_comments','engagements_rollup_curr_reactions','engagements_rollup_curr_saved','engagements_rollup_curr_count',
                      'engagements_rollup_prev_engagement','engagements_rollup_prev_avg_engagement','engagements_rollup_prev_comments','engagements_rollup_prev_reactions','engagements_rollup_prev_saved','engagements_rollup_prev_count'],
    pubRollup:       ['pub_beh_rlu_media_type','pub_beh_rlu_curr_total_posts','pub_beh_rlu_curr_likes','pub_beh_rlu_curr_comments','pub_beh_rlu_curr_saved','pub_beh_rlu_curr_engagement','pub_beh_rlu_curr_reach','pub_beh_rlu_curr_views',
                      'pub_beh_rlu_prev_total_posts','pub_beh_rlu_prev_likes','pub_beh_rlu_prev_comments','pub_beh_rlu_prev_saved','pub_beh_rlu_prev_engagement','pub_beh_rlu_prev_reach','pub_beh_rlu_prev_views'],
    stories:         ['date','stories_performance_published_stories','stories_performance_story_impressions','stories_performance_avg_story_impressions','stories_performance_story_reach','stories_performance_story_reply','stories_performance_story_exits','stories_performance_story_taps_forward','stories_performance_story_taps_back'],
    reels:           ['date','reels_total_posts','reels_engagement','reels_likes','reels_comments','reels_saves','reels_shares','reels_avg_watch_time','reels_total_watch_time'],
    impTrend:        ['date','impressions_impressions'],
    engTrend:        ['date','engagements_engagement','engagements_comments','engagements_reactions'],
    audience:        ['date','audience_growth_followers','audience_growth_followers_daily'],
    pub:             ['date','publishing_behaviour_total_posts','publishing_behaviour_likes','publishing_behaviour_comments','publishing_behaviour_saved','publishing_behaviour_engagement','publishing_behaviour_reach','publishing_behaviour_impressions','publishing_behaviour_views'],
    summary:         ['period',
                      'sum_curr_total_posts','sum_curr_post_engagement','sum_curr_post_reactions','sum_curr_post_comments','sum_curr_post_saves','sum_curr_post_reach',
                      'sum_curr_profile_impressions','sum_curr_post_views','sum_curr_total_stories','sum_curr_profile_views','sum_curr_followers_count','sum_curr_follows_count',
                      'sum_curr_accounts_engaged','sum_curr_profile_engagement','sum_curr_profile_reach','sum_curr_doc_count','sum_curr_eng_rate',
                      'sum_prev_total_posts','sum_prev_post_engagement','sum_prev_post_reactions','sum_prev_post_comments','sum_prev_post_saves',
                      'sum_prev_post_reach','sum_prev_profile_impressions','sum_prev_post_views','sum_prev_total_stories','sum_prev_profile_views',
                      'sum_prev_followers_count','sum_prev_follows_count','sum_prev_accounts_engaged','sum_prev_profile_engagement','sum_prev_profile_reach',
                      'sum_prev_doc_count','sum_prev_eng_rate']
  });

  switch(best) {
    case 'igLocCity':      return ig_fetchAudienceLocation(p, 'city');
    case 'igLocCountry':   return ig_fetchAudienceLocation(p, 'country');
    case 'ageDemo':        return ig_fetchDemographics(p, 'age');
    case 'genderDemo':     return ig_fetchDemographics(p, 'gender');
    case 'activeHours':    return ig_fetchActiveHours(p);
    case 'activeDays':     return ig_fetchActiveDays(p);
    case 'topPosts':       return ig_fetchTopPosts(p);
    case 'hashtagsRollup': return ig_fetchHashtagsRollup(p);
    case 'hashtags':       return ig_fetchHashtags(p);
    case 'storiesRollup':  return ig_fetchStoriesRollup(p);
    case 'reelsRollup':    return ig_fetchReelsRollup(p);
    case 'audienceRollup': return ig_fetchAudienceRollup(p);
    case 'impRollup':      return ig_fetchImpressionsRollup(p);
    case 'engRollup':      return ig_fetchEngagementRollup(p);
    case 'pubRollup':      return ig_fetchPubRollup(p);
    case 'stories':        return ig_fetchStories(p);
    case 'reels':          return ig_fetchReels(p);
    case 'impTrend':       return ig_fetchImpressionsTrend(p);
    case 'engTrend':       return ig_fetchEngagementTrend(p);
    case 'audience':       return ig_fetchAudienceGrowth(p);
    case 'pub':            return ig_fetchPubDaily(p);
    default:               return ig_fetchSummary(p);
  }
}

function ig_fetchSummary(p) {
  var j = analyticsGet(buildUrl_ig('summary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                       'Current',
    sum_curr_total_posts:         Number(c.total_posts         || 0),
    sum_curr_post_engagement:     Number(c.post_engagement     || 0),
    sum_curr_post_reactions:      Number(c.post_reactions      || 0),
    sum_curr_post_comments:       Number(c.post_comments       || 0),
    sum_curr_post_saves:          Number(c.post_saves          || 0),
    sum_curr_post_reach:          Number(c.post_reach          || 0),
    sum_curr_profile_impressions: Number(c.profile_impressions || 0),
    sum_curr_post_views:          Number(c.post_views          || 0),
    sum_curr_total_stories:       Number(c.total_stories       || 0),
    sum_curr_profile_views:       Number(c.profile_views       || 0),
    sum_curr_followers_count:     Number(c.followers_count     || 0),
    sum_curr_follows_count:       Number(c.follows_count       || 0),
    sum_curr_accounts_engaged:    Number(c.accounts_engaged    || 0),
    sum_curr_profile_engagement:  Number(c.profile_engagement  || 0),
    sum_curr_profile_reach:       Number(c.profile_reach       || 0),
    sum_curr_doc_count:           Number(c.doc_count           || 0),
    sum_curr_eng_rate:            Number(c.eng_rate            || 0),
    sum_prev_total_posts:         Number(v.total_posts         || 0),
    sum_prev_post_engagement:     Number(v.post_engagement     || 0),
    sum_prev_post_reactions:      Number(v.post_reactions      || 0),
    sum_prev_post_comments:       Number(v.post_comments       || 0),
    sum_prev_post_saves:          Number(v.post_saves          || 0),
    sum_prev_post_reach:          Number(v.post_reach          || 0),
    sum_prev_profile_impressions: Number(v.profile_impressions || 0),
    sum_prev_post_views:          Number(v.post_views          || 0),
    sum_prev_total_stories:       Number(v.total_stories       || 0),
    sum_prev_profile_views:       Number(v.profile_views       || 0),
    sum_prev_followers_count:     Number(v.followers_count     || 0),
    sum_prev_follows_count:       Number(v.follows_count       || 0),
    sum_prev_accounts_engaged:    Number(v.accounts_engaged    || 0),
    sum_prev_profile_engagement:  Number(v.profile_engagement  || 0),
    sum_prev_profile_reach:       Number(v.profile_reach       || 0),
    sum_prev_doc_count:           Number(v.doc_count           || 0),
    sum_prev_eng_rate:            Number(v.eng_rate            || 0)
  }];
}

function ig_fetchAudienceGrowth(p) {
  var j  = analyticsGet(buildUrl_ig('audienceGrowth', p), p.access_token);
  var ag = j.audience_growth || {};
  return (ag.buckets || []).reduce(function(acc, date, i) {
    var f = (ag.followers || [])[i] || 0;
    if (f > 0) acc.push({
      date:                           date.replace(/-/g, ''),
      audience_growth_followers:      f,
      audience_growth_followers_daily:(ag.followers_daily || [])[i] || 0
    });
    return acc;
  }, []);
}

function ig_fetchAudienceRollup(p) {
  var j = analyticsGet(buildUrl_ig('audienceGrowth', p), p.access_token);
  var c = (j.audience_growth_rollup || {}).current  || {};
  var v = (j.audience_growth_rollup || {}).previous || {};
  return [{
    period:                                     'Current',
    audience_growth_rollup_curr_follower_count:  Number(c.follower_count  || 0),
    audience_growth_rollup_curr_follower_gained: Number(c.follower_gained || 0),
    audience_growth_rollup_prev_follower_count:  Number(v.follower_count  || 0),
    audience_growth_rollup_prev_follower_gained: Number(v.follower_gained || 0)
  }];
}

// ============================================================
// UPDATED FUNCTION WITH DEBUG LOGGING
// ============================================================
function ig_fetchPubDaily(p) {
  var url = buildUrl_ig('publishingBehaviour', p);
  console.log('=== INSTAGRAM API DEBUG ===');
  console.log('API URL:', url);

  var j  = analyticsGet(url, p.access_token);

  console.log('API Response Status:', j.status || 'unknown');

  var pb = j.publishing_behaviour || {};

  // Debug: Log the actual arrays from API
  console.log('Likes array (first 10):', JSON.stringify((pb.likes || []).slice(0, 10)));
  console.log('Comments array (first 10):', JSON.stringify((pb.comments || []).slice(0, 10)));
  console.log('Engagement array (first 10):', JSON.stringify((pb.engagement || []).slice(0, 10)));
  console.log('Total Posts array (first 10):', JSON.stringify((pb.total_posts || []).slice(0, 10)));
  console.log('Buckets array (first 5):', JSON.stringify((pb.buckets || []).slice(0, 5)));

  // Log array lengths
  console.log('Array lengths - Likes:', (pb.likes || []).length, 'Comments:', (pb.comments || []).length, 'Engagement:', (pb.engagement || []).length);

  // Check for Feb 6, 2026 (index 18 according to API response)
  console.log('Feb 6 (2026-02-06) data check:');
  console.log('  Date:', (pb.buckets || [])[18]);
  console.log('  Likes:', (pb.likes || [])[18]);
  console.log('  Comments:', (pb.comments || [])[18]);
  console.log('  Engagement:', (pb.engagement || [])[18]);
  console.log('  Total Posts:', (pb.total_posts || [])[18]);

  var result = (pb.buckets || []).map(function(date, i) {
    return {
      date:                              date.replace(/-/g, ''),
      publishing_behaviour_total_posts:  (pb.total_posts  || [])[i] || 0,
      publishing_behaviour_likes:        (pb.likes        || [])[i] || 0,
      publishing_behaviour_comments:     (pb.comments     || [])[i] || 0,
      publishing_behaviour_saved:        (pb.saved        || [])[i] || 0,
      publishing_behaviour_engagement:   (pb.engagement   || [])[i] || 0,
      publishing_behaviour_reach:        (pb.reach        || [])[i] || 0,
      publishing_behaviour_impressions:  (pb.impressions  || [])[i] || 0,
      publishing_behaviour_views:        (pb.views        || [])[i] || 0
    };
  });

  console.log('Processed result count:', result.length);
  console.log('Sample processed data (first 3):');
  result.slice(0, 3).forEach(function(row, i) {
    console.log('  Row', i, ':', JSON.stringify(row));
  });

  console.log('=== END INSTAGRAM DEBUG ===');

  return result;
}

function ig_fetchPubRollup(p) {
  var j    = analyticsGet(
    buildUrl_ig('publishingBehaviour', p, '&media_type=IMAGE,VIDEO,CAROUSEL_ALBUM,REELS'),
    p.access_token
  );
  var cur  = (j.publishing_behaviour_rollup || {}).current  || [];
  var prev = (j.publishing_behaviour_rollup || {}).previous || [];
  var prevMap = {};
  prev.forEach(function(item) { prevMap[item.media_type] = item; });
  return cur.map(function(item) {
    var v = prevMap[item.media_type] || {};
    return {
      pub_beh_rlu_media_type:      item.media_type          || '',
      pub_beh_rlu_curr_total_posts:Number(item.total_posts  || 0),
      pub_beh_rlu_curr_likes:      Number(item.likes        || 0),
      pub_beh_rlu_curr_comments:   Number(item.comments     || 0),
      pub_beh_rlu_curr_saved:      Number(item.saved        || 0),
      pub_beh_rlu_curr_engagement: Number(item.engagement   || 0),
      pub_beh_rlu_curr_reach:      Number(item.reach        || 0),
      pub_beh_rlu_curr_views:      Number(item.views        || 0),
      pub_beh_rlu_prev_total_posts:Number(v.total_posts     || 0),
      pub_beh_rlu_prev_likes:      Number(v.likes           || 0),
      pub_beh_rlu_prev_comments:   Number(v.comments        || 0),
      pub_beh_rlu_prev_saved:      Number(v.saved           || 0),
      pub_beh_rlu_prev_engagement: Number(v.engagement      || 0),
      pub_beh_rlu_prev_reach:      Number(v.reach           || 0),
      pub_beh_rlu_prev_views:      Number(v.views           || 0)
    };
  });
}

function ig_fetchStories(p) {
  var j  = analyticsGet(buildUrl_ig('storiesPerformance', p), p.access_token);
  var sp = j.stories_performance || {};
  return (sp.buckets || []).map(function(date, i) {
    return {
      date:                                      date.replace(/-/g, ''),
      stories_performance_published_stories:     (sp.published_stories     || [])[i] || 0,
      stories_performance_story_impressions:     (sp.story_impressions     || [])[i] || 0,
      stories_performance_avg_story_impressions: (sp.avg_story_impressions || [])[i] || 0,
      stories_performance_story_reach:           (sp.story_reach           || [])[i] || 0,
      stories_performance_story_reply:           (sp.story_reply           || [])[i] || 0,
      stories_performance_story_exits:           (sp.story_exits           || [])[i] || 0,
      stories_performance_story_taps_forward:    (sp.story_taps_forward    || [])[i] || 0,
      stories_performance_story_taps_back:       (sp.story_taps_back       || [])[i] || 0
    };
  });
}

function ig_fetchStoriesRollup(p) {
  var j = analyticsGet(buildUrl_ig('storiesPerformance', p), p.access_token);
  var c = (j.stories_rollup || {}).current  || {};
  var v = (j.stories_rollup || {}).previous || {};
  return [{
    period:                                        'Current',
    stories_rollup_curr_published_stories:         Number(c.published_stories     || 0),
    stories_rollup_curr_story_impressions:         Number(c.story_impressions     || 0),
    stories_rollup_curr_avg_story_impressions:     Number(c.avg_story_impressions || 0),
    stories_rollup_curr_story_reach:               Number(c.story_reach           || 0),
    stories_rollup_curr_story_reply:               Number(c.story_reply           || 0),
    stories_rollup_curr_story_exits:               Number(c.story_exits           || 0),
    stories_rollup_curr_story_taps_forward:        Number(c.story_taps_forward    || 0),
    stories_rollup_curr_story_taps_back:           Number(c.story_taps_back       || 0),
    stories_rollup_prev_published_stories:         Number(v.published_stories     || 0),
    stories_rollup_prev_story_impressions:         Number(v.story_impressions     || 0),
    stories_rollup_prev_avg_story_impressions:     Number(v.avg_story_impressions || 0),
    stories_rollup_prev_story_reach:               Number(v.story_reach           || 0),
    stories_rollup_prev_story_reply:               Number(v.story_reply           || 0),
    stories_rollup_prev_story_exits:               Number(v.story_exits           || 0),
    stories_rollup_prev_story_taps_forward:        Number(v.story_taps_forward    || 0),
    stories_rollup_prev_story_taps_back:           Number(v.story_taps_back       || 0)
  }];
}

function ig_fetchReels(p) {
  var j = analyticsGet(buildUrl_ig('reelsPerformance', p), p.access_token);
  var r = j.reels || {};
  return (r.buckets || []).map(function(date, i) {
    return {
      date:                    date.replace(/-/g, ''),
      reels_total_posts:       (r.total_posts      || [])[i] || 0,
      reels_engagement:        (r.engagement       || [])[i] || 0,
      reels_likes:             (r.likes            || [])[i] || 0,
      reels_comments:          (r.comments         || [])[i] || 0,
      reels_saves:             (r.saves            || [])[i] || 0,
      reels_shares:            (r.shares           || [])[i] || 0,
      reels_avg_watch_time:    (r.avg_watch_time   || [])[i] || 0,
      reels_total_watch_time:  (r.total_watch_time || [])[i] || 0
    };
  });
}

function ig_fetchReelsRollup(p) {
  var j = analyticsGet(buildUrl_ig('reelsPerformance', p), p.access_token);
  var c = (j.reels_rollup || {}).current  || {};
  var v = (j.reels_rollup || {}).previous || {};
  return [{
    period:                           'Current',
    reels_rollup_curr_engagement:     Number(c.engagement       || 0),
    reels_rollup_curr_likes:          Number(c.likes            || 0),
    reels_rollup_curr_comments:       Number(c.comments         || 0),
    reels_rollup_curr_saves:          Number(c.saves            || 0),
    reels_rollup_curr_total_posts:    Number(c.total_posts      || 0),
    reels_rollup_curr_shares:         Number(c.shares           || 0),
    reels_rollup_curr_avg_watch_time: Number(c.avg_watch_time   || 0),
    reels_rollup_curr_total_watch_time:Number(c.total_watch_time|| 0),
    reels_rollup_prev_engagement:     Number(v.engagement       || 0),
    reels_rollup_prev_likes:          Number(v.likes            || 0),
    reels_rollup_prev_comments:       Number(v.comments         || 0),
    reels_rollup_prev_saves:          Number(v.saves            || 0),
    reels_rollup_prev_total_posts:    Number(v.total_posts      || 0),
    reels_rollup_prev_shares:         Number(v.shares           || 0),
    reels_rollup_prev_avg_watch_time: Number(v.avg_watch_time   || 0),
    reels_rollup_prev_total_watch_time:Number(v.total_watch_time|| 0)
  }];
}

function ig_fetchHashtags(p) {
  var j  = analyticsGet(buildUrl_ig('hashtags', p), p.access_token);
  var th = j.top_hashtags || {};
  return (th.name || []).map(function(name, i) {
    return {
      top_hashtags_name:       name,
      top_hashtags_posts:      (th.posts      || [])[i] || 0,
      top_hashtags_engagement: (th.engagement || [])[i] || 0,
      top_hashtags_likes:      (th.likes      || [])[i] || 0,
      top_hashtags_comments:   (th.comments   || [])[i] || 0,
      top_hashtags_saved:      (th.saved      || [])[i] || 0
    };
  });
}

function ig_fetchHashtagsRollup(p) {
  var j = analyticsGet(buildUrl_ig('hashtags', p), p.access_token);
  var c = (j.top_hashtags_rollup || {}).current  || {};
  var v = (j.top_hashtags_rollup || {}).previous || {};
  return [{
    period:                                       'Current',
    top_hashtags_rollup_curr_total_engagement:    Number(c.total_engagement      || 0),
    top_hashtags_rollup_curr_total_likes:         Number(c.total_likes           || 0),
    top_hashtags_rollup_curr_total_comments:      Number(c.total_comments        || 0),
    top_hashtags_rollup_curr_total_saves:         Number(c.total_saves           || 0),
    top_hashtags_rollup_curr_total_unique_hashtags:Number(c.total_unique_hashtags|| 0),
    top_hashtags_rollup_curr_total_hashtag_uses:  Number(c.total_hashtag_uses    || 0),
    top_hashtags_rollup_prev_total_engagement:    Number(v.total_engagement      || 0),
    top_hashtags_rollup_prev_total_likes:         Number(v.total_likes           || 0),
    top_hashtags_rollup_prev_total_comments:      Number(v.total_comments        || 0),
    top_hashtags_rollup_prev_total_saves:         Number(v.total_saves           || 0),
    top_hashtags_rollup_prev_total_unique_hashtags:Number(v.total_unique_hashtags|| 0),
    top_hashtags_rollup_prev_total_hashtag_uses:  Number(v.total_hashtag_uses    || 0)
  }];
}

function ig_fetchImpressionsTrend(p) {
  var j   = analyticsGet(buildUrl_ig('impressions', p), p.access_token);
  var imp = j.impressions || {};
  return (imp.buckets || []).map(function(date, i) {
    return {
      date:                    date.replace(/-/g, ''),
      impressions_impressions: (imp.impressions || [])[i] || 0
    };
  });
}

function ig_fetchImpressionsRollup(p) {
  var j = analyticsGet(buildUrl_ig('impressions', p), p.access_token);
  var c = (j.impressions_rollup || {}).current  || {};
  var v = (j.impressions_rollup || {}).previous || {};
  return [{
    period:                                    'Current',
    impressions_rollup_curr_total_impressions:  Number(c.total_impressions || 0),
    impressions_rollup_curr_avg_impressions:    Number(c.avg_impressions   || 0),
    impressions_rollup_prev_total_impressions:  Number(v.total_impressions || 0),
    impressions_rollup_prev_avg_impressions:    Number(v.avg_impressions   || 0)
  }];
}

function ig_fetchEngagementTrend(p) {
  var j   = analyticsGet(buildUrl_ig('engagement', p), p.access_token);
  var eng = j.engagements || {};
  return (eng.buckets || []).map(function(date, i) {
    return {
      date:                   date.replace(/-/g, ''),
      engagements_engagement: (eng.engagement || [])[i] || 0,
      engagements_comments:   (eng.comments   || [])[i] || 0,
      engagements_reactions:  (eng.reactions  || [])[i] || 0
    };
  });
}

function ig_fetchEngagementRollup(p) {
  var j = analyticsGet(buildUrl_ig('engagement', p), p.access_token);
  var c = (j.engagements_rollup || {}).current  || {};
  var v = (j.engagements_rollup || {}).previous || {};
  return [{
    period:                               'Current',
    engagements_rollup_curr_engagement:   Number(c.engagement     || 0),
    engagements_rollup_curr_avg_engagement:Number(c.avg_engagement|| 0),
    engagements_rollup_curr_comments:     Number(c.comments       || 0),
    engagements_rollup_curr_reactions:    Number(c.reactions      || 0),
    engagements_rollup_curr_saved:        Number(c.saved          || 0),
    engagements_rollup_curr_count:        Number(c.count          || 0),
    engagements_rollup_prev_engagement:   Number(v.engagement     || 0),
    engagements_rollup_prev_avg_engagement:Number(v.avg_engagement|| 0),
    engagements_rollup_prev_comments:     Number(v.comments       || 0),
    engagements_rollup_prev_reactions:    Number(v.reactions      || 0),
    engagements_rollup_prev_saved:        Number(v.saved          || 0),
    engagements_rollup_prev_count:        Number(v.count          || 0)
  }];
}

function ig_fetchActiveDays(p) {
  var j    = analyticsGet(buildUrl_ig('activeUsers', p), p.access_token);
  var days = j.active_users_days || {};
  return (days.buckets || []).map(function(day, i) {
    return {
      active_users_days_buckets:       day,
      active_users_days_values:        (days.values || [])[i]    || 0,
      active_users_days_highest_value: Number(days.highest_value || 0),
      active_users_days_highest_day:   days.highest_day          || ''
    };
  });
}

function ig_fetchActiveHours(p) {
  var j     = analyticsGet(buildUrl_ig('activeUsers', p), p.access_token);
  var hours = j.active_users_hours || {};
  return (hours.buckets || []).map(function(hour, i) {
    return {
      active_users_hours_buckets:       Number(hour                    || 0),
      active_users_hours_values:        (hours.values || [])[i]        || 0,
      active_users_hours_highest_value: Number(hours.highest_value     || 0),
      active_users_hours_highest_hour:  Number(hours.highest_hour      || 0)
    };
  });
}

// ============================================================
// UPDATED TOP POSTS WITH MORE COMPREHENSIVE DATA FETCHING
// ============================================================
// ============================================================
// UPDATED TOP POSTS - SINGLE CALL WITH LIMIT=100
// ============================================================
function ig_fetchTopPosts(p) {
  var data = analyticsGet(
    buildUrl_ig('getTopPosts', p, '&limit=100&order_by=views'),
    p.access_token
  );

  return (data.top_posts || []).map(function(post) {
    return {
      top_posts_post_created_at:        toDateStr(post.post_created_at    || ''),
      top_posts_stored_event_at:        toDateStr(post.stored_event_at    || ''),
      top_posts_instagram_id:           post.instagram_id                  || '',
      top_posts_media_id:               post.media_id                      || '',
      top_posts_caption:                post.caption                       || '',
      top_posts_media_type:             post.media_type                    || '',
      top_posts_entity_type:            post.entity_type                   || '',
      top_posts_media_url:              (post.media_url  || []).join(','),
      top_posts_video_url:              (post.video_url  || []).join(','),
      top_posts_permalink:              post.permalink                     || '',
      top_posts_hashtags:               (post.hashtags   || []).join(','),
      top_posts_day_of_week:            post.day_of_week                   || '',
      top_posts_hour_of_day:            Number(post.hour_of_day            || 0),
      top_posts_like_count:             Number(post.like_count             || 0),
      top_posts_comments_count:         Number(post.comments_count         || 0),
      top_posts_saved:                  Number(post.saved                  || 0),
      top_posts_engagement:             Number(post.engagement             || 0),
      top_posts_reach:                  Number(post.reach                  || 0),
      top_posts_impressions:            Number(post.impressions            || 0),
      top_posts_views:                  Number(post.views                  || 0),
      top_posts_shares:                 Number(post.shares                 || 0),
      top_posts_reels_avg_watch_time:   Number(post.reels_avg_watch_time   || 0),
      top_posts_reels_total_watch_time: Number(post.reels_total_watch_time || 0),
      top_posts_exits:                  Number(post.exits                  || 0),
      top_posts_replies:                Number(post.replies                || 0)
    };
  });
}

function ig_fetchDemographics(p, mode) {
  var j = analyticsGet(buildUrl_ig('demographicsAge', p), p.access_token);
  if (mode === 'age') {
    var ageMap = j.audience_age || {};
    return Object.keys(ageMap).map(function(bracket) {
      return { audience_age_bracket: bracket, audience_age_count: ageMap[bracket] || 0 };
    });
  }
  var genderMap = j.audience_gender || {};
  return Object.keys(genderMap).map(function(g) {
    return { audience_gender_gender: g, audience_gender_count: genderMap[g] || 0 };
  });
}

function ig_fetchAudienceLocation(p, mode) {
  var j = analyticsGet(buildUrl_ig('countryCity', p), p.access_token);
  if (mode === 'city') {
    var cityMap = j.audience_city || {};
    return Object.keys(cityMap).map(function(c) {
      return { audience_city_city: c, audience_city_count: cityMap[c] || 0 };
    });
  }
  var countryMap = j.audience_country || {};
  return Object.keys(countryMap).map(function(c) {
    return { audience_country_country: c, audience_country_count: countryMap[c] || 0 };
  });
}