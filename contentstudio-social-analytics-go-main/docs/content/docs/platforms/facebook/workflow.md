---
title: Workflow
description: Documentation for workflow
---

## Overview

This document provides a comprehensive guide to the Facebook analytics data processing pipeline implemented in the ContentStudio Social Analytics Go project. The workflow extracts, processes, analyzes, and stores Facebook page data including posts, videos, insights, and media assets with production-grade performance and reliability.

## Architecture

The Facebook workflow follows a microservices architecture with event-driven communication via Kafka, featuring parallel processing, batching, and multi-stage data transformation:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  Account        │    │  Facebook        │    │  Facebook       │    │  Facebook       │    │  Facebook       │
│  Fetcher        │───▶│  Fetcher         │───▶│  Parser         │───▶│  Immediate      │───▶│  ClickHouse     │
│                 │    │                  │    │                 │    │  Processor      │    │  Sink           │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │                       │                       │
         ▼                       ▼                       ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ work-order-     │    │ raw-facebook-*   │    │ parsed-facebook │    │ processed-       │    │ ClickHouse      │
│ facebook        │    │ topics           │    │ topics          │    │ facebook-topics  │    │ Database        │
└─────────────────┘    └──────────────────┘    └─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Complete Data Pipeline Flow

```
📋 SCHEDULING
   Account Fetcher → work-order-facebook

📥 EXTRACTION  
   work-order-facebook → Facebook Fetcher → {
     raw-facebook-posts,
     raw-facebook-videos, 
     raw-facebook-insights
   }

🔄 PARSING
   raw-facebook-* → Facebook Parser → {
     parsed-facebook-posts,
     parsed-facebook-video-insights,
     parsed-facebook-insights, 
     parsed-facebook-media-assets,
     parsed-facebook-reels-insights
   }

⚡ IMMEDIATE PROCESSING
   parsed-facebook-* → Facebook Immediate Processor → {
     processed-facebook-posts,
     processed-facebook-video-insights,
     processed-facebook-insights,
     processed-facebook-media-assets,
     processed-facebook-reels-insights
   }

💾 BATCH STORAGE
   processed-facebook-* → Facebook ClickHouse Sink → ClickHouse Database
                                     ↓
                        ⚡ High-Performance Batching ⚡
                        • 1000-item batches or 15-second timeout
                        • 15 parallel processors (3 per data type)
                        • Zero data loss with backpressure handling
                        • Real-time channel utilization monitoring
```

## Components

… (content preserved from original `docs/facebook-workflow.md`, including batch architecture, message formats, config, performance, and deployment notes)
