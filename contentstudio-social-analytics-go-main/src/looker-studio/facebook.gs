// ============================================================
// facebook.gs ? Facebook Analytics
// ============================================================

function getFields_facebook(fields, types) {
  // Shared dimensions
  fields.newDimension().setId('date').setName('Date').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('period').setName('Period').setType(types.TEXT);

  // Summary ? current period (root: overview)
  fields.newMetric().setId('sum_curr_fan_count').setName('Sum Curr Fan Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_doc_count').setName('Sum Curr Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_total_engagement').setName('Sum Curr Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_reactions').setName('Sum Curr Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_comments').setName('Sum Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_posts_clicks').setName('Sum Curr Posts Clicks').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_impressions').setName('Sum Curr Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_reach').setName('Sum Curr Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_repost').setName('Sum Curr Repost').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_positive_sentiment').setName('Sum Curr Positive Sentiment').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_negative_sentiment').setName('Sum Curr Negative Sentiment').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_impressions').setName('Sum Curr Page Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_impressions_paid').setName('Sum Curr Page Impressions Paid').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_impressions_organic').setName('Sum Curr Page Impressions Organic').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_engagements').setName('Sum Curr Page Engagements').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_positive_feedback').setName('Sum Curr Page Positive Feedback').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_negative_feedback').setName('Sum Curr Page Negative Feedback').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_talking_about_count').setName('Sum Curr Talking About Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_curr_page_follows').setName('Sum Curr Page Follows').setType(types.NUMBER);
  // Summary ? previous period
  fields.newMetric().setId('sum_prev_fan_count').setName('Sum Prev Fan Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_doc_count').setName('Sum Prev Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_total_engagement').setName('Sum Prev Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_reactions').setName('Sum Prev Reactions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_comments').setName('Sum Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_posts_clicks').setName('Sum Prev Posts Clicks').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_impressions').setName('Sum Prev Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_reach').setName('Sum Prev Reach').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_repost').setName('Sum Prev Repost').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_positive_sentiment').setName('Sum Prev Positive Sentiment').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_negative_sentiment').setName('Sum Prev Negative Sentiment').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_impressions').setName('Sum Prev Page Impressions').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_impressions_paid').setName('Sum Prev Page Impressions Paid').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_impressions_organic').setName('Sum Prev Page Impressions Organic').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_engagements').setName('Sum Prev Page Engagements').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_positive_feedback').setName('Sum Prev Page Positive Feedback').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_negative_feedback').setName('Sum Prev Page Negative Feedback').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_talking_about_count').setName('Sum Prev Talking About Count').setType(types.NUMBER);
  fields.newMetric().setId('sum_prev_page_follows').setName('Sum Prev Page Follows').setType(types.NUMBER);

  // Audience growth ? daily (root: audience_growth)
  fields.newMetric().setId('audience_growth_fan_count').setName('Audience Growth Fan Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_page_fans_daily').setName('Audience Growth Page Fans Daily').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_page_fans_by_like').setName('Audience Growth Page Fans By Like').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_page_fans_by_unlike').setName('Audience Growth Page Fans By Unlike').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_page_impressions').setName('Audience Growth Page Impressions').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_page_engagements').setName('Audience Growth Page Engagements').setType(types.NUMBER);
  // Audience growth rollup ? current (root: audience_growth_rollup)
  fields.newMetric().setId('audience_growth_rollup_curr_fan_count').setName('Audience Growth Rollup Curr Fan Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_avg_page_fans_by_like').setName('Audience Growth Rollup Curr Avg Page Fans By Like').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_avg_page_fans_by_unlike').setName('Audience Growth Rollup Curr Avg Page Fans By Unlike').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_talking_about_count').setName('Audience Growth Rollup Curr Talking About Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_curr_doc_count').setName('Audience Growth Rollup Curr Doc Count').setType(types.NUMBER);
  // Audience growth rollup ? previous
  fields.newMetric().setId('audience_growth_rollup_prev_fan_count').setName('Audience Growth Rollup Prev Fan Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_avg_page_fans_by_like').setName('Audience Growth Rollup Prev Avg Page Fans By Like').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_avg_page_fans_by_unlike').setName('Audience Growth Rollup Prev Avg Page Fans By Unlike').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_talking_about_count').setName('Audience Growth Rollup Prev Talking About Count').setType(types.NUMBER);
  fields.newMetric().setId('audience_growth_rollup_prev_doc_count').setName('Audience Growth Rollup Prev Doc Count').setType(types.NUMBER);

  // Publishing behaviour ? daily (root: publishing_behaviour)
  fields.newMetric().setId('publishing_behaviour_post_count').setName('Publishing Behaviour Post Count').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_reactions_engagement').setName('Publishing Behaviour Reactions Engagement').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_comments_engagement').setName('Publishing Behaviour Comments Engagement').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_shares_engagement').setName('Publishing Behaviour Shares Engagement').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_organic_reach').setName('Publishing Behaviour Organic Reach').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_paid_reach').setName('Publishing Behaviour Paid Reach').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_viral_reach').setName('Publishing Behaviour Viral Reach').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_organic_impressions').setName('Publishing Behaviour Organic Impressions').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_paid_impressions').setName('Publishing Behaviour Paid Impressions').setType(types.NUMBER);
  fields.newMetric().setId('publishing_behaviour_viral_impressions').setName('Publishing Behaviour Viral Impressions').setType(types.NUMBER);
  // Publishing behaviour rollup ? current (root: publishing_behaviour_rollup)
  fields.newMetric().setId('pub_beh_rlu_curr_doc_count').setName('Pub Beh Rlu Curr Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_total_engagement').setName('Pub Beh Rlu Curr Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_reactions').setName('Pub Beh Rlu Curr Reactions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_comments').setName('Pub Beh Rlu Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_post_clicks').setName('Pub Beh Rlu Curr Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_impressions').setName('Pub Beh Rlu Curr Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_curr_shares').setName('Pub Beh Rlu Curr Shares').setType(types.NUMBER);
  // Publishing behaviour rollup ? previous
  fields.newMetric().setId('pub_beh_rlu_prev_doc_count').setName('Pub Beh Rlu Prev Doc Count').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_total_engagement').setName('Pub Beh Rlu Prev Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_reactions').setName('Pub Beh Rlu Prev Reactions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_comments').setName('Pub Beh Rlu Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_post_clicks').setName('Pub Beh Rlu Prev Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_impressions').setName('Pub Beh Rlu Prev Impressions').setType(types.NUMBER);
  fields.newMetric().setId('pub_beh_rlu_prev_shares').setName('Pub Beh Rlu Prev Shares').setType(types.NUMBER);

  // Impressions ? daily (root: impressions)
  fields.newMetric().setId('impressions_page_impressions').setName('Impressions Page Impressions').setType(types.NUMBER);
  // Impressions rollup ? current (root: impressions_rollup)
  fields.newMetric().setId('impressions_rollup_curr_total_impressions').setName('Impressions Rollup Curr Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_curr_avg_impressions_per_day').setName('Impressions Rollup Curr Avg Impressions Per Day').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_curr_avg_impressions_per_week').setName('Impressions Rollup Curr Avg Impressions Per Week').setType(types.NUMBER);
  // Impressions rollup ? previous
  fields.newMetric().setId('impressions_rollup_prev_total_impressions').setName('Impressions Rollup Prev Total Impressions').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_prev_avg_impressions_per_day').setName('Impressions Rollup Prev Avg Impressions Per Day').setType(types.NUMBER);
  fields.newMetric().setId('impressions_rollup_prev_avg_impressions_per_week').setName('Impressions Rollup Prev Avg Impressions Per Week').setType(types.NUMBER);

  // Engagement ? daily (root: engagement)
  fields.newMetric().setId('engagement_page_engagements').setName('Engagement Page Engagements').setType(types.NUMBER);
  // Engagement rollup ? current (root: engagement_rollup)
  fields.newMetric().setId('engagement_rollup_curr_page_engagements').setName('Engagement Rollup Curr Page Engagements').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rollup_curr_avg_engagements_per_day').setName('Engagement Rollup Curr Avg Engagements Per Day').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rollup_curr_avg_engagements_per_week').setName('Engagement Rollup Curr Avg Engagements Per Week').setType(types.NUMBER);
  // Engagement rollup ? previous
  fields.newMetric().setId('engagement_rollup_prev_page_engagements').setName('Engagement Rollup Prev Page Engagements').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rollup_prev_avg_engagements_per_day').setName('Engagement Rollup Prev Avg Engagements Per Day').setType(types.NUMBER);
  fields.newMetric().setId('engagement_rollup_prev_avg_engagements_per_week').setName('Engagement Rollup Prev Avg Engagements Per Week').setType(types.NUMBER);

  // Video insights ? daily (root: video_insights)
  fields.newMetric().setId('video_insights_total_views').setName('Video Insights Total Views').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_organic_views').setName('Video Insights Organic Views').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_paid_views').setName('Video Insights Paid Views').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_total_view_time').setName('Video Insights Total View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_organic_view_time').setName('Video Insights Organic View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_paid_view_time').setName('Video Insights Paid View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_comments').setName('Video Insights Comments').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_reactions').setName('Video Insights Reactions').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_shares').setName('Video Insights Shares').setType(types.NUMBER);
  fields.newMetric().setId('video_insights_total_posts').setName('Video Insights Total Posts').setType(types.NUMBER);
  // Video rollup ? current (root: video_rollup)
  fields.newMetric().setId('video_rollup_curr_total_views').setName('Video Rollup Curr Total Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_organic_views').setName('Video Rollup Curr Organic Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_paid_views').setName('Video Rollup Curr Paid Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_total_view_time').setName('Video Rollup Curr Total View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_organic_view_time').setName('Video Rollup Curr Organic View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_paid_view_time').setName('Video Rollup Curr Paid View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_comments').setName('Video Rollup Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_reactions').setName('Video Rollup Curr Reactions').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_shares').setName('Video Rollup Curr Shares').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_curr_total_posts').setName('Video Rollup Curr Total Posts').setType(types.NUMBER);
  // Video rollup ? previous
  fields.newMetric().setId('video_rollup_prev_total_views').setName('Video Rollup Prev Total Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_organic_views').setName('Video Rollup Prev Organic Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_paid_views').setName('Video Rollup Prev Paid Views').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_total_view_time').setName('Video Rollup Prev Total View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_organic_view_time').setName('Video Rollup Prev Organic View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_paid_view_time').setName('Video Rollup Prev Paid View Time').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_comments').setName('Video Rollup Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_reactions').setName('Video Rollup Prev Reactions').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_shares').setName('Video Rollup Prev Shares').setType(types.NUMBER);
  fields.newMetric().setId('video_rollup_prev_total_posts').setName('Video Rollup Prev Total Posts').setType(types.NUMBER);

  // Reels ? daily (root: reels)
  fields.newMetric().setId('reels_initial_plays').setName('Reels Initial Plays').setType(types.NUMBER);
  fields.newMetric().setId('reels_total_reels').setName('Reels Total Reels').setType(types.NUMBER);
  fields.newMetric().setId('reels_total_seconds_watched').setName('Reels Total Seconds Watched').setType(types.NUMBER);
  fields.newMetric().setId('reels_engagement').setName('Reels Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_reactions').setName('Reels Reactions').setType(types.NUMBER);
  fields.newMetric().setId('reels_comments').setName('Reels Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_shares').setName('Reels Shares').setType(types.NUMBER);
  // Reels rollup ? current (root: reels_rollup)
  fields.newMetric().setId('reels_rollup_curr_total_reels').setName('Reels Rollup Curr Total Reels').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_initial_plays').setName('Reels Rollup Curr Initial Plays').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_total_seconds_watched').setName('Reels Rollup Curr Total Seconds Watched').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_average_seconds_watched').setName('Reels Rollup Curr Average Seconds Watched').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_reach').setName('Reels Rollup Curr Reach').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_engagement').setName('Reels Rollup Curr Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_reactions').setName('Reels Rollup Curr Reactions').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_comments').setName('Reels Rollup Curr Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_curr_shares').setName('Reels Rollup Curr Shares').setType(types.NUMBER);
  // Reels rollup ? previous
  fields.newMetric().setId('reels_rollup_prev_total_reels').setName('Reels Rollup Prev Total Reels').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_initial_plays').setName('Reels Rollup Prev Initial Plays').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_total_seconds_watched').setName('Reels Rollup Prev Total Seconds Watched').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_average_seconds_watched').setName('Reels Rollup Prev Average Seconds Watched').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_reach').setName('Reels Rollup Prev Reach').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_engagement').setName('Reels Rollup Prev Engagement').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_reactions').setName('Reels Rollup Prev Reactions').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_comments').setName('Reels Rollup Prev Comments').setType(types.NUMBER);
  fields.newMetric().setId('reels_rollup_prev_shares').setName('Reels Rollup Prev Shares').setType(types.NUMBER);

  // Top posts ? dimensions (root: top_posts)
  fields.newDimension().setId('top_posts_created_time').setName('Top Posts Created Time').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_updated_time').setName('Top Posts Updated Time').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_saving_time').setName('Top Posts Saving Time').setType(types.YEAR_MONTH_DAY);
  fields.newDimension().setId('top_posts_page_name').setName('Top Posts Page Name').setType(types.TEXT);
  fields.newDimension().setId('top_posts_page_id').setName('Top Posts Page ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_post_id').setName('Top Posts Post ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_permalink').setName('Top Posts Permalink').setType(types.URL);
  fields.newDimension().setId('top_posts_status_type').setName('Top Posts Status Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_media_type').setName('Top Posts Media Type').setType(types.TEXT);
  fields.newDimension().setId('top_posts_video_id').setName('Top Posts Video ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_category').setName('Top Posts Category').setType(types.TEXT);
  fields.newDimension().setId('top_posts_published_by').setName('Top Posts Published By').setType(types.TEXT);
  fields.newDimension().setId('top_posts_published_by_url').setName('Top Posts Published By URL').setType(types.URL);
  fields.newDimension().setId('top_posts_shared_from_name').setName('Top Posts Shared From Name').setType(types.TEXT);
  fields.newDimension().setId('top_posts_shared_from_id').setName('Top Posts Shared From ID').setType(types.TEXT);
  fields.newDimension().setId('top_posts_shared_from_link').setName('Top Posts Shared From Link').setType(types.URL);
  fields.newDimension().setId('top_posts_caption').setName('Top Posts Caption').setType(types.TEXT);
  fields.newDimension().setId('top_posts_description').setName('Top Posts Description').setType(types.TEXT);
  fields.newDimension().setId('top_posts_full_picture').setName('Top Posts Full Picture').setType(types.URL);
  fields.newDimension().setId('top_posts_link').setName('Top Posts Link').setType(types.URL);
  fields.newDimension().setId('top_posts_day_of_week').setName('Top Posts Day Of Week').setType(types.TEXT);
  fields.newDimension().setId('top_posts_message_tags').setName('Top Posts Message Tags').setType(types.TEXT);
  fields.newDimension().setId('top_posts_post_metadata').setName('Top Posts Post Metadata').setType(types.TEXT);
  fields.newDimension().setId('tp_medass_media_id').setName('Media Asset ID').setType(types.TEXT);
  fields.newDimension().setId('tp_medass_caption').setName('Media Asset Caption').setType(types.TEXT);
  fields.newDimension().setId('tp_medass_link').setName('Media Asset Link').setType(types.URL);
  fields.newDimension().setId('tp_medass_asset_type').setName('Media Asset Type').setType(types.TEXT);
  fields.newDimension().setId('tp_medass_call_to_action').setName('Media Asset Call To Action').setType(types.URL);
  fields.newDimension().setId('tp_medass_created_at').setName('Media Asset Created At').setType(types.TEXT);
  // Top posts ? metrics
  fields.newMetric().setId('top_posts_hour_of_day').setName('Top Posts Hour Of Day').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_like').setName('Top Posts Like').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_love').setName('Top Posts Love').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_haha').setName('Top Posts Haha').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_wow').setName('Top Posts Wow').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_sad').setName('Top Posts Sad').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_angry').setName('Top Posts Angry').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_total').setName('Top Posts Total').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_shares').setName('Top Posts Shares').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_comments').setName('Top Posts Comments').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_clicks').setName('Top Posts Post Clicks').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_total_engagement').setName('Top Posts Total Engagement').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_engaged_users').setName('Top Posts Post Engaged Users').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions').setName('Top Posts Post Impressions').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_unique').setName('Top Posts Post Impressions Unique').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_paid').setName('Top Posts Post Impressions Paid').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_paid_unique').setName('Top Posts Post Impressions Paid Unique').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_organic').setName('Top Posts Post Impressions Organic').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_organic_unique').setName('Top Posts Post Impressions Organic Unique').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_viral').setName('Top Posts Post Impressions Viral').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_impressions_viral_unique').setName('Top Posts Post Impressions Viral Unique').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_post_video_views').setName('Top Posts Post Video Views').setType(types.NUMBER);
  fields.newMetric().setId('top_posts_total_impressions').setName('Top Posts Total Impressions').setType(types.NUMBER);

  // Active users ? days (root: active_users ? active_users_days)
  fields.newDimension().setId('active_users_days_buckets').setName('Active Users Days Buckets').setType(types.TEXT);
  fields.newMetric().setId('active_users_days_values').setName('Active Users Days Values').setType(types.NUMBER);
  fields.newMetric().setId('active_users_days_highest_value').setName('Active Users Days Highest Value').setType(types.NUMBER);
  fields.newDimension().setId('active_users_days_highest_day').setName('Active Users Days Highest Day').setType(types.TEXT);

  // Demographics ? age (root: audience_age ? fans_age)
  fields.newDimension().setId('audience_age_fans_age_bracket').setName('Audience Age Fans Age Bracket').setType(types.TEXT);
  fields.newMetric().setId('audience_age_fans_age_count').setName('Audience Age Fans Age Count').setType(types.NUMBER);

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

