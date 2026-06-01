---
title: Immediate Work API
description: Documentation for immediate work API
---

Purpose
- Exposes an internal API to submit “immediate work orders” into Kafka, triggering near-real-time processing for a single account.

Endpoints (example)
- `POST /v1/immediate/facebook` → emits to `immediate-work-order-facebook`
- `POST /v1/immediate/instagram` → emits to `immediate-work_order-instagram`
- `POST /v1/immediate/linkedin` → emits to `immediate-work_order-linkedin`

Request Body (example)
```json
{
  "workspace_id": "...",
  "account_id": "...",
  "sync_type": "incremental"
}
```

Run
```bash
cd src && make immediate_work_api
../bin/immediate_work_api
```

Notes
- Validate input; include basic auth or internal auth; return request-id and Kafka metadata for traceability.
