# Build stage
FROM golang:1.25-alpine AS builder

# Install necessary build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files first (for better caching)
COPY src/go.mod src/go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY src/ .
#COPY src/.env /app/.env

# Build all binaries with optimization flags
# -ldflags="-s -w": omit symbol table and DWARF debug info for smaller binaries
RUN go build -ldflags="-s -w" -o /bin/account_fetcher ./cmd/jobs && \
    go build -ldflags="-s -w" -o /bin/competitor_fetcher ./cmd/competitor-jobs && \
    go build -ldflags="-s -w" -o /bin/competitor_insights_filler ./cmd/competitor-insights-filler && \
    go build -ldflags="-s -w" -o /bin/url_refresher ./cmd/jobs/url-refresher/ && \
    go build -ldflags="-s -w" -o /bin/validate_tokens ./cmd/validate-tokens && \
    # Facebook services
    go build -ldflags="-s -w" -o /bin/facebook_fetcher ./services/facebook/facebook-fetcher && \
    go build -ldflags="-s -w" -o /bin/facebook_posts_videos_parser ./services/facebook/facebook-posts-parser && \
    go build -ldflags="-s -w" -o /bin/facebook_immediate_processor ./services/facebook/facebook-immediate-processor && \
    go build -ldflags="-s -w" -o /bin/facebook_clickhouse_sink ./services/facebook/facebook-clickhouse-sink && \
    go build -ldflags="-s -w" -o /bin/facebook_competitor_analysis ./services/facebook/facebook-competitor-analysis && \
    # Instagram services
    go build -ldflags="-s -w" -o /bin/instagram_fetcher ./services/instagram/instagram-fetcher && \
    go build -ldflags="-s -w" -o /bin/instagram_parser ./services/instagram/instagram-posts-parser && \
    go build -ldflags="-s -w" -o /bin/instagram_immediate_processor ./services/instagram/instagram-immediate-processor && \
    go build -ldflags="-s -w" -o /bin/instagram_clickhouse_sink ./services/instagram/instagram-clickhouse-sink && \
    go build -ldflags="-s -w" -o /bin/instagram_competitor_analysis ./services/instagram/instagram-competitor-analysis && \
    # LinkedIn services
    go build -ldflags="-s -w" -o /bin/linkedin_fetcher ./services/linkedin/linkedin-fetcher && \
    go build -ldflags="-s -w" -o /bin/linkedin_parser ./services/linkedin/linkedin-parser && \
    go build -ldflags="-s -w" -o /bin/linkedin_immediate_processor ./services/linkedin/linkedin-immediate-processor && \
    go build -ldflags="-s -w" -o /bin/linkedin_clickhouse_sink ./services/linkedin/linkedin-clickhouse-sink && \
    # TikTok services
    go build -ldflags="-s -w" -o /bin/tiktok_fetcher ./services/tiktok/tiktok-fetcher && \
    # Twitter services
    go build -ldflags="-s -w" -o /bin/twitter_fetcher ./services/twitter/twitter-fetcher && \
    # GMB services
    go build -ldflags="-s -w" -o /bin/gmb_fetcher ./services/gmb/gmb-fetcher && \
    # YouTube services
    go build -ldflags="-s -w" -o /bin/youtube_fetcher ./services/youtube/youtube-fetcher && \
    go build -ldflags="-s -w" -o /bin/youtube_immediate_processor ./services/youtube/youtube-immediate-processor && \
    go build -ldflags="-s -w" -o /bin/youtube_analytics_sink ./services/youtube/youtube-analytics-sink && \
    # Pinterest services
    go build -ldflags="-s -w" -o /bin/pinterest_fetcher ./services/pinterest/pinterest-fetcher && \
    go build -ldflags="-s -w" -o /bin/pinterest_parser ./services/pinterest/pinterest-parser && \
    go build -ldflags="-s -w" -o /bin/pinterest_immediate_processor ./services/pinterest/pinterest-immediate-processor && \
    go build -ldflags="-s -w" -o /bin/pinterest_analytics_sink ./services/pinterest/pinterest-analytics-sink && \
    # Meta Ads services
    go build -ldflags="-s -w" -o /bin/meta_ads_fetcher ./services/meta-ads/meta-ads-fetcher && \
    # API services
    go build -ldflags="-s -w" -o /bin/immediate_work_api ./cmd/api-server/ && \
    # Unified services
    go build -ldflags="-s -w" -o /bin/unified_account_fetcher ./services/unified/account-fetcher && \
    go build -ldflags="-s -w" -o /bin/unified_immediate_processor ./services/unified/immediate-processor && \
    # Analytics sink services (merged parser+sink)
    go build -ldflags="-s -w" -o /bin/instagram_analytics_sink ./services/instagram/instagram-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/facebook_analytics_sink ./services/facebook/facebook-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/linkedin_analytics_sink ./services/linkedin/linkedin-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/tiktok_analytics_sink ./services/tiktok/tiktok-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/twitter_analytics_sink ./services/twitter/twitter-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/gmb_analytics_sink ./services/gmb/gmb-analytics-sink && \
    go build -ldflags="-s -w" -o /bin/meta_ads_sink ./services/meta-ads/meta-ads-sink && \
    go build -ldflags="-s -w" -o /bin/listening_scheduler ./services/listening/listening-scheduler && \
    # Listening consumer (fetcher + parser + sink + async enrichment pipeline)
    go build -ldflags="-s -w" -o /bin/listening_consumer ./services/listening/listening-consumer && \
    # Listening API (HTTP server for listening endpoints)
    go build -ldflags="-s -w" -o /bin/listening_api ./services/listening/listening-api