function buildUrl_fb(endpoint, p, extra) {
  return p.analytics_go + 'facebook/' + endpoint
    + buildBaseParams(p,
        '&facebook_id=' + encodeURIComponent(p.account_id)
        + '&type=' + endpoint
        + (extra || ''));
}

function getData_facebook(p) {
  var reqIds = p._reqIds || [];
  var best = bestMatch(reqIds, {
    fbLocCity:            ['audience_city_city','audience_city_count'],
    fbLocCountry:         ['audience_country_country','audience_country_count'],
    ageDemo:              ['audience_age_fans_age_bracket','audience_age_fans_age_count'],
    genderDemo:           ['audience_gender_gender','audience_gender_count'],
    active:               ['active_users_days_buckets','active_users_days_values','active_users_days_highest_value','active_users_days_highest_day'],
    engagementTrend:      ['date','engagement_page_engagements'],
    topPosts:             ['top_posts_created_time','top_posts_updated_time','top_posts_saving_time',
                           'top_posts_page_name','top_posts_page_id','top_posts_post_id','top_posts_permalink',
                           'top_posts_status_type','top_posts_media_type','top_posts_video_id','top_posts_category',
                           'top_posts_published_by','top_posts_published_by_url','top_posts_shared_from_name','top_posts_shared_from_id','top_posts_shared_from_link',
                           'top_posts_caption','top_posts_description','top_posts_full_picture','top_posts_link',
                           'top_posts_day_of_week','top_posts_hour_of_day','top_posts_message_tags','top_posts_post_metadata',
                           'tp_medass_media_id','tp_medass_caption','tp_medass_link','tp_medass_asset_type','tp_medass_call_to_action','tp_medass_created_at',
                           'top_posts_like','top_posts_love','top_posts_haha','top_posts_wow','top_posts_sad','top_posts_angry','top_posts_total',
                           'top_posts_shares','top_posts_comments','top_posts_post_clicks','top_posts_total_engagement','top_posts_post_engaged_users',
                           'top_posts_post_impressions','top_posts_post_impressions_unique',
                           'top_posts_post_impressions_paid','top_posts_post_impressions_paid_unique',
                           'top_posts_post_impressions_organic','top_posts_post_impressions_organic_unique',
                           'top_posts_post_impressions_viral','top_posts_post_impressions_viral_unique',
                           'top_posts_post_video_views','top_posts_total_impressions'],
    reelsRollup:          ['reels_rollup_curr_total_reels','reels_rollup_curr_initial_plays','reels_rollup_curr_total_seconds_watched','reels_rollup_curr_average_seconds_watched','reels_rollup_curr_reach','reels_rollup_curr_engagement','reels_rollup_curr_reactions','reels_rollup_curr_comments','reels_rollup_curr_shares',
                           'reels_rollup_prev_total_reels','reels_rollup_prev_initial_plays','reels_rollup_prev_total_seconds_watched','reels_rollup_prev_average_seconds_watched','reels_rollup_prev_reach','reels_rollup_prev_engagement','reels_rollup_prev_reactions','reels_rollup_prev_comments','reels_rollup_prev_shares'],
    videoRollup:          ['video_rollup_curr_total_views','video_rollup_curr_organic_views','video_rollup_curr_paid_views','video_rollup_curr_total_view_time','video_rollup_curr_organic_view_time','video_rollup_curr_paid_view_time','video_rollup_curr_comments','video_rollup_curr_reactions','video_rollup_curr_shares','video_rollup_curr_total_posts',
                           'video_rollup_prev_total_views','video_rollup_prev_organic_views','video_rollup_prev_paid_views','video_rollup_prev_total_view_time','video_rollup_prev_organic_view_time','video_rollup_prev_paid_view_time','video_rollup_prev_comments','video_rollup_prev_reactions','video_rollup_prev_shares','video_rollup_prev_total_posts'],
    pubRollup:            ['pub_beh_rlu_curr_doc_count','pub_beh_rlu_curr_total_engagement','pub_beh_rlu_curr_reactions','pub_beh_rlu_curr_comments','pub_beh_rlu_curr_post_clicks','pub_beh_rlu_curr_impressions','pub_beh_rlu_curr_shares',
                           'pub_beh_rlu_prev_doc_count','pub_beh_rlu_prev_total_engagement','pub_beh_rlu_prev_reactions','pub_beh_rlu_prev_comments','pub_beh_rlu_prev_post_clicks','pub_beh_rlu_prev_impressions','pub_beh_rlu_prev_shares'],
    impRollup:            ['impressions_rollup_curr_total_impressions','impressions_rollup_curr_avg_impressions_per_day','impressions_rollup_curr_avg_impressions_per_week',
                           'impressions_rollup_prev_total_impressions','impressions_rollup_prev_avg_impressions_per_day','impressions_rollup_prev_avg_impressions_per_week'],
    engRollup:            ['engagement_rollup_curr_page_engagements','engagement_rollup_curr_avg_engagements_per_day','engagement_rollup_curr_avg_engagements_per_week',
                           'engagement_rollup_prev_page_engagements','engagement_rollup_prev_avg_engagements_per_day','engagement_rollup_prev_avg_engagements_per_week'],
    audienceGrowthRollup: ['audience_growth_rollup_curr_fan_count','audience_growth_rollup_curr_avg_page_fans_by_like','audience_growth_rollup_curr_avg_page_fans_by_unlike','audience_growth_rollup_curr_talking_about_count','audience_growth_rollup_curr_doc_count',
                           'audience_growth_rollup_prev_fan_count','audience_growth_rollup_prev_avg_page_fans_by_like','audience_growth_rollup_prev_avg_page_fans_by_unlike','audience_growth_rollup_prev_talking_about_count','audience_growth_rollup_prev_doc_count'],
    reels:                ['date','reels_initial_plays','reels_total_reels','reels_total_seconds_watched','reels_engagement','reels_reactions','reels_comments','reels_shares'],
    video:                ['date','video_insights_total_views','video_insights_organic_views','video_insights_paid_views','video_insights_total_view_time','video_insights_organic_view_time','video_insights_paid_view_time','video_insights_comments','video_insights_reactions','video_insights_shares','video_insights_total_posts'],
    summary:              ['period',
                           'sum_curr_fan_count','sum_curr_doc_count','sum_curr_total_engagement','sum_curr_reactions','sum_curr_comments','sum_curr_posts_clicks','sum_curr_impressions','sum_curr_reach','sum_curr_repost','sum_curr_positive_sentiment','sum_curr_negative_sentiment',
                           'sum_curr_page_impressions','sum_curr_page_impressions_paid','sum_curr_page_impressions_organic','sum_curr_page_engagements','sum_curr_page_positive_feedback','sum_curr_page_negative_feedback','sum_curr_talking_about_count','sum_curr_page_follows',
                           'sum_prev_fan_count','sum_prev_doc_count','sum_prev_total_engagement','sum_prev_reactions','sum_prev_comments','sum_prev_posts_clicks','sum_prev_impressions','sum_prev_reach','sum_prev_repost','sum_prev_positive_sentiment','sum_prev_negative_sentiment',
                           'sum_prev_page_impressions','sum_prev_page_impressions_paid','sum_prev_page_impressions_organic','sum_prev_page_engagements','sum_prev_page_positive_feedback','sum_prev_page_negative_feedback','sum_prev_talking_about_count','sum_prev_page_follows'],
    audience:             ['date','audience_growth_fan_count','audience_growth_page_fans_daily','audience_growth_page_fans_by_like','audience_growth_page_fans_by_unlike','audience_growth_page_impressions','audience_growth_page_engagements'],
    impressionsTrend:     ['date','impressions_page_impressions'],
    pub:                  ['date','publishing_behaviour_post_count','publishing_behaviour_reactions_engagement','publishing_behaviour_comments_engagement','publishing_behaviour_shares_engagement',
                           'publishing_behaviour_organic_reach','publishing_behaviour_paid_reach','publishing_behaviour_viral_reach',
                           'publishing_behaviour_organic_impressions','publishing_behaviour_paid_impressions','publishing_behaviour_viral_impressions']
  });

  switch(best) {
    case 'fbLocCity':            return fb_fetchAudienceLocation(p, 'city');
    case 'fbLocCountry':         return fb_fetchAudienceLocation(p, 'country');
    case 'ageDemo':              return fb_fetchDemographics(p, 'age');
    case 'genderDemo':           return fb_fetchDemographics(p, 'gender');
    case 'active':               return fb_fetchActiveUsers(p);
    case 'engagementTrend':      return fb_fetchEngagementTrend(p);
    case 'impressionsTrend':     return fb_fetchImpressionsTrend(p);
    case 'topPosts':             return fb_fetchTopPosts(p);
    case 'reelsRollup':          return fb_fetchReelsRollup(p);
    case 'videoRollup':          return fb_fetchVideoRollup(p);
    case 'pubRollup':            return fb_fetchPubRollup(p);
    case 'impRollup':            return fb_fetchImpressionsRollup(p);
    case 'engRollup':            return fb_fetchEngagementRollup(p);
    case 'audienceGrowthRollup': return fb_fetchAudienceGrowthRollup(p);
    case 'reels':                return fb_fetchReels(p);
    case 'video':                return fb_fetchVideoInsights(p);
    case 'audience':             return fb_fetchAudienceGrowth(p);
    case 'pub':                  return fb_fetchPublishingBehaviour(p);
    default:                     return fb_fetchSummary(p);
  }
}

