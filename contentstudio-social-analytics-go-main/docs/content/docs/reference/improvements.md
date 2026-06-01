---
title: Improvements
description: Documentation for improvements
---

This document lists areas for future improvement and refactoring in the project.

1) MongoDB Timestamp Format for Facebook Accounts
- Issue: Several timestamps are stored as strings without timezone (e.g., `2025-06-08T12:01:38.683882`), causing parsing errors in Go.
- Workaround: Use `MongoTime` wrapper or `*string` for legacy fields.
- Recommended: Migrate to BSON Date or full ISO 8601 with timezone.

2) Account Type Casing
- Issue: `accountType` casing differs (e.g., "page" vs "Page").
- Improvement: Centralize mapping/config to avoid ad-hoc conditionals.

3) Sensitive Logging
- Ensure no tokens/PII are logged; continue using generic messages for token updates.

4) Repository Error Handling
- Standardize on app-specific `ErrNotFound` to abstract DB details from callers.

5) FacebookAccount Struct Gaps
- Decide whether `Name` and `Email` belong to the struct or `ExtraData`.
