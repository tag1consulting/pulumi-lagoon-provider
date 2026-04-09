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

# Model IDs — use latest available on the Bedrock proxy
MODEL_HAIKU="us.anthropic.claude-3-5-haiku-20241022-v1:0"
MODEL_SONNET="us.anthropic.claude-sonnet-4-6"
MODEL_OPUS="global.anthropic.claude-opus-4-6-v1"

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
# Phase 1: Prepare agent messages and call agents
# ---------------------------------------------------------------------------
echo "--- Calling agents ---" >&2

# --- Build shared message files ---

# Full context message: manifest + commit log + project context + language context + diff
FULL_CONTEXT_MSG=$(mktemp_tracked /tmp/ai-review-full-ctx-XXXXXXXX.md)
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
  if [[ -n "$LANGUAGE_CONTEXT" ]]; then
    echo -e "$LANGUAGE_CONTEXT"
    echo ""
  fi
  echo "## Diff"
  cat "$DIFF_FILE"
} > "$FULL_CONTEXT_MSG"

# Code context message: manifest + language context + diff (no commit log/project context)
CODE_CONTEXT_MSG=$(mktemp_tracked /tmp/ai-review-code-ctx-XXXXXXXX.md)
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
} > "$CODE_CONTEXT_MSG"

# Blind message: ONLY the raw diff (zero context — this is intentional)
BLIND_MSG=$(mktemp_tracked /tmp/ai-review-blind-XXXXXXXX.md)
cat "$DIFF_FILE" > "$BLIND_MSG"

# --- Helper: call agent and handle failure ---
call_agent() {
  local name="$1" model="$2" prompt="$3" msg="$4" output="$5" max_tokens="${6:-4096}"
  echo "Calling ${name} (${model##*.claude-})..." >&2
  "${SCRIPT_DIR}/bedrock-call.sh" "$model" "$prompt" "$msg" "$max_tokens" > "$output" || {
    echo "WARNING: ${name} failed. Continuing without its output." >&2
    echo "NONE" > "$output"
  }
}

# --- Detect conditional agent triggers ---
HAS_ERROR_PATTERNS=0
if grep -qE '(catch|if err|try \{|rescue|Result<|unwrap|except|\.catch\(||| true)' "$DIFF_FILE" 2>/dev/null; then
  HAS_ERROR_PATTERNS=1
fi

HAS_TEST_FILES=0
if [[ -n "$TEST_FILES" ]]; then
  HAS_TEST_FILES=1
fi

# --- Output files ---
SUMMARY_FILE=$(mktemp_tracked /tmp/ai-review-summary-XXXXXXXX.md)
FINDINGS_FILE=$(mktemp_tracked /tmp/ai-review-findings-XXXXXXXX.md)

# --- Agent roster ---
# Tier 1: Always run (quick + full)
AGENT_OUTPUTS=()

call_agent "pr-summarizer" "$MODEL_SONNET" \
  "${SCRIPT_DIR}/prompts/pr-summarizer.md" "$FULL_CONTEXT_MSG" "$SUMMARY_FILE"

call_agent "code-reviewer" "$MODEL_SONNET" \
  "${SCRIPT_DIR}/prompts/code-reviewer.md" "$CODE_CONTEXT_MSG" "$FINDINGS_FILE"
AGENT_OUTPUTS+=("$FINDINGS_FILE")

# Tier 1 conditional: run in both quick and full when triggered
if [[ "$HAS_ERROR_PATTERNS" -eq 1 ]]; then
  SFH_FILE=$(mktemp_tracked /tmp/ai-review-sfh-XXXXXXXX.md)
  call_agent "silent-failure-hunter" "$MODEL_SONNET" \
    "${SCRIPT_DIR}/prompts/silent-failure-hunter.md" "$CODE_CONTEXT_MSG" "$SFH_FILE"
  AGENT_OUTPUTS+=("$SFH_FILE")
fi

# Tier 2: Full mode only
if [[ "$REVIEW_MODE" == "full" ]]; then
  ARCH_FILE=$(mktemp_tracked /tmp/ai-review-arch-XXXXXXXX.md)
  call_agent "architecture-reviewer" "$MODEL_OPUS" \
    "${SCRIPT_DIR}/prompts/architecture-reviewer.md" "$FULL_CONTEXT_MSG" "$ARCH_FILE"
  AGENT_OUTPUTS+=("$ARCH_FILE")

  SEC_FILE=$(mktemp_tracked /tmp/ai-review-sec-XXXXXXXX.md)
  call_agent "security-reviewer" "$MODEL_OPUS" \
    "${SCRIPT_DIR}/prompts/security-reviewer.md" "$CODE_CONTEXT_MSG" "$SEC_FILE"
  AGENT_OUTPUTS+=("$SEC_FILE")

  BLIND_FILE=$(mktemp_tracked /tmp/ai-review-blind-XXXXXXXX.md)
  call_agent "blind-hunter" "$MODEL_SONNET" \
    "${SCRIPT_DIR}/prompts/blind-hunter.md" "$BLIND_MSG" "$BLIND_FILE"
  AGENT_OUTPUTS+=("$BLIND_FILE")

  EDGE_FILE=$(mktemp_tracked /tmp/ai-review-edge-XXXXXXXX.md)
  call_agent "edge-case-hunter" "$MODEL_SONNET" \
    "${SCRIPT_DIR}/prompts/edge-case-hunter.md" "$CODE_CONTEXT_MSG" "$EDGE_FILE"
  AGENT_OUTPUTS+=("$EDGE_FILE")

  ADV_FILE=$(mktemp_tracked /tmp/ai-review-adv-XXXXXXXX.md)
  call_agent "adversarial-general" "$MODEL_SONNET" \
    "${SCRIPT_DIR}/prompts/adversarial-general.md" "$CODE_CONTEXT_MSG" "$ADV_FILE"
  AGENT_OUTPUTS+=("$ADV_FILE")