function fb_fetchSummary(p) {
  var j = analyticsGet(buildUrl_fb('summary', p), p.access_token);
  var c = (j.overview || {}).current  || {};
  var v = (j.overview || {}).previous || {};
  return [{
    period:                              'Current',
    sum_curr_fan_count:                  c.fan_count                  || 0,
    sum_curr_doc_count:                  c.doc_count                  || 0,
    sum_curr_total_engagement:           c.total_engagement           || 0,
    sum_curr_reactions:                  c.reactions                  || 0,
    sum_curr_comments:                   c.comments                   || 0,
    sum_curr_posts_clicks:               c.posts_clicks               || 0,
    sum_curr_impressions:                c.impressions                || 0,
    sum_curr_reach:                      c.reach                      || 0,
    sum_curr_repost:                     c.repost                     || 0,
    sum_curr_positive_sentiment:         c.positive_sentiment         || 0,
    sum_curr_negative_sentiment:         c.negative_sentiment         || 0,
    sum_curr_page_impressions:           c.page_impressions           || 0,
    sum_curr_page_impressions_paid:      c.page_impressions_paid      || 0,
    sum_curr_page_impressions_organic:   c.page_impressions_organic   || 0,
    sum_curr_page_engagements:           c.page_engagements           || 0,
    sum_curr_page_positive_feedback:     c.page_positive_feedback     || 0,
    sum_curr_page_negative_feedback:     c.page_negative_feedback     || 0,
    sum_curr_talking_about_count:        c.talking_about_count        || 0,
    sum_curr_page_follows:               c.page_follows               || 0,
    sum_prev_fan_count:                  v.fan_count                  || 0,
    sum_prev_doc_count:                  v.doc_count                  || 0,
    sum_prev_total_engagement:           v.total_engagement           || 0,
    sum_prev_reactions:                  v.reactions                  || 0,
    sum_prev_comments:                   v.comments                   || 0,
    sum_prev_posts_clicks:               v.posts_clicks               || 0,
    sum_prev_impressions:                v.impressions                || 0,
    sum_prev_reach:                      v.reach                      || 0,
    sum_prev_repost:                     v.repost                     || 0,
    sum_prev_positive_sentiment:         v.positive_sentiment         || 0,
    sum_prev_negative_sentiment:         v.negative_sentiment         || 0,
    sum_prev_page_impressions:           v.page_impressions           || 0,
    sum_prev_page_impressions_paid:      v.page_impressions_paid      || 0,
    sum_prev_page_impressions_organic:   v.page_impressions_organic   || 0,
    sum_prev_page_engagements:           v.page_engagements           || 0,
    sum_prev_page_positive_feedback:     v.page_positive_feedback     || 0,
    sum_prev_page_negative_feedback:     v.page_negative_feedback     || 0,
    sum_prev_talking_about_count:        v.talking_about_count        || 0,
    sum_prev_page_follows:               v.page_follows               || 0
  }];
}

