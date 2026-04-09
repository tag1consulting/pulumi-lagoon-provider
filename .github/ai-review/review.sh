#!/usr/bin/env bash
#
# review.sh — AI PR Review orchestrator.
#
# Computes the diff, builds a file manifest, detects languages, calls
# pr-summarizer and code-reviewer agents via the Bedrock proxy, assembles
# the results, and posts them to the PR.
#
# Environment (required):
#   BEDROCK_API_URL   — Bedrock proxy base URL
#   BEDROCK_API_KEY   — Bearer token for proxy auth
#   GH_TOKEN          — GitHub token for posting reviews
#   PR_NUMBER         — Pull request number
#   BASE_REF          — Base branch name (e.g., main)
#   HEAD_SHA          — Head commit SHA
#   GITHUB_REPOSITORY — owner/repo
#
# Environment (optional):
#   AI_REVIEW_MODE    — "auto" (default), "quick", or "full"

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REVIEW_MODE="${AI_REVIEW_MODE:-auto}"

# Model IDs
MODEL_SONNET="us.anthropic.claude-sonnet-4-20250514-v1:0"

# Temp files — cleaned up on exit
TMPFILES=()
cleanup() {
  rm -f "${TMPFILES[@]}" 2>/dev/null || true
}
trap cleanup EXIT

mktemp_tracked() {
  local f
  f=$(mktemp "$@")
  TMPFILES+=("$f")
  echo "$f"
}

# ---------------------------------------------------------------------------
# Phase 0: Pre-flight — compute diff, build manifest
# ---------------------------------------------------------------------------
echo "=== AI PR Review ===" >&2
echo "PR: #${PR_NUMBER} | Base: ${BASE_REF} | Head: ${HEAD_SHA}" >&2
echo "Mode: ${REVIEW_MODE}" >&2

# Ensure we have the base branch for diffing
git fetch origin "${BASE_REF}" --depth=50 2>/dev/null || true

# Compute the diff
DIFF_FILE=$(mktemp_tracked /tmp/ai-review-diff-XXXXXXXX.txt)
git diff "origin/${BASE_REF}...${HEAD_SHA}" -- \
  ':!*lock.json' ':!*lock.yaml' ':!vendor/*' ':!*.sum' ':!node_modules/*' \
  > "$DIFF_FILE" 2>/dev/null || true

# Check for empty diff
DIFF_LINES=$(wc -l < "$DIFF_FILE" | tr -d ' ')
if [[ "$DIFF_LINES" -eq 0 ]]; then
  echo "No changes detected. Skipping review." >&2
  exit 0
fi

echo "Diff: ${DIFF_LINES} lines" >&2

# Build file manifest
CHANGED_FILES=$(git diff --name-only "origin/${BASE_REF}...${HEAD_SHA}" -- \
  ':!*lock.json' ':!*lock.yaml' ':!vendor/*' ':!*.sum' ':!node_modules/*' 2>/dev/null || true)
if [[ -z "$CHANGED_FILES" ]]; then
  echo "No changed files after exclusions. Skipping review." >&2
  exit 0
fi
DIFF_STAT=$(git diff --stat "origin/${BASE_REF}...${HEAD_SHA}" -- \
  ':!*lock.json' ':!*lock.yaml' ':!vendor/*' ':!*.sum' ':!node_modules/*' 2>/dev/null | tail -1)
FILE_COUNT=$(echo "$CHANGED_FILES" | wc -l | tr -d ' ')

# Detect languages from extensions
LANGUAGES=""
detect_language() {
  local ext="$1"
  case "$ext" in
    go) echo "Go" ;;
    py) echo "Python" ;;
    js|jsx) echo "JavaScript" ;;
    ts|tsx) echo "TypeScript" ;;
    php|module|theme|inc) echo "PHP" ;;
    tf|tfvars) echo "Terraform" ;;
    sh|bash) echo "Shell" ;;
    yaml|yml) echo "YAML" ;;
    *) echo "" ;;
  esac
}

DETECTED_LANGS=()
while IFS= read -r file; do
  ext="${file##*.}"
  lang=$(detect_language "$ext")
  if [[ -n "$lang" ]]; then
    # Add to array if not already present
    found=0
    for existing in "${DETECTED_LANGS[@]+"${DETECTED_LANGS[@]}"}"; do
      if [[ "$existing" == "$lang" ]]; then
        found=1
        break
      fi
    done
    if [[ "$found" -eq 0 ]]; then
      DETECTED_LANGS+=("$lang")
    fi
  fi
done <<< "$CHANGED_FILES"

LANGUAGES=$(IFS=", "; echo "${DETECTED_LANGS[*]+"${DETECTED_LANGS[*]}"}")
echo "Languages: ${LANGUAGES:-none detected}" >&2
echo "Files: ${FILE_COUNT} | ${DIFF_STAT}" >&2