# Final stage - minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN adduser -D -g '' appuser

# Create app directory
WORKDIR /app
#COPY src/.env /app/.env

# Copy wait script for docker-compose orchestration
COPY src/deployments/.wait-for-services.sh /wait-for-services.sh
RUN chmod +x /wait-for-services.sh

# Copy binaries from builder stage
# Scheduler services
COPY --from=builder /bin/account_fetcher /usr/local/bin/
COPY --from=builder /bin/url_refresher /usr/local/bin/
COPY --from=builder /bin/competitor_fetcher /usr/local/bin/
COPY --from=builder /bin/competitor_insights_filler /usr/local/bin/
COPY --from=builder /bin/validate_tokens /usr/local/bin/

# Facebook services
COPY --from=builder /bin/facebook_fetcher /usr/local/bin/
COPY --from=builder /bin/facebook_posts_videos_parser /usr/local/bin/
COPY --from=builder /bin/facebook_immediate_processor /usr/local/bin/
COPY --from=builder /bin/facebook_clickhouse_sink /usr/local/bin/
COPY --from=builder /bin/facebook_competitor_analysis /usr/local/bin/

# Instagram services
COPY --from=builder /bin/instagram_fetcher /usr/local/bin/
COPY --from=builder /bin/instagram_parser /usr/local/bin/
COPY --from=builder /bin/instagram_immediate_processor /usr/local/bin/
COPY --from=builder /bin/instagram_clickhouse_sink /usr/local/bin/
COPY --from=builder /bin/instagram_competitor_analysis /usr/local/bin/

# LinkedIn services
COPY --from=builder /bin/linkedin_fetcher /usr/local/bin/
COPY --from=builder /bin/linkedin_parser /usr/local/bin/
COPY --from=builder /bin/linkedin_immediate_processor /usr/local/bin/
COPY --from=builder /bin/linkedin_clickhouse_sink /usr/local/bin/

# TikTok services
COPY --from=builder /bin/tiktok_fetcher /usr/local/bin/

# Twitter services
COPY --from=builder /bin/twitter_fetcher /usr/local/bin/

# GMB services
COPY --from=builder /bin/gmb_fetcher /usr/local/bin/

# YouTube services
COPY --from=builder /bin/youtube_fetcher /usr/local/bin/
COPY --from=builder /bin/youtube_immediate_processor /usr/local/bin/
COPY --from=builder /bin/youtube_analytics_sink /usr/local/bin/

# Pinterest services
COPY --from=builder /bin/pinterest_fetcher /usr/local/bin/
COPY --from=builder /bin/pinterest_parser /usr/local/bin/
COPY --from=builder /bin/pinterest_immediate_processor /usr/local/bin/
COPY --from=builder /bin/pinterest_analytics_sink /usr/local/bin/

# Meta Ads services
COPY --from=builder /bin/meta_ads_fetcher /usr/local/bin/

# API services
COPY --from=builder /bin/immediate_work_api /usr/local/bin/

# Unified services
COPY --from=builder /bin/unified_account_fetcher /usr/local/bin/
COPY --from=builder /bin/unified_immediate_processor /usr/local/bin/

# Analytics sink services
COPY --from=builder /bin/instagram_analytics_sink /usr/local/bin/
COPY --from=builder /bin/facebook_analytics_sink /usr/local/bin/
COPY --from=builder /bin/linkedin_analytics_sink /usr/local/bin/
COPY --from=builder /bin/tiktok_analytics_sink /usr/local/bin/
COPY --from=builder /bin/twitter_analytics_sink /usr/local/bin/
COPY --from=builder /bin/gmb_analytics_sink /usr/local/bin/
COPY --from=builder /bin/meta_ads_sink /usr/local/bin/

# Listening services
COPY --from=builder /bin/listening_scheduler /usr/local/bin/
COPY --from=builder /bin/listening_consumer /usr/local/bin/
COPY --from=builder /bin/listening_api /usr/local/bin/


# Set ownership to non-root user
RUN chown -R appuser:appuser /usr/local/bin/

# Switch to non-root user
USER appuser

# Expose ports for HTTP APIs
EXPOSE 8080
EXPOSE 7085

# Default command shows available binaries
CMD ["echo", "Specify binary: account_fetcher, url_refresher, competitor_fetcher, facebook_fetcher, instagram_fetcher, linkedin_fetcher, tiktok_fetcher, twitter_fetcher, youtube_fetcher, pinterest_fetcher, unified_account_fetcher, unified_immediate_processor, or use -h for help"]