function fb_fetchAudienceGrowth(p) {
  var j  = analyticsGet(buildUrl_fb('overviewAudienceGrowth', p), p.access_token);
  var ag = j.audience_growth || {};
  return (ag.buckets || []).reduce(function(acc, date, i) {
    var f = (ag.fan_count || [])[i] || 0;
    if (f > 0) acc.push({
      date:                                  date.replace(/-/g, ''),
      audience_growth_fan_count:             f,
      audience_growth_page_fans_daily:       (ag.page_fans_daily    || [])[i] || 0,
      audience_growth_page_fans_by_like:     (ag.page_fans_by_like  || [])[i] || 0,
      audience_growth_page_fans_by_unlike:   (ag.page_fans_by_unlike|| [])[i] || 0,
      audience_growth_page_impressions:      (ag.page_impressions   || [])[i] || 0,
      audience_growth_page_engagements:      (ag.page_engagements   || [])[i] || 0
    });
    return acc;
  }, []);
}

function fb_fetchAudienceGrowthRollup(p) {
  var j = analyticsGet(buildUrl_fb('overviewAudienceGrowth', p), p.access_token);
  var c = (j.audience_growth_rollup || {}).current  || {};
  var v = (j.audience_growth_rollup || {}).previous || {};
  return [{
    period:                                          'Current',
    audience_growth_rollup_curr_fan_count:           Number(c.fan_count                || 0),
    audience_growth_rollup_curr_avg_page_fans_by_like:  Number(c.avg_page_fans_by_like  || 0),
    audience_growth_rollup_curr_avg_page_fans_by_unlike:Number(c.avg_page_fans_by_unlike|| 0),
    audience_growth_rollup_curr_talking_about_count: Number(c.talking_about_count      || 0),
    audience_growth_rollup_curr_doc_count:           Number(c.doc_count                || 0),
    audience_growth_rollup_prev_fan_count:           Number(v.fan_count                || 0),
    audience_growth_rollup_prev_avg_page_fans_by_like:  Number(v.avg_page_fans_by_like  || 0),
    audience_growth_rollup_prev_avg_page_fans_by_unlike:Number(v.avg_page_fans_by_unlike|| 0),
    audience_growth_rollup_prev_talking_about_count: Number(v.talking_about_count      || 0),
    audience_growth_rollup_prev_doc_count:           Number(v.doc_count                || 0)
  }];
}