# Categorize files
SOURCE_FILES=""
TEST_FILES=""
CONFIG_FILES=""
DOC_FILES=""
while IFS= read -r file; do
  if [[ "$file" =~ _test\.go$ ]] || [[ "$file" =~ test_.*\.py$ ]] || \
     [[ "$file" =~ \.test\.[jt]sx?$ ]] || [[ "$file" =~ \.spec\.[jt]sx?$ ]] || \
     [[ "$file" =~ Test\.php$ ]] || [[ "$file" =~ /tests/ ]]; then
    TEST_FILES="${TEST_FILES}${file}\n"
  elif [[ "$file" =~ \.(md|txt|rst)$ ]]; then
    DOC_FILES="${DOC_FILES}${file}\n"
  elif [[ "$file" =~ \.(yml|yaml|json|toml|cfg|ini|env)$ ]] || \
       [[ "$file" =~ Makefile$ ]] || [[ "$file" =~ Dockerfile$ ]] || \
       [[ "$file" =~ \.github/ ]]; then
    CONFIG_FILES="${CONFIG_FILES}${file}\n"
  else
    SOURCE_FILES="${SOURCE_FILES}${file}\n"
  fi
done <<< "$CHANGED_FILES"

# Build manifest text
MANIFEST="BASE: ${BASE_REF} | LANGUAGES: ${LANGUAGES:-unknown} | FILES: ${FILE_COUNT} | ${DIFF_STAT}"
if [[ -n "$SOURCE_FILES" ]]; then
  MANIFEST="${MANIFEST}\n\nSource: $(echo -e "$SOURCE_FILES" | head -20 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$TEST_FILES" ]]; then
  MANIFEST="${MANIFEST}\nTests: $(echo -e "$TEST_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$CONFIG_FILES" ]]; then
  MANIFEST="${MANIFEST}\nConfig: $(echo -e "$CONFIG_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$DOC_FILES" ]]; then
  MANIFEST="${MANIFEST}\nDocs: $(echo -e "$DOC_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi

# Commit log
COMMIT_LOG=$(git log --oneline "origin/${BASE_REF}..${HEAD_SHA}" 2>/dev/null | head -20)

# Determine review mode
TOTAL_CHANGED=$(echo "$DIFF_STAT" | grep -oE '[0-9]+ insertions?' | grep -o '[0-9]*' || echo "0")
TOTAL_REMOVED=$(echo "$DIFF_STAT" | grep -oE '[0-9]+ deletions?' | grep -o '[0-9]*' || echo "0")
TOTAL_LINES=$(( ${TOTAL_CHANGED:-0} + ${TOTAL_REMOVED:-0} ))

if [[ "$REVIEW_MODE" == "auto" ]]; then
  if [[ "$TOTAL_LINES" -gt 2000 ]]; then
    echo "WARNING: Large diff (${TOTAL_LINES} lines). Forcing quick mode." >&2
    REVIEW_MODE="quick"
  elif [[ "$TOTAL_LINES" -lt 100 ]]; then
    REVIEW_MODE="quick"
    echo "Small diff (${TOTAL_LINES} lines). Using quick mode." >&2
  else
    REVIEW_MODE="full"
    echo "Medium diff (${TOTAL_LINES} lines). Using full mode." >&2
  fi
fi

# Load language profile(s)
LANGUAGE_CONTEXT=""
for lang in "${DETECTED_LANGS[@]+"${DETECTED_LANGS[@]}"}"; do
  lang_lower=$(echo "$lang" | tr '[:upper:]' '[:lower:]')
  profile="${SCRIPT_DIR}/language-profiles/${lang_lower}.md"
  if [[ -f "$profile" ]]; then
    LANGUAGE_CONTEXT="${LANGUAGE_CONTEXT}\n$(cat "$profile")\n"
  fi
done

# Read project context (CLAUDE.md) if available
PROJECT_CONTEXT=""
if [[ -f "CLAUDE.md" ]]; then
  # Extract first ~500 tokens (~2000 chars) of project context
  PROJECT_CONTEXT=$(head -c 2000 CLAUDE.md)
  if [[ $(wc -c < CLAUDE.md) -gt 2000 ]]; then
    echo "NOTE: CLAUDE.md truncated to 2000 chars for agent context." >&2
  fi
fi

# ---------------------------------------------------------------------------
# Phase 1: Call agents
# ---------------------------------------------------------------------------
echo "--- Calling agents ---" >&2

# Prepare user messages for each agent
SUMMARIZER_MSG_FILE=$(mktemp_tracked /tmp/ai-review-summarizer-msg-XXXXXXXX.md)
REVIEWER_MSG_FILE=$(mktemp_tracked /tmp/ai-review-reviewer-msg-XXXXXXXX.md)

