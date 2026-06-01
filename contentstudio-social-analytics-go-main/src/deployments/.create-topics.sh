#!/bin/bash

TOPICS=(
  immediate-work-order-facebook
  work-order-facebook
  raw-facebook-posts
  raw-facebook-videos
  raw-facebook-insights
  parsed-facebook-posts
  parsed-facebook-media-assets
  parsed-facebook-video-insights
  parsed-facebook-reels-insights
  parsed-facebook-insights
  immediate-work-order-instagram
  work-order-instagram
  raw-instagram-media
  raw-instagram-insights
  parsed-instagram-posts
  parsed-instagram-insights
  immediate-work-order-linkedin
  raw-linkedin-posts
  raw-linkedin-images
  raw-linkedin-videos
  raw-linkedin-stats
  raw-linkedin-insights
  raw-linkedin-organization
  parsed-linkedin-posts
  parsed-linkedin-media-assets
  parsed-linkedin-stats
  parsed-linkedin-insights
  work-order-linkedin
  # LinkedIn Page topics
  work-order-linkedin-page
  work-order-linkedin-page-batch
  raw-linkedin-page-posts
  raw-linkedin-page-insights
  raw-linkedin-page-organization
  parsed-linkedin-page-posts
  parsed-linkedin-page-insights
  # LinkedIn Profile topics
  work-order-linkedin-profile
  work-order-linkedin-profile-batch
  raw-linkedin-profile-posts
  raw-linkedin-profile-insights
  parsed-linkedin-profile-posts
  parsed-linkedin-profile-insights
  # Competitor topics
  competitor-work-order-facebook
  competitor-work-order-instagram
  competitor-work-order-facebook-batch
  competitor-work-order-instagram-batch
  # YouTube topics
  immediate-work-order-youtube
  work-order-youtube
  work-order-youtube-batch
  raw-youtube-channels
  raw-youtube-videos
  raw-youtube-activity-insights
  raw-youtube-traffic-insights
  raw-youtube-shared-insights
  # TikTok immediate work orders
  immediate-work-order-tiktok
  work-order-tiktok-batch
  # TikTok raw data topics
  raw-tiktok-posts
  raw-tiktok-insights
  # Twitter topics
  immediate-work-order-twitter
  work-order-twitter-batch
  raw-twitter-posts
  raw-twitter-insights
  # Pinterest topics
  immediate-work-order-pinterest
  work-order-pinterest
  raw-pinterest-users
  raw-pinterest-boards
  raw-pinterest-pins
  raw-pinterest-pin-insights
  raw-pinterest-user-insights
  parsed-pinterest-users
  parsed-pinterest-boards
  parsed-pinterest-pins
  parsed-pinterest-pin-insights
  parsed-pinterest-user-insights
  # GMB topics
  immediate-work-order-gmb
  work-order-gmb
  raw-gmb-data
  # Unified work order topics (supports all platforms)
  unified-work-order
  unified-work-order-batch
  # Social listening pipeline topics
  listening-work
  listening-raw
  listening-parsed
  listening-enriched
  listening-dlq
  # Meta Ads topics
  immediate-work-order-meta-ads
  work-order-meta-ads
  raw-meta-ads-account-info
  raw-meta-ads-campaigns
  raw-meta-ads-adsets
  raw-meta-ads-ads
  raw-meta-ads-campaign-insights
  raw-meta-ads-adset-insights
  raw-meta-ads-ad-insights
  raw-meta-ads-demographics-age-gender
  raw-meta-ads-demographics-device-platform
  raw-meta-ads-demographics-region-country
)

# Wait for Kafka to start up
echo "Waiting for Kafka to be ready..."
sleep 10  # crude but effective for local setups

for topic in "${TOPICS[@]}"; do
  echo "Creating topic: $topic"
  /opt/kafka/bin/kafka-topics.sh \
    --create \
    --if-not-exists \
    --topic "$topic" \
    --bootstrap-server kafka:9092 \
    --partitions 1 \
    --replication-factor 1
done

echo "✅ Topic creation complete."
exit 0
