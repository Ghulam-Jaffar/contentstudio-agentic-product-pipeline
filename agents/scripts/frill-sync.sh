#!/usr/bin/env bash
# Frill idea sync — deterministic fetch + snapshot layer for the /frill pipeline.
#
# Read-only against Frill. Never mutates the board.
# The raw snapshot contains customer emails and private ideas, so it lives in
# docs/frill/cache/ which is gitignored. Never commit it, never quote private
# ideas or author emails into a deliverable.
#
# Usage:
#   frill-sync.sh sync [--full]      Update the local snapshot (default: incremental)
#   frill-sync.sh meta               Refresh statuses + topics lookup
#   frill-sync.sh candidates [N]     Print the ranked actionable pool (default 30)
#   frill-sync.sh comments <idea>    Fetch comments for one idea (idx, slug or URL)
#   frill-sync.sh show <idea>        Print one idea as JSON (idx, slug or URL)
#   frill-sync.sh stats              Snapshot summary
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
CACHE="$ROOT/docs/frill/cache"
IDEAS="$CACHE/ideas.ndjson"
STATE="$CACHE/sync-state.json"
META="$CACHE/meta.json"
LEDGER="$ROOT/docs/frill/ledger.json"
API="https://api.frill.co/v1"
PAGE_LIMIT=100          # hard API cap; larger values are silently clamped
THROTTLE=0.3            # no documented rate limit, so be polite
MAX_PAGES=60            # runaway guard

mkdir -p "$CACHE"

die() { printf 'frill-sync: %s\n' "$*" >&2; exit 1; }
note() { printf '  %s\n' "$*" >&2; }

# --- credentials -------------------------------------------------------------
# Prefer the environment; fall back to .env. The key is never printed.
load_key() {
  if [[ -n "${FRILL_API_KEY:-}" ]]; then return; fi
  [[ -f "$ROOT/.env" ]] || die "no FRILL_API_KEY in env and no .env file at repo root"
  FRILL_API_KEY="$(grep -E '^FRILL_API_KEY=' "$ROOT/.env" | tail -1 | cut -d= -f2- | tr -d '"'"'"' \r\n')"
  [[ -n "$FRILL_API_KEY" ]] || die "FRILL_API_KEY missing or empty in .env"
}

# --- transport ---------------------------------------------------------------
# GET with retry/backoff. Handles 429 (undocumented limit) and 5xx.
# Fails loudly on 401/403 so a rotated key never looks like an empty board.
api_get() {
  local path="$1" attempt=1 max=5 body code wait
  local tmp; tmp="$(mktemp)"
  while :; do
    code="$(curl -sS -o "$tmp" -w '%{http_code}' \
      -H "Authorization: Bearer $FRILL_API_KEY" \
      -H 'Accept: application/json' \
      --max-time 45 "$API/$path" || echo 000)"
    case "$code" in
      200) cat "$tmp"; rm -f "$tmp"; return 0 ;;
      401|403) rm -f "$tmp"; die "Frill returned $code — the API key is invalid, expired or lacks scope. Rotate it in Frill and update .env. (key not printed)" ;;
      404) rm -f "$tmp"; die "Frill returned 404 for /$path — endpoint does not exist" ;;
      429|500|502|503|504|000)
        if (( attempt >= max )); then rm -f "$tmp"; die "Frill returned $code for /$path after $max attempts — aborting without touching the snapshot"; fi
        wait=$(( 2 ** attempt ))
        note "HTTP $code on /$path — retrying in ${wait}s (attempt $attempt/$max)"
        sleep "$wait"; attempt=$(( attempt + 1 )) ;;
      *) body="$(head -c 300 "$tmp")"; rm -f "$tmp"; die "Frill returned $code for /$path: $body" ;;
    esac
  done
}