# pr-summarizer gets: manifest + commit log + diff
{
  echo "## File Manifest"
  echo -e "$MANIFEST"
  echo ""
  echo "## Commit Log"
  echo "$COMMIT_LOG"
  echo ""
  if [[ -n "$PROJECT_CONTEXT" ]]; then
    echo "## Project Context"
    echo "$PROJECT_CONTEXT"
    echo ""
  fi
  echo "## Diff"
  cat "$DIFF_FILE"
} > "$SUMMARIZER_MSG_FILE"

# code-reviewer gets: manifest + language context + diff
{
  echo "## File Manifest"
  echo -e "$MANIFEST"
  echo ""
  if [[ -n "$LANGUAGE_CONTEXT" ]]; then
    echo -e "$LANGUAGE_CONTEXT"
    echo ""
  fi
  echo "## Diff"
  cat "$DIFF_FILE"
} > "$REVIEWER_MSG_FILE"

SUMMARY_FILE=$(mktemp_tracked /tmp/ai-review-summary-XXXXXXXX.md)
FINDINGS_FILE=$(mktemp_tracked /tmp/ai-review-findings-XXXXXXXX.md)

# Call both agents (sequentially for PoC; parallel in future phases)
echo "Calling pr-summarizer (sonnet)..." >&2
"${SCRIPT_DIR}/bedrock-call.sh" "$MODEL_SONNET" \
  "${SCRIPT_DIR}/prompts/pr-summarizer.md" \
  "$SUMMARIZER_MSG_FILE" \
  4096 > "$SUMMARY_FILE" || {
    echo "WARNING: pr-summarizer failed. Continuing without summary." >&2
    echo "NONE" > "$SUMMARY_FILE"
  }

echo "Calling code-reviewer (sonnet)..." >&2
"${SCRIPT_DIR}/bedrock-call.sh" "$MODEL_SONNET" \
  "${SCRIPT_DIR}/prompts/code-reviewer.md" \
  "$REVIEWER_MSG_FILE" \
  4096 > "$FINDINGS_FILE" || {
    echo "WARNING: code-reviewer failed. Continuing without findings." >&2
    echo "NONE" > "$FINDINGS_FILE"
  }

echo "Agents complete." >&2

# ---------------------------------------------------------------------------
# Phase 2: Parse findings JSON
# ---------------------------------------------------------------------------
FINDINGS_JSON_FILE=$(mktemp_tracked /tmp/ai-review-findings-json-XXXXXXXX.json)

# Extract json-findings block from code-reviewer output
if grep -q '```json-findings' "$FINDINGS_FILE"; then
  sed -n '/```json-findings/,/```/p' "$FINDINGS_FILE" | \
    sed '1d;$d' > "$FINDINGS_JSON_FILE"
  # Validate JSON
  if ! jq -e 'type == "array"' "$FINDINGS_JSON_FILE" > /dev/null 2>&1; then
    echo "WARNING: Could not parse findings JSON. Inline comments will not be posted." >&2
    echo "[]" > "$FINDINGS_JSON_FILE"
  fi
else
  echo "[]" > "$FINDINGS_JSON_FILE"
fi

# Strip the json-findings block from the findings markdown
FINDINGS_CLEAN_FILE=$(mktemp_tracked /tmp/ai-review-findings-clean-XXXXXXXX.md)
sed '/```json-findings/,/```/d' "$FINDINGS_FILE" > "$FINDINGS_CLEAN_FILE"

# ---------------------------------------------------------------------------
# Phase 3: Post to GitHub
# ---------------------------------------------------------------------------
echo "--- Posting to GitHub ---" >&2

"${SCRIPT_DIR}/post-review.sh" \
  "$PR_NUMBER" \
  "$SUMMARY_FILE" \
  "$FINDINGS_CLEAN_FILE" \
  "$FINDINGS_JSON_FILE" \
  "$DIFF_FILE" \
  "$HEAD_SHA"

# ---------------------------------------------------------------------------
# Phase 4: Summary to step summary
# ---------------------------------------------------------------------------
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "## AI PR Review Results"
    echo ""
    echo "**Mode:** ${REVIEW_MODE}"
    echo "**Files:** ${FILE_COUNT}"
    echo "**Languages:** ${LANGUAGES:-none detected}"
    echo ""
    FINDING_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE" 2>/dev/null || echo "0")
    echo "**Findings:** ${FINDING_COUNT}"
    echo ""
    echo "### Summary"
    cat "$SUMMARY_FILE"
  } >> "$GITHUB_STEP_SUMMARY"
fi

echo "=== AI PR Review complete ===" >&2