function fb_fetchPublishingBehaviour(p) {
  var j  = analyticsGet(buildUrl_fb('overviewPublishingBehaviour', p), p.access_token);
  var pb = j.publishing_behaviour || {};
  return (pb.buckets || []).map(function(date, i) {
    return {
      date:                                        date.replace(/-/g, ''),
      publishing_behaviour_post_count:             (pb.post_count            || [])[i] || 0,
      publishing_behaviour_reactions_engagement:   (pb.reactions_engagement  || [])[i] || 0,
      publishing_behaviour_comments_engagement:    (pb.comments_engagement   || [])[i] || 0,
      publishing_behaviour_shares_engagement:      (pb.shares_engagement     || [])[i] || 0,
      publishing_behaviour_organic_reach:          (pb.organic_reach         || [])[i] || 0,
      publishing_behaviour_paid_reach:             (pb.paid_reach            || [])[i] || 0,
      publishing_behaviour_viral_reach:            (pb.viral_reach           || [])[i] || 0,
      publishing_behaviour_organic_impressions:    (pb.organic_impressions   || [])[i] || 0,
      publishing_behaviour_paid_impressions:       (pb.paid_impressions      || [])[i] || 0,
      publishing_behaviour_viral_impressions:      (pb.viral_impressions     || [])[i] || 0
    };
  });
}

