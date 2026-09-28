#!/usr/bin/env bash
# Frill changelog writer — creates ONE announcement as a DRAFT.
#
# Deliberately separate from frill-sync.sh, which is read-only and must stay
# that way. This is the only path in the repo that writes to Frill.
#
# HARD RULES, enforced below, not left to the caller:
#   * published_at is ALWAYS null. This script cannot publish. A human presses
#     publish in the Frill admin UI. There is no --publish flag and there must
#     never be one.
#   * An idea's STATUS may be changed, and nothing else. No creating ideas, no
#     editing an idea's body, no deleting, no commenting. A status change goes
#     out only on the PO's explicit approval, never automatically.
#     NOTE: this emails everyone who voted. That is intended: it tells the
#     customer their request shipped.
#
# Usage:
#   frill-announce.sh draft --name "<title>" --content-file <path> --author <user_idx> \
#       [--categories idx,idx] [--ideas idx,idx] [--dry-run]
#   frill-announce.sh status <idea_idx> <status_idx> [--dry-run]
#
# Status idxs come from: frill-sync.sh meta
#
# NOTE: Frill REQUIRES author_idx on create, despite its docs listing it as optional.
# Find it on any existing announcement: frill-sync.sh announcements 1 | jq -r .data[0].author.idx
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
API="https://api.frill.co/v1"

die() { printf 'frill-announce: %s\n' "$*" >&2; exit 1; }
note() { printf '  %s\n' "$*" >&2; }

load_key() {
  if [[ -n "${FRILL_API_KEY:-}" ]]; then return; fi
  [[ -f "$ROOT/.env" ]] || die "no FRILL_API_KEY in env and no .env file at repo root"
  FRILL_API_KEY="$(grep -E '^FRILL_API_KEY=' "$ROOT/.env" | tail -1 | cut -d= -f2- | tr -d '"'"'"' \r\n')"
  [[ -n "$FRILL_API_KEY" ]] || die "FRILL_API_KEY missing or empty in .env"
}

cmd_draft() {
  local name="" content_file="" categories="" ideas="" author="" dry=0
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --name)         name="${2:?}"; shift 2 ;;
      --content-file) content_file="${2:?}"; shift 2 ;;
      --categories)   categories="${2:?}"; shift 2 ;;
      --ideas)        ideas="${2:?}"; shift 2 ;;
      --author)       author="${2:?}"; shift 2 ;;
      --dry-run)      dry=1; shift ;;
      *) die "unknown argument: $1" ;;
    esac
  done
  [[ -n "$name" ]] || die "--name is required"
  [[ -n "$content_file" && -s "$content_file" ]] || die "--content-file must point at a non-empty file"
  [[ -n "$author" ]] || die "--author <user_idx> is required by Frill (its docs wrongly imply otherwise)"

  # Guard the house rules at the point of no return.
  grep -q '—' "$content_file" && die "content contains an em dash; house style forbids them (see docs/frill/changelog-template.md)"

  local payload
  payload="$(CONTENT_FILE="$content_file" NAME="$name" CATS="$categories" IDEAS="$ideas" AUTHOR="$author" python3 - <<'PY'
import json, os
body = open(os.environ['CONTENT_FILE']).read().strip()
p = {"name": os.environ['NAME'], "content": body, "published_at": None, "author_idx": os.environ['AUTHOR']}
cats = [c.strip() for c in os.environ.get('CATS','').split(',') if c.strip()]
if cats: p["category_idxs"] = cats
ideas = [i.strip() for i in os.environ.get('IDEAS','').split(',') if i.strip()]
if ideas: p["idea_idxs"] = ideas
print(json.dumps(p))
PY
)"

  if (( dry )); then
    note "DRY RUN, nothing sent. Payload:"
    printf '%s\n' "$payload" | python3 -c 'import json,sys; d=json.load(sys.stdin); d["content"]=d["content"][:200]+"... [truncated]"; print(json.dumps(d, indent=2))'
    return 0
  fi

  load_key
  local tmp code; tmp="$(mktemp)"
  code="$(curl -sS -o "$tmp" -w '%{http_code}' -X POST \
    -H "Authorization: Bearer $FRILL_API_KEY" \
    -H 'Content-Type: application/json' \
    -H 'Accept: application/json' \
    --max-time 60 --data "$payload" "$API/announcements" || echo 000)"
  case "$code" in
    200|201) jq . "$tmp" 2>/dev/null || cat "$tmp"; rm -f "$tmp" ;;
    401|403) rm -f "$tmp"; die "Frill returned $code — API key invalid or lacks write scope (key not printed)" ;;
    *) note "Frill returned $code"; head -c 600 "$tmp" >&2; echo >&2; rm -f "$tmp"; die "announcement NOT created" ;;
  esac
}

# Change ONE idea's status. Nothing else about the idea is touched.
cmd_status() {
  local idea="${1:?idea idx required}" status="${2:?status idx required}"; shift 2 || true
  local dry=0; [[ "${1:-}" == "--dry-run" ]] && dry=1
  [[ "$idea" == idea_* ]] || die "first argument must be an idea idx like idea_xxxxxxxx (got: $idea)"
  [[ "$status" == status_* ]] || die "second argument must be a status idx like status_xxxxxxxx (got: $status)"

  local payload; payload="$(printf '{"status_idx":"%s"}' "$status")"
  if (( dry )); then note "DRY RUN, nothing sent. POST /ideas/$idea  $payload"; return 0; fi

  load_key
  local tmp code; tmp="$(mktemp)"
  code="$(curl -sS -o "$tmp" -w '%{http_code}' -X POST -H "Authorization: Bearer $FRILL_API_KEY" -H 'Content-Type: application/json' -H 'Accept: application/json' --max-time 60 --data "$payload" "$API/ideas/$idea" || echo 000)"
  case "$code" in
    200|201) jq . "$tmp" 2>/dev/null || cat "$tmp"; rm -f "$tmp" ;;
    401|403) rm -f "$tmp"; die "Frill returned $code - API key invalid or lacks write scope (key not printed)" ;;
    *) note "Frill returned $code"; head -c 600 "$tmp" >&2; echo >&2; rm -f "$tmp"; die "status NOT changed" ;;
  esac
}

case "${1:-}" in
  draft) shift; cmd_draft "$@" ;;
  status) shift; cmd_status "$@" ;;
  *) sed -n '2,22p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'; exit 1 ;;
esac