# --- ordering assertion ------------------------------------------------------
# Frill SILENTLY IGNORES unknown sortBy values: sortBy=-vote_count returns HTTP
# 200 in plain created_at order. A typo would quietly hand us the wrong ideas,
# so every sorted fetch is verified to actually be non-increasing.
assert_desc() {
  local file="$1" field="$2"
  jq -e --arg f "$field" '
    [.[] | (.[$f] // 0)] as $v
    | ($v == ($v | sort | reverse))
  ' "$file" >/dev/null 2>&1 \
    || die "Frill ignored sortBy=$field (response was not sorted). The API contract changed — refusing to process possibly-wrong ideas."
}

# --- sync --------------------------------------------------------------------
# Default order is created_at desc, so new ideas are always at the front and an
# incremental sync is normally a single request (~16 new ideas/month).
# Incremental only discovers NEW ideas; vote counts and statuses on older ideas
# go stale, so --full refreshes everything (14 pages) and stats warns when the
# last full sync is over 7 days old.
cmd_sync() {
  local full=0
  [[ "${1:-}" == "--full" ]] && full=1
  load_key

  local known; known="$(mktemp)"
  if [[ -s "$IDEAS" && $full -eq 0 ]]; then
    jq -r '.idx' "$IDEAS" | sort -u > "$known"
  else
    : > "$known"
    [[ $full -eq 1 ]] && note "full re-sync requested"
  fi

  local fetched; fetched="$(mktemp)"
  local cursor="" page=0 total="?" newcount=0

  while (( page < MAX_PAGES )); do
    page=$(( page + 1 ))
    local q="ideas?limit=$PAGE_LIMIT"
    [[ -n "$cursor" ]] && q="$q&after=$cursor"
    local resp; resp="$(api_get "$q")"

    # A malformed page must not corrupt a good snapshot.
    jq -e '.data | type == "array"' <<<"$resp" >/dev/null 2>&1 \
      || die "unexpected response shape on page $page — snapshot left unchanged"

    total="$(jq -r '.pagination.total // "?"' <<<"$resp")"
    local got; got="$(jq -r '.pagination.count // 0' <<<"$resp")"
    jq -c '.data[]' <<<"$resp" >> "$fetched"

    # Ideas created mid-pagination shift the window, so dedupe by idx later.
    local fresh
    fresh="$(jq -r '.data[].idx' <<<"$resp" | sort -u | comm -23 - "$known" | wc -l | tr -d ' ')"
    newcount=$(( newcount + fresh ))
    note "page $page: $got fetched, $fresh new (board total $total)"

    local has; has="$(jq -r '.pagination.hasNextPage // false' <<<"$resp")"
    cursor="$(jq -r '.pagination.endCursor // ""' <<<"$resp")"
    [[ "$has" == "true" ]] || break
    [[ -z "$cursor" ]] && break
    # Incremental: stop once a whole page contained nothing new.
    if (( full == 0 )) && (( fresh == 0 )); then
      note "no new ideas on this page — stopping incremental sync"
      break
    fi
    sleep "$THROTTLE"
  done

  (( page >= MAX_PAGES )) && note "WARNING: hit MAX_PAGES=$MAX_PAGES guard"

  # Merge: fetched rows win over cached ones. Dedupe by idx, newest first.
  # Written to a temp file and moved into place so an abort keeps the last good snapshot.
  local merged; merged="$(mktemp)"
  cat "$IDEAS" 2>/dev/null > "$merged" || true
  cat "$fetched" >> "$merged"
  local out; out="$(mktemp)"
  jq -s 'group_by(.idx) | map(.[-1]) | sort_by(.created_at) | reverse | .[]' -c "$merged" > "$out"
  mv "$out" "$IDEAS"

  local cached; cached="$(wc -l < "$IDEAS" | tr -d ' ')"
  local now; now="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  local lastfull
  if (( full == 1 )); then lastfull="$now"; else
    lastfull="$(jq -r '.last_full_sync // ""' "$STATE" 2>/dev/null || echo "")"
  fi
  jq -n --arg now "$now" --arg lf "$lastfull" --argjson cached "$cached" \
        --argjson new "$newcount" --arg total "$total" --argjson pages "$page" '{
    last_sync: $now, last_full_sync: (if $lf == "" then null else $lf end),
    cached_ideas: $cached, new_this_sync: $new, board_total: $total, pages_fetched: $pages
  }' > "$STATE"

  rm -f "$known" "$fetched" "$merged"
  printf 'synced: %s new, %s cached (board reports %s)\n' "$newcount" "$cached" "$total"
  cmd_meta_quiet
}

# --- statuses + topics -------------------------------------------------------
cmd_meta_quiet() {
  load_key
  local s t out; out="$(mktemp)"
  s="$(api_get 'statuses?limit=100')"; sleep "$THROTTLE"
  t="$(api_get 'topics?limit=100')"
  jq -n --argjson s "$s" --argjson t "$t" \
    '{statuses: [$s.data[] | {idx,name,is_completed}], topics: [$t.data[] | {idx,name}]}' > "$out"
  mv "$out" "$META"
}
cmd_meta() { cmd_meta_quiet; jq . "$META"; }

# --- triage ------------------------------------------------------------------
# The actionable pool. Excludes, in order:
#   completed/shipped   — already built
#   "Unlikely to build" — already declined
#   archived / unapproved
#   private             — 41 ideas that must never reach a deliverable
#   ledger'd            — already processed by a previous /frill run
# 615 ideas carry NO status at all, so a null status is treated as untriaged and
# kept in the pool rather than dropped.
pool() {
  [[ -s "$IDEAS" ]] || die "no snapshot yet — run: frill-sync.sh sync --full"
  local seen; seen="$(mktemp)"
  if [[ -f "$LEDGER" ]]; then
    jq -r '(.processed // {}) | keys[]' "$LEDGER" 2>/dev/null | sort -u > "$seen" || : > "$seen"
  else : > "$seen"; fi
  jq -c --slurpfile _ /dev/null --rawfile seen "$seen" '
    ($seen | split("\n") | map(select(length>0))) as $done
    | select(.is_completed != true)
    | select(.is_archived != true)
    | select(.is_private != true)
    | select(.approval_status == "approved")
    | select((.status.is_completed // false) != true)
    | select(((.status.name // "") | test("Unlikely")) | not)
    | select((.idx | IN($done[])) | not)
  ' "$IDEAS"
}

# Ranking. Votes are a weak signal on this board (median ~1, so a raw vote sort
# just surfaces 2021 ideas), and priority_score exists for only 56 ideas — so
# use the team's own score when present and otherwise blend log-scaled votes,
# discussion volume, recency and committed-status.
#
# NOTE: priority_score_normalized is a 0-100 scale on this board (observed
# 1..103), NOT 0..1 — it is divided by 100 before weighting. Getting this wrong
# lets 3 prioritized ideas dominate the entire ranking.
#
# _flags carries deterministic triage warnings so a known-dead ask can never be
# silently promoted to a pipeline. They are advisory: the /frill command decides.
cmd_candidates() {
  local n="${1:-30}"
  pool | jq -s --argjson n "$n" '
    (now) as $now
    | map(
        ((.created_at | sub("\\.[0-9]+Z$"; "Z") | fromdateiso8601? ) // 0) as $ts
        | (($now - $ts) / 86400) as $age_days
        | (if $ts == 0 then 0 else (3 - ($age_days / 365)) | if . < 0 then 0 else . end end) as $recency
        | ((.status.name // "") | if test("In Development") then 3 elif test("Planned") then 2 elif test("Under Review") then 1 else 0 end) as $status_w
        | (((.priority_score_normalized // 0) / 100) * 15) as $prio
        | ((.vote_count // 0) + 1 | log2 * 3) as $votes
        | ((.comment_count // 0) + 1 | log2 * 2) as $disc
        | ((.name + " " + (.content_markdown // "")) | ascii_downcase) as $txt
        | ([
            # Out of scope per CLAUDE.md: ContentStudio supports neither.
            (if ($txt | test("dark mode|dark theme")) then "OUT-OF-SCOPE:no-dark-mode" else empty end),
            (if ($txt | test("\\brtl\\b|right.to.left")) then "OUT-OF-SCOPE:no-rtl" else empty end),
            # Blog publishing was sunset; exclude blog destinations/post types.
            (if ((.name | ascii_downcase) | test("blog|contentpen|wordpress"))
                or ($txt | test("blog composer|blog post|blog publish|publish(ing)? to (my |a |the )?blog|blog destination|contentpen|wordpress"))
              then "SUNSET:blog-publishing" else empty end),
            # Bug reports on a feature board: only 5 carry is_bug, topic Bug has 60.
            (if (.is_bug == true) or ([.topics[]?.name] | any(test("Bug"))) or ($txt | test("repair|broken|not working|doesn.t work|bug\\b|error\\b|fails? to")) then "LIKELY-BUG:not-a-feature" else empty end),
            # Median body is ~200 chars; anything this thin cannot be specced as-is.
            (if ((.content_markdown // "") | length) < 80 then "THIN:enrich-from-comments" else empty end),
            # 2021-era ideas need revalidation against the current product.
            (if $age_days > 1095 and (((.status.name // "") | test("Planned|In Development")) | not)
              then "STALE:revalidate" else empty end),
            # Compound asks must be split before routing.
            (if (.name | test(";")) or ($txt | test("\\band also\\b|\\benhance\\b.*\\band\\b.*\\brevamp\\b")) then "COMPOUND:split-first" else empty end),
            # Non-ASCII beyond emoji suggests a non-English report needing translation.
            (if (.name | test("[\u0400-\u04FF\u0100-\u017F\u00C0-\u00FF]")) then "NON-ENGLISH:translate" else empty end)
          ]) as $flags
        | . + { _score: (($prio + $votes + $disc + $recency + $status_w) * 100 | round / 100), _flags: $flags }
      )
    | sort_by(-._score) | .[:$n]
    | map({score: ._score, flags: ._flags, idx, name, votes: .vote_count, comments: .comment_count,
           status: (.status.name // "untriaged"), topics: [.topics[]?.name],
           created: .created_at[0:10], chars: ((.content_markdown // "") | length), url})
  '
}

cmd_show() {
  local ref; ref="$(normalize_ref "${1:?idea idx, slug or URL required}")"
  [[ -s "$IDEAS" ]] || die "no snapshot yet — run: frill-sync.sh sync --full"
  jq -e --arg r "$ref" 'select(.idx == $r or .slug == $r)' "$IDEAS" \
    || die "no cached idea matching '$ref' — try: frill-sync.sh sync --full"
}

# Accepts idea_xxxx, a board slug, or a full contentstudio.frill.co/board/<slug> URL.
normalize_ref() {
  local r="$1"
  r="${r%/}"; r="${r##*/}"; r="${r%%\?*}"
  printf '%s' "$r"
}

# Comments are fetched on demand, not cached: bodies average ~200 chars, so the
# comment thread is usually where the actual requirement lives.
cmd_comments() {
  load_key
  local ref; ref="$(normalize_ref "${1:?idea idx, slug or URL required}")"
  local idx="$ref"
  if [[ "$ref" != idea_* ]]; then
    idx="$(jq -r --arg r "$ref" 'select(.slug == $r) | .idx' "$IDEAS" 2>/dev/null | head -1)"
    [[ -n "$idx" ]] || die "could not resolve '$ref' to an idea idx from the snapshot"
  fi
  local cursor="" page=0 out; out="$(mktemp)"
  while (( page < 20 )); do
    page=$(( page + 1 ))
    local q="comments?idea_idx=$idx&limit=100"
    [[ -n "$cursor" ]] && q="$q&after=$cursor"
    local resp; resp="$(api_get "$q")"
    jq -c '(.data // [])[]' <<<"$resp" >> "$out"
    local has; has="$(jq -r '.pagination.hasNextPage // false' <<<"$resp")"
    cursor="$(jq -r '.pagination.endCursor // ""' <<<"$resp")"
    [[ "$has" == "true" && -n "$cursor" ]] || break
    sleep "$THROTTLE"
  done
  cat "$out"; rm -f "$out"
}

cmd_stats() {
  [[ -s "$IDEAS" ]] || die "no snapshot yet — run: frill-sync.sh sync --full"
  echo "=== snapshot ==="; jq . "$STATE" 2>/dev/null || echo "(no sync state)"
  local lf; lf="$(jq -r '.last_full_sync // ""' "$STATE" 2>/dev/null || echo "")"
  if [[ -z "$lf" || "$lf" == "null" ]]; then
    echo "WARNING: no full sync on record. Vote counts and statuses may be stale — run: sync --full"
  else
    local age=$(( ( $(date -u +%s) - $(date -u -d "$lf" +%s 2>/dev/null || echo 0) ) / 86400 ))
    (( age > 7 )) && echo "WARNING: last full sync was ${age}d ago; votes/statuses on older ideas are stale — run: sync --full"
  fi
  echo "=== cached: $(wc -l < "$IDEAS" | tr -d ' ') ideas ==="
  echo "=== actionable pool: $(pool | wc -l | tr -d ' ') ==="
  echo "=== by status ==="; jq -r '.status.name // "(untriaged)"' "$IDEAS" | sort | uniq -c | sort -rn
  echo "=== excluded ==="
  printf '  shipped/completed: %s\n' "$(jq -r 'select(.is_completed==true or (.status.is_completed//false)==true)|.idx' "$IDEAS" | wc -l | tr -d ' ')"
  printf '  private (never quote): %s\n' "$(jq -r 'select(.is_private==true)|.idx' "$IDEAS" | wc -l | tr -d ' ')"
  printf '  archived: %s\n' "$(jq -r 'select(.is_archived==true)|.idx' "$IDEAS" | wc -l | tr -d ' ')"
  printf '  unlikely to build: %s\n' "$(jq -r 'select((.status.name//"")|test("Unlikely"))|.idx' "$IDEAS" | wc -l | tr -d ' ')"
  printf '  already processed (ledger): %s\n' "$(jq -r '(.processed // {})|keys|length' "$LEDGER" 2>/dev/null || echo 0)"
}

case "${1:-}" in
  sync)       shift; cmd_sync "${1:-}" ;;
  meta)       cmd_meta ;;
  candidates) shift; cmd_candidates "${1:-30}" ;;
  comments)   shift; cmd_comments "${1:-}" ;;
  show)       shift; cmd_show "${1:-}" ;;
  stats)      cmd_stats ;;
  *) sed -n '2,20p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; exit 1 ;;
esac