function fb_fetchPubRollup(p) {
  var j = analyticsGet(buildUrl_fb('overviewPublishingBehaviour', p), p.access_token);
  var c = (j.publishing_behaviour_rollup || {}).current  || {};
  var v = (j.publishing_behaviour_rollup || {}).previous || {};
  return [{
    period:                            'Current',
    pub_beh_rlu_curr_doc_count:        Number(c.doc_count        || 0),
    pub_beh_rlu_curr_total_engagement: Number(c.total_engagement || 0),
    pub_beh_rlu_curr_reactions:        Number(c.reactions        || 0),
    pub_beh_rlu_curr_comments:         Number(c.comments         || 0),
    pub_beh_rlu_curr_post_clicks:      Number(c.post_clicks      || 0),
    pub_beh_rlu_curr_impressions:      Number(c.impressions      || 0),
    pub_beh_rlu_curr_shares:           Number(c.shares           || 0),
    pub_beh_rlu_prev_doc_count:        Number(v.doc_count        || 0),
    pub_beh_rlu_prev_total_engagement: Number(v.total_engagement || 0),
    pub_beh_rlu_prev_reactions:        Number(v.reactions        || 0),
    pub_beh_rlu_prev_comments:         Number(v.comments         || 0),
    pub_beh_rlu_prev_post_clicks:      Number(v.post_clicks      || 0),
    pub_beh_rlu_prev_impressions:      Number(v.impressions      || 0),
    pub_beh_rlu_prev_shares:           Number(v.shares           || 0)
  }];
}

function fb_fetchImpressionsTrend(p) {
  var j   = analyticsGet(buildUrl_fb('overviewImpressions', p), p.access_token);
  var imp = j.impressions || {};
  return (imp.buckets || []).map(function(date, i) {
    return {
      date:                        date.replace(/-/g, ''),
      impressions_page_impressions:(imp.page_impressions || [])[i] || 0
    };
  });
}

function fb_fetchImpressionsRollup(p) {
  var j = analyticsGet(buildUrl_fb('overviewImpressions', p), p.access_token);
  var c = (j.impressions_rollup || {}).current  || {};
  var v = (j.impressions_rollup || {}).previous || {};
  return [{
    period:                                          'Current',
    impressions_rollup_curr_total_impressions:        Number(c.total_impressions        || 0),
    impressions_rollup_curr_avg_impressions_per_day:  Number(c.avg_impressions_per_day  || 0),
    impressions_rollup_curr_avg_impressions_per_week: Number(c.avg_impressions_per_week || 0),
    impressions_rollup_prev_total_impressions:        Number(v.total_impressions        || 0),
    impressions_rollup_prev_avg_impressions_per_day:  Number(v.avg_impressions_per_day  || 0),
    impressions_rollup_prev_avg_impressions_per_week: Number(v.avg_impressions_per_week || 0)
  }];
}