fi

AGENT_COUNT=${#AGENT_OUTPUTS[@]}
echo "Agents complete. (${AGENT_COUNT} finding agents ran)" >&2

# --- Run shellcheck if shell files changed ---
SHELLCHECK_JSON="[]"
if [[ -n "$CHANGED_FILES" ]]; then
  SHELLCHECK_JSON=$("${SCRIPT_DIR}/run-shellcheck.sh" "$CHANGED_FILES" 2>/dev/null || echo "[]")
  SC_COUNT=$(echo "$SHELLCHECK_JSON" | jq 'length' 2>/dev/null || echo "0")
  if [[ "$SC_COUNT" -gt 0 ]]; then
    echo "Shellcheck: ${SC_COUNT} findings" >&2
  fi
fi

# ---------------------------------------------------------------------------
# Phase 2: Parse and merge findings JSON from all agents
# ---------------------------------------------------------------------------
FINDINGS_JSON_FILE=$(mktemp_tracked /tmp/ai-review-findings-json-XXXXXXXX.json)
echo "[]" > "$FINDINGS_JSON_FILE"

# Extract json-findings from each agent output and merge
extract_findings() {
  local agent_file="$1"
  if grep -q '```json-findings' "$agent_file" 2>/dev/null; then
    local extracted
    extracted=$(sed -n '/```json-findings/,/```/p' "$agent_file" | sed '1d;$d')
    if echo "$extracted" | jq -e 'type == "array"' > /dev/null 2>&1; then
      printf '%s' "$extracted"
      return
    fi
  fi
  echo "[]"
}

for agent_output in "${AGENT_OUTPUTS[@]}"; do
  AGENT_JSON=$(extract_findings "$agent_output")
  if [[ "$AGENT_JSON" != "[]" ]]; then
    jq -s '.[0] + .[1]' "$FINDINGS_JSON_FILE" <(echo "$AGENT_JSON") > "${FINDINGS_JSON_FILE}.tmp"
    mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
  fi
done

# Merge shellcheck findings
if [[ "$SHELLCHECK_JSON" != "[]" ]]; then
  jq -s '.[0] + .[1]' "$FINDINGS_JSON_FILE" <(echo "$SHELLCHECK_JSON") > "${FINDINGS_JSON_FILE}.tmp"
  mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
fi

# Filter out findings below confidence threshold (75)
PRE_FILTER_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
jq '[.[] | select((.confidence // 0) >= 75)]' "$FINDINGS_JSON_FILE" > "${FINDINGS_JSON_FILE}.tmp"
mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
POST_FILTER_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
if [[ "$PRE_FILTER_COUNT" -ne "$POST_FILTER_COUNT" ]]; then
  echo "Filtered findings: ${PRE_FILTER_COUNT} → ${POST_FILTER_COUNT} (confidence >= 75)" >&2
fi

# Deduplicate findings on same file:line (keep highest severity)
jq '
  def sev_rank: if . == "Critical" then 4 elif . == "High" then 3
    elif . == "Medium" then 2 else 1 end;
  group_by(.file + ":" + (.line | tostring))
  | map(sort_by(.severity | sev_rank) | reverse | .[0])
' "$FINDINGS_JSON_FILE" > "${FINDINGS_JSON_FILE}.tmp"
mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
DEDUP_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
echo "Total findings after dedup: ${DEDUP_COUNT}" >&2

# Build merged findings markdown from all agent outputs (strip json-findings blocks)
FINDINGS_CLEAN_FILE=$(mktemp_tracked /tmp/ai-review-findings-clean-XXXXXXXX.md)
: > "$FINDINGS_CLEAN_FILE"
for agent_output in "${AGENT_OUTPUTS[@]}"; do
  AGENT_CONTENT=$(sed '/```json-findings/,/```/d' "$agent_output")
  if [[ -n "$AGENT_CONTENT" && "$AGENT_CONTENT" != "NONE" ]]; then
    echo "$AGENT_CONTENT" >> "$FINDINGS_CLEAN_FILE"
    echo "" >> "$FINDINGS_CLEAN_FILE"
  fi
done

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
    echo "**Agents:** ${AGENT_COUNT} finding agents"
    echo ""
    FINDING_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE" 2>/dev/null || echo "0")
    echo "**Findings:** ${FINDING_COUNT}"
    echo ""
    echo "### Summary"
    cat "$SUMMARY_FILE"
  } >> "$GITHUB_STEP_SUMMARY"
fi

echo "=== AI PR Review complete ===" >&2