function fb_fetchEngagementTrend(p) {
  var j   = analyticsGet(buildUrl_fb('overviewEngagement', p), p.access_token);
  var eng = ((j.engagement || {}).engagement) || {};
  return (eng.buckets || []).map(function(date, i) {
    return {
      date:                        date.replace(/-/g, ''),
      engagement_page_engagements: (eng.page_engagements || [])[i] || 0
    };
  });
}

function fb_fetchEngagementRollup(p) {
  var j   = analyticsGet(buildUrl_fb('overviewEngagement', p), p.access_token);
  var eng = (j.engagement || {});
  var c   = (eng.engagement_rollup || {}).current  || {};
  var v   = (eng.engagement_rollup || {}).previous || {};
  return [{
    period:                                           'Current',
    engagement_rollup_curr_page_engagements:           Number(c.page_engagements        || 0),
    engagement_rollup_curr_avg_engagements_per_day:    Number(c.avg_engagements_per_day  || 0),
    engagement_rollup_curr_avg_engagements_per_week:   Number(c.avg_engagements_per_week || 0),
    engagement_rollup_prev_page_engagements:           Number(v.page_engagements        || 0),
    engagement_rollup_prev_avg_engagements_per_day:    Number(v.avg_engagements_per_day  || 0),
    engagement_rollup_prev_avg_engagements_per_week:   Number(v.avg_engagements_per_week || 0)
  }];
}

function fb_fetchVideoInsights(p) {
  var j = analyticsGet(buildUrl_fb('overviewVideoInsights', p), p.access_token);
  var v = j.video_insights || {};
  return (v.buckets || []).map(function(date, i) {
    return {
      date:                            date.replace(/-/g, ''),
      video_insights_total_views:      (v.total_views       || [])[i] || 0,
      video_insights_organic_views:    (v.organic_views     || [])[i] || 0,
      video_insights_paid_views:       (v.paid_views        || [])[i] || 0,
      video_insights_total_view_time:  (v.total_view_time   || [])[i] || 0,
      video_insights_organic_view_time:(v.organic_view_time || [])[i] || 0,
      video_insights_paid_view_time:   (v.paid_view_time    || [])[i] || 0,
      video_insights_comments:         (v.comments          || [])[i] || 0,
      video_insights_reactions:        (v.reactions         || [])[i] || 0,
      video_insights_shares:           (v.shares            || [])[i] || 0,
      video_insights_total_posts:      (v.total_posts       || [])[i] || 0
    };
  });
}

function fb_fetchVideoRollup(p) {
  var j = analyticsGet(buildUrl_fb('overviewVideoInsights', p), p.access_token);
  var c = (j.video_rollup || {}).current  || {};
  var v = (j.video_rollup || {}).previous || {};
  return [{
    period:                              'Current',
    video_rollup_curr_total_views:       Number(c.total_views       || 0),
    video_rollup_curr_organic_views:     Number(c.organic_views     || 0),
    video_rollup_curr_paid_views:        Number(c.paid_views        || 0),
    video_rollup_curr_total_view_time:   Number(c.total_view_time   || 0),
    video_rollup_curr_organic_view_time: Number(c.organic_view_time || 0),
    video_rollup_curr_paid_view_time:    Number(c.paid_view_time    || 0),
    video_rollup_curr_comments:          Number(c.comments          || 0),
    video_rollup_curr_reactions:         Number(c.reactions         || 0),
    video_rollup_curr_shares:            Number(c.shares            || 0),
    video_rollup_curr_total_posts:       Number(c.total_posts       || 0),
    video_rollup_prev_total_views:       Number(v.total_views       || 0),
    video_rollup_prev_organic_views:     Number(v.organic_views     || 0),
    video_rollup_prev_paid_views:        Number(v.paid_views        || 0),
    video_rollup_prev_total_view_time:   Number(v.total_view_time   || 0),
    video_rollup_prev_organic_view_time: Number(v.organic_view_time || 0),
    video_rollup_prev_paid_view_time:    Number(v.paid_view_time    || 0),
    video_rollup_prev_comments:          Number(v.comments          || 0),
    video_rollup_prev_reactions:         Number(v.reactions         || 0),
    video_rollup_prev_shares:            Number(v.shares            || 0),
    video_rollup_prev_total_posts:       Number(v.total_posts       || 0)
  }];
}

function fb_fetchReels(p) {
  var j = analyticsGet(buildUrl_fb('overviewReelsAnalytics', p), p.access_token);
  var r = j.reels || {};
  return (r.buckets || []).map(function(date, i) {
    return {
      date:                        date.replace(/-/g, ''),
      reels_initial_plays:         (r.initial_plays         || [])[i] || 0,
      reels_total_reels:           (r.total_reels           || [])[i] || 0,
      reels_total_seconds_watched: (r.total_seconds_watched || [])[i] || 0,
      reels_engagement:            (r.engagement            || [])[i] || 0,
      reels_reactions:             (r.reactions             || [])[i] || 0,
      reels_comments:              (r.comments              || [])[i] || 0,
      reels_shares:                (r.shares                || [])[i] || 0
    };
  });
}

function fb_fetchReelsRollup(p) {
  var j = analyticsGet(buildUrl_fb('overviewReelsAnalytics', p), p.access_token);
  var c = (j.reels_rollup || {}).current  || {};
  var v = (j.reels_rollup || {}).previous || {};
  return [{
    period:                                   'Current',
    reels_rollup_curr_total_reels:            Number(c.total_reels             || 0),
    reels_rollup_curr_initial_plays:          Number(c.initial_plays           || 0),
    reels_rollup_curr_total_seconds_watched:  Number(c.total_seconds_watched   || 0),
    reels_rollup_curr_average_seconds_watched:Number(c.average_seconds_watched || 0),
    reels_rollup_curr_reach:                  Number(c.reach                   || 0),
    reels_rollup_curr_engagement:             Number(c.engagement              || 0),
    reels_rollup_curr_reactions:              Number(c.reactions               || 0),
    reels_rollup_curr_comments:               Number(c.comments                || 0),
    reels_rollup_curr_shares:                 Number(c.shares                  || 0),
    reels_rollup_prev_total_reels:            Number(v.total_reels             || 0),
    reels_rollup_prev_initial_plays:          Number(v.initial_plays           || 0),
    reels_rollup_prev_total_seconds_watched:  Number(v.total_seconds_watched   || 0),
    reels_rollup_prev_average_seconds_watched:Number(v.average_seconds_watched || 0),
    reels_rollup_prev_reach:                  Number(v.reach                   || 0),
    reels_rollup_prev_engagement:             Number(v.engagement              || 0),
    reels_rollup_prev_reactions:              Number(v.reactions               || 0),
    reels_rollup_prev_comments:               Number(v.comments                || 0),
    reels_rollup_prev_shares:                 Number(v.shares                  || 0)
  }];
}

function fb_fetchTopPosts(p) {
  var j = analyticsGet(
    buildUrl_fb('overviewTopPosts', p, '&limit=10&order_by=engagement'),
    p.access_token
  );
  return (j.top_posts || []).map(function(post) {
    return {
      top_posts_created_time:                   toDateStr(post.created_time  || ''),
      top_posts_updated_time:                   toDateStr(post.updated_time  || ''),
      top_posts_saving_time:                    toDateStr(post.saving_time   || ''),
      top_posts_page_name:                      post.page_name               || '',
      top_posts_page_id:                        post.page_id                 || '',
      top_posts_post_id:                        post.post_id                 || '',
      top_posts_permalink:                      post.permalink               || '',
      top_posts_status_type:                    post.status_type             || '',
      top_posts_media_type:                     post.media_type              || '',
      top_posts_video_id:                       post.video_id                || '',
      top_posts_category:                       post.category                || '',
      top_posts_published_by:                   post.published_by            || '',
      top_posts_published_by_url:               post.published_by_url        || '',
      top_posts_shared_from_name:               post.shared_from_name        || '',
      top_posts_shared_from_id:                 post.shared_from_id          || '',
      top_posts_shared_from_link:               post.shared_from_link        || '',
      top_posts_caption:                        post.caption                 || '',
      top_posts_description:                    post.description             || '',
      top_posts_full_picture:                   post.full_picture            || '',
      top_posts_link:                           post.link                    || '',
      top_posts_day_of_week:                    post.day_of_week             || '',
      top_posts_hour_of_day:                    Number(post.hour_of_day      || 0),
      top_posts_message_tags:                   JSON.stringify(post.message_tags  || []),
      top_posts_post_metadata:                  JSON.stringify(post.post_metadata || {}),
      tp_medass_media_id:                       (post.media_assets || []).map(function(a){return a.media_id      || '';}).join(','),
      tp_medass_caption:                        (post.media_assets || []).map(function(a){return a.caption       || '';}).join(','),
      tp_medass_link:                           (post.media_assets || []).map(function(a){return a.link          || '';}).join(','),
      tp_medass_asset_type:                     (post.media_assets || []).map(function(a){return a.assetType     || '';}).join(','),
      tp_medass_call_to_action:                 (post.media_assets || []).map(function(a){return a.callToAction  || '';}).join(','),
      tp_medass_created_at:                     (post.media_assets || []).map(function(a){return a.createdAt     || '';}).join(','),
      top_posts_like:                           Number(post.like                           || 0),
      top_posts_love:                           Number(post.love                           || 0),
      top_posts_haha:                           Number(post.haha                           || 0),
      top_posts_wow:                            Number(post.wow                            || 0),
      top_posts_sad:                            Number(post.sad                            || 0),
      top_posts_angry:                          Number(post.angry                          || 0),
      top_posts_total:                          Number(post.total                          || 0),
      top_posts_shares:                         Number(post.shares                         || 0),
      top_posts_comments:                       Number(post.comments                       || 0),
      top_posts_post_clicks:                    Number(post.post_clicks                    || 0),
      top_posts_total_engagement:               Number(post.total_engagement               || 0),
      top_posts_post_engaged_users:             Number(post.post_engaged_users             || 0),
      top_posts_post_impressions:               Number(post.post_impressions               || 0),
      top_posts_post_impressions_unique:        Number(post.post_impressions_unique        || 0),
      top_posts_post_impressions_paid:          Number(post.post_impressions_paid          || 0),
      top_posts_post_impressions_paid_unique:   Number(post.post_impressions_paid_unique   || 0),
      top_posts_post_impressions_organic:       Number(post.post_impressions_organic       || 0),
      top_posts_post_impressions_organic_unique:Number(post.post_impressions_organic_unique|| 0),
      top_posts_post_impressions_viral:         Number(post.post_impressions_viral         || 0),
      top_posts_post_impressions_viral_unique:  Number(post.post_impressions_viral_unique  || 0),
      top_posts_post_video_views:               Number(post.post_video_views               || 0),
      top_posts_total_impressions:              Number(post.total_impressions              || 0)
    };
  });
}

function fb_fetchActiveUsers(p) {
  var j    = analyticsGet(buildUrl_fb('overviewActiveUsers', p), p.access_token);
  var days = ((j.active_users || {}).active_users_days) || {};
  return (days.buckets || []).map(function(day, i) {
    return {
      active_users_days_buckets:       day,
      active_users_days_values:        (days.values || [])[i]    || 0,
      active_users_days_highest_value: Number(days.highest_value || 0),
      active_users_days_highest_day:   days.highest_day          || ''
    };
  });
}

function fb_fetchDemographics(p, mode) {
  var j = analyticsGet(buildUrl_fb('overviewDemographics', p), p.access_token);
  if (mode === 'age') {
    var fansAge = ((j.audience_age || {}).fans_age) || {};
    return Object.keys(fansAge).map(function(bracket) {
      return { audience_age_fans_age_bracket: bracket, audience_age_fans_age_count: fansAge[bracket] || 0 };
    });
  }
  var genderMap = j.audience_gender || {};
  return Object.keys(genderMap).map(function(g) {
    return { audience_gender_gender: g, audience_gender_count: genderMap[g] || 0 };
  });
}

function fb_fetchAudienceLocation(p, mode) {
  var j = analyticsGet(buildUrl_fb('overviewAudienceLocation', p), p.access_token);
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
