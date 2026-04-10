#!/usr/bin/env bash
#
# review.sh — AI PR Review orchestrator.
#
# Computes the diff, builds a file manifest, detects languages, calls
# review agents via the configured LLM provider, assembles the results,
# and posts them to the PR.
#
# Environment (required):
#   AI_PROVIDER       — LLM provider: anthropic | openai | openai-compatible | google | bedrock-proxy
#   GH_TOKEN          — GitHub token for posting reviews
#   PR_NUMBER         — Pull request number
#   BASE_REF          — Base branch name (e.g., main)
#   HEAD_SHA          — Head commit SHA
#   GITHUB_REPOSITORY — owner/repo
#
#   Provider credentials (one set required based on AI_PROVIDER):
#     anthropic:          ANTHROPIC_API_KEY
#     openai:             OPENAI_API_KEY
#     openai-compatible:  OPENAI_API_KEY, OPENAI_BASE_URL
#     google:             GOOGLE_API_KEY
#     bedrock-proxy:      BEDROCK_API_URL, BEDROCK_API_KEY
#
# Environment (optional):
#   AI_REVIEW_MODE    — "quick" (default) or "full"
#                       Add the "ai-review-full" label to a PR for full mode.
#   AI_MODEL_STANDARD — Model for standard agents (pr-summarizer, code-reviewer, etc.)
#                       Defaults are chosen per provider if not set.
#   AI_MODEL_PREMIUM  — Model for deep agents (architecture-reviewer, security-reviewer)
#                       Defaults to AI_MODEL_STANDARD if not set.
#   AI_TEMPERATURE    — Sampling temperature (default: 0.3)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REVIEW_MODE="${AI_REVIEW_MODE:-quick}"

: "${AI_PROVIDER:?AI_PROVIDER is required (anthropic|openai|openai-compatible|google|bedrock-proxy)}"

# Set per-provider model defaults; user env vars take precedence.
case "$AI_PROVIDER" in
  anthropic)
    AI_MODEL_STANDARD="${AI_MODEL_STANDARD:-claude-sonnet-4-6-20250514}"
    AI_MODEL_PREMIUM="${AI_MODEL_PREMIUM:-claude-opus-4-6-20250514}"
    ;;
  openai|openai-compatible)
    AI_MODEL_STANDARD="${AI_MODEL_STANDARD:-gpt-4o}"
    AI_MODEL_PREMIUM="${AI_MODEL_PREMIUM:-${AI_MODEL_STANDARD}}"
    ;;
  google)
    AI_MODEL_STANDARD="${AI_MODEL_STANDARD:-gemini-2.5-flash}"
    AI_MODEL_PREMIUM="${AI_MODEL_PREMIUM:-gemini-2.5-pro}"
    ;;
  bedrock-proxy)
    AI_MODEL_STANDARD="${AI_MODEL_STANDARD:-us.anthropic.claude-sonnet-4-6}"
    AI_MODEL_PREMIUM="${AI_MODEL_PREMIUM:-global.anthropic.claude-opus-4-6-v1}"
    ;;
  *)
    # openai-compatible without standard: user must set AI_MODEL_STANDARD
    AI_MODEL_STANDARD="${AI_MODEL_STANDARD:?AI_MODEL_STANDARD is required for AI_PROVIDER=${AI_PROVIDER}}"
    AI_MODEL_PREMIUM="${AI_MODEL_PREMIUM:-${AI_MODEL_STANDARD}}"
    ;;
esac

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
git fetch origin "${BASE_REF}" --depth=50 2>/dev/null || {
  echo "WARNING: git fetch failed; attempting to proceed with existing local refs." >&2
  if ! git rev-parse --verify "origin/${BASE_REF}" > /dev/null 2>&1; then
    echo "ERROR: origin/${BASE_REF} is not reachable. Cannot compute diff. Aborting." >&2
    exit 1
  fi
}

# ---------------------------------------------------------------------------
# Incremental diff: only review commits since the last review run.
# Fall back to the full PR diff on first run or if the last SHA is unreachable.
# ---------------------------------------------------------------------------
LAST_REVIEWED_SHA=$("${SCRIPT_DIR}/post-review.sh" --get-last-sha 2>/dev/null) || {
  echo "WARNING: Could not retrieve last-reviewed SHA; falling back to full PR diff." >&2
  LAST_REVIEWED_SHA=""
}
DIFF_BASE=""
DIFF_LABEL=""

if [[ -n "$LAST_REVIEWED_SHA" && "$LAST_REVIEWED_SHA" != "$HEAD_SHA" ]]; then
  # Verify the SHA is reachable AND is an ancestor of HEAD (guards against force-push/rebase)
  if git cat-file -e "${LAST_REVIEWED_SHA}^{commit}" 2>/dev/null && \
     git merge-base --is-ancestor "$LAST_REVIEWED_SHA" "$HEAD_SHA" 2>/dev/null; then
    DIFF_BASE="$LAST_REVIEWED_SHA"
    DIFF_LABEL="incremental (${LAST_REVIEWED_SHA:0:7}..${HEAD_SHA:0:7})"
    echo "Incremental review: diffing ${LAST_REVIEWED_SHA:0:7}..${HEAD_SHA:0:7}" >&2
  else
    echo "Last-reviewed SHA ${LAST_REVIEWED_SHA:0:7} not reachable or not an ancestor; falling back to full PR diff." >&2
  fi
fi

if [[ -z "$DIFF_BASE" ]]; then
  DIFF_LABEL="full PR diff"
  echo "Full PR review: diffing origin/${BASE_REF}...${HEAD_SHA}" >&2
fi

# Compute the diff
DIFF_FILE=$(mktemp_tracked /tmp/ai-review-diff-XXXXXXXX.txt)
EXCL=(':!*lock.json' ':!*lock.yaml' ':!vendor/*' ':!*.sum' ':!node_modules/*')
if [[ -n "$DIFF_BASE" ]]; then
  if ! git diff "${DIFF_BASE}...${HEAD_SHA}" -- "${EXCL[@]}" > "$DIFF_FILE" 2>/dev/null; then
    : > "$DIFF_FILE"
    echo "WARNING: git diff failed; diff output may be empty or incomplete." >&2
  fi
else
  if ! git diff "origin/${BASE_REF}...${HEAD_SHA}" -- "${EXCL[@]}" > "$DIFF_FILE" 2>/dev/null; then
    : > "$DIFF_FILE"
    echo "WARNING: git diff failed; diff output may be empty or incomplete." >&2
  fi
fi

# Check for empty diff
DIFF_LINES=$(wc -l < "$DIFF_FILE" | tr -d ' ')
if [[ "$DIFF_LINES" -eq 0 ]]; then
  echo "No new changes since last review. Skipping." >&2
  exit 0
fi

echo "Diff: ${DIFF_LINES} lines (${DIFF_LABEL})" >&2

# Build file manifest (same range as diff)
if [[ -n "$DIFF_BASE" ]]; then
  CHANGED_FILES=$(git diff --name-only -z "${DIFF_BASE}...${HEAD_SHA}" -- "${EXCL[@]}" 2>/dev/null | tr '\0' '\n' || true)
  DIFF_STAT=$(git diff --stat "${DIFF_BASE}...${HEAD_SHA}" -- "${EXCL[@]}" 2>/dev/null | tail -1)
else
  CHANGED_FILES=$(git diff --name-only -z "origin/${BASE_REF}...${HEAD_SHA}" -- "${EXCL[@]}" 2>/dev/null | tr '\0' '\n' || true)
  DIFF_STAT=$(git diff --stat "origin/${BASE_REF}...${HEAD_SHA}" -- "${EXCL[@]}" 2>/dev/null | tail -1)
fi

if [[ -z "$CHANGED_FILES" ]]; then
  echo "No changed files after exclusions. Skipping review." >&2
  exit 0
fi
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
MANIFEST="BASE: ${BASE_REF} | DIFF: ${DIFF_LABEL} | LANGUAGES: ${LANGUAGES:-unknown} | FILES: ${FILE_COUNT} | ${DIFF_STAT}"
if [[ -n "$SOURCE_FILES" ]]; then
  MANIFEST="${MANIFEST}\n\nSource: $(printf '%b' "$SOURCE_FILES" | head -20 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$TEST_FILES" ]]; then
  MANIFEST="${MANIFEST}\nTests: $(printf '%b' "$TEST_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$CONFIG_FILES" ]]; then
  MANIFEST="${MANIFEST}\nConfig: $(printf '%b' "$CONFIG_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi
if [[ -n "$DOC_FILES" ]]; then
  MANIFEST="${MANIFEST}\nDocs: $(printf '%b' "$DOC_FILES" | head -10 | tr '\n' ', ' | sed 's/,$//')"
fi

# Commit log — scoped to the same range as the diff
if [[ -n "$DIFF_BASE" ]]; then
  COMMIT_LOG=$(git log --oneline "${DIFF_BASE}..${HEAD_SHA}" 2>/dev/null | head -20)
else
  COMMIT_LOG=$(git log --oneline "origin/${BASE_REF}..${HEAD_SHA}" 2>/dev/null | head -20)
fi

# Validate and log review mode
if [[ "$REVIEW_MODE" != "quick" && "$REVIEW_MODE" != "full" ]]; then
  echo "WARNING: Unknown AI_REVIEW_MODE '${REVIEW_MODE}'. Defaulting to quick." >&2
  REVIEW_MODE="quick"
fi

# Compute diff size for informational logging (and large-diff warning)
# Note: || echo "0" here is intentional — parse failures default to 0 lines,
# which is the safe direction (does not suppress any review output).
TOTAL_CHANGED=$(echo "$DIFF_STAT" | grep -oE '[0-9]+ insertions?' | grep -o '[0-9]*' || echo "0")
TOTAL_REMOVED=$(echo "$DIFF_STAT" | grep -oE '[0-9]+ deletions?' | grep -o '[0-9]*' || echo "0")
TOTAL_LINES=$(( ${TOTAL_CHANGED:-0} + ${TOTAL_REMOVED:-0} ))

if [[ "$TOTAL_LINES" -gt 2000 ]]; then
  echo "WARNING: Large diff (${TOTAL_LINES} changed lines). Consider reviewing incrementally." >&2
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

# Track agents that fail and their token usage
FAILED_AGENTS=()
TOKEN_LOG=()  # entries: "agent_name input=N output=N"

# --- Helper: call agent and handle failure ---
# Intercepts TOKENS: lines from llm-call.sh stderr for usage tracking;
# forwards all other stderr to the workflow log.
call_agent() {
  local name="$1" model="$2" prompt="$3" msg="$4" output="$5" max_tokens="${6:-4096}"
  echo "Calling ${name} (${model##*.})..." >&2

  local agent_stderr
  agent_stderr=$(mktemp_tracked /tmp/ai-review-stderr-XXXXXXXX.txt)

  local exit_code=0
  "${SCRIPT_DIR}/llm-call.sh" "$model" "$prompt" "$msg" "$max_tokens" \
    > "$output" 2> "$agent_stderr" || exit_code=$?
  if [[ "$exit_code" -ne 0 ]]; then
    local last_err
    last_err=$(tail -1 "$agent_stderr" 2>/dev/null)
    echo "WARNING: ${name} failed (exit ${exit_code}): ${last_err:-no stderr output}. Continuing without its output." >&2
    cat "$agent_stderr" >&2
    FAILED_AGENTS+=("$name")
    echo "" > "$output"
    return
  fi

  # Parse token usage line; forward remaining stderr to workflow log
  local token_line=""
  while IFS= read -r line; do
    if [[ "$line" == TOKENS:* ]]; then
      token_line="$line"
    else
      echo "$line" >&2
    fi
  done < "$agent_stderr"

  if [[ -n "$token_line" ]]; then
    local input_tokens output_tokens
    input_tokens=$(echo "$token_line" | grep -oE 'input=[0-9]+' | sed 's/input=//' || echo "0")
    output_tokens=$(echo "$token_line" | grep -oE 'output=[0-9]+' | sed 's/output=//' || echo "0")
    echo "  tokens: input=${input_tokens} output=${output_tokens}" >&2
    TOKEN_LOG+=("${name}: input=${input_tokens} output=${output_tokens}")
  fi
}

# --- Detect conditional agent triggers ---
HAS_ERROR_PATTERNS=0
if grep -qE '(catch|if err|try \{|rescue|Result<|unwrap|except|\.catch\()' "$DIFF_FILE" 2>/dev/null; then
  HAS_ERROR_PATTERNS=1
fi

# --- Output files ---
SUMMARY_FILE=$(mktemp_tracked /tmp/ai-review-summary-XXXXXXXX.md)
FINDINGS_FILE=$(mktemp_tracked /tmp/ai-review-findings-XXXXXXXX.md)

# --- Agent roster ---
# Tier 1: Always run (quick + full)
AGENT_OUTPUTS=()

# pr-summarizer: only on the first run. Follow-up runs leave the existing
# summary comment untouched — it describes the full PR; re-running it on an
# incremental diff would replace a useful whole-PR overview with a partial one.
if [[ -z "$LAST_REVIEWED_SHA" ]]; then
  call_agent "pr-summarizer" "$AI_MODEL_STANDARD" \
    "${SCRIPT_DIR}/prompts/pr-summarizer.md" "$FULL_CONTEXT_MSG" "$SUMMARY_FILE"
else
  echo "Skipping pr-summarizer (incremental run — summary already posted)." >&2
fi

call_agent "code-reviewer" "$AI_MODEL_STANDARD" \
  "${SCRIPT_DIR}/prompts/code-reviewer.md" "$CODE_CONTEXT_MSG" "$FINDINGS_FILE"
AGENT_OUTPUTS+=("$FINDINGS_FILE")

# Tier 1 conditional: run in both quick and full when triggered
if [[ "$HAS_ERROR_PATTERNS" -eq 1 ]]; then
  SFH_FILE=$(mktemp_tracked /tmp/ai-review-sfh-XXXXXXXX.md)
  call_agent "silent-failure-hunter" "$AI_MODEL_STANDARD" \
    "${SCRIPT_DIR}/prompts/silent-failure-hunter.md" "$CODE_CONTEXT_MSG" "$SFH_FILE"
  AGENT_OUTPUTS+=("$SFH_FILE")
fi

# Tier 2: Full mode only
if [[ "$REVIEW_MODE" == "full" ]]; then
  ARCH_FILE=$(mktemp_tracked /tmp/ai-review-arch-XXXXXXXX.md)
  call_agent "architecture-reviewer" "$AI_MODEL_PREMIUM" \
    "${SCRIPT_DIR}/prompts/architecture-reviewer.md" "$FULL_CONTEXT_MSG" "$ARCH_FILE"
  AGENT_OUTPUTS+=("$ARCH_FILE")

  SEC_FILE=$(mktemp_tracked /tmp/ai-review-sec-XXXXXXXX.md)
  call_agent "security-reviewer" "$AI_MODEL_PREMIUM" \
    "${SCRIPT_DIR}/prompts/security-reviewer.md" "$CODE_CONTEXT_MSG" "$SEC_FILE"
  AGENT_OUTPUTS+=("$SEC_FILE")

  BLIND_FILE=$(mktemp_tracked /tmp/ai-review-blind-XXXXXXXX.md)
  call_agent "blind-hunter" "$AI_MODEL_STANDARD" \
    "${SCRIPT_DIR}/prompts/blind-hunter.md" "$BLIND_MSG" "$BLIND_FILE"
  AGENT_OUTPUTS+=("$BLIND_FILE")

  EDGE_FILE=$(mktemp_tracked /tmp/ai-review-edge-XXXXXXXX.md)
  call_agent "edge-case-hunter" "$AI_MODEL_STANDARD" \
    "${SCRIPT_DIR}/prompts/edge-case-hunter.md" "$CODE_CONTEXT_MSG" "$EDGE_FILE"
  AGENT_OUTPUTS+=("$EDGE_FILE")

  ADV_FILE=$(mktemp_tracked /tmp/ai-review-adv-XXXXXXXX.md)
  call_agent "adversarial-general" "$AI_MODEL_STANDARD" \
    "${SCRIPT_DIR}/prompts/adversarial-general.md" "$CODE_CONTEXT_MSG" "$ADV_FILE"
  AGENT_OUTPUTS+=("$ADV_FILE")
fi

AGENT_COUNT=${#AGENT_OUTPUTS[@]}
FAILED_COUNT=${#FAILED_AGENTS[@]}
if [[ "$FAILED_COUNT" -gt 0 ]]; then
  echo "Agents complete. (${AGENT_COUNT} finding agents ran, ${FAILED_COUNT} failed: ${FAILED_AGENTS[*]})" >&2
else
  echo "Agents complete. (${AGENT_COUNT} finding agents ran)" >&2
fi

# Log token usage summary
if [[ "${#TOKEN_LOG[@]}" -gt 0 ]]; then
  echo "--- Token usage ---" >&2
  TOTAL_INPUT=0
  TOTAL_OUTPUT=0
  for entry in "${TOKEN_LOG[@]}"; do
    echo "  ${entry}" >&2
    in_tok=$(echo "$entry" | grep -oE 'input=[0-9]+' | sed 's/input=//' || echo "0")
    out_tok=$(echo "$entry" | grep -oE 'output=[0-9]+' | sed 's/output=//' || echo "0")
    TOTAL_INPUT=$(( TOTAL_INPUT + in_tok ))
    TOTAL_OUTPUT=$(( TOTAL_OUTPUT + out_tok ))
  done
  echo "  TOTAL: input=${TOTAL_INPUT} output=${TOTAL_OUTPUT} (combined=$(( TOTAL_INPUT + TOTAL_OUTPUT )))" >&2
fi

# --- Run shellcheck if shell files changed ---
SHELLCHECK_JSON="[]"
if [[ -n "$CHANGED_FILES" ]]; then
  SHELLCHECK_JSON=$("${SCRIPT_DIR}/run-shellcheck.sh" "$CHANGED_FILES") || {
    echo "WARNING: run-shellcheck.sh failed; shellcheck findings will be skipped." >&2
    SHELLCHECK_JSON="[]"
  }
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

# Extract json-findings from each agent output, validate shape, and return clean array.
extract_findings() {
  local agent_file="$1"
  if grep -q '```json-findings' "$agent_file" 2>/dev/null; then
    local extracted
    extracted=$(sed -n '/```json-findings/,/```/p' "$agent_file" | sed '1d;$d')
    # Validate it is an array and each item has required fields
    if echo "$extracted" | jq -e '
      type == "array" and
      (if length > 0 then
        all(.[]; has("severity") and has("finding") and has("confidence"))
      else true end)
    ' > /dev/null 2>&1; then
      printf '%s' "$extracted"
      return
    else
      local preview
      preview=$(printf '%s' "$extracted" | head -c 200)
      echo "WARNING: $(basename "$agent_file") produced invalid or malformed json-findings; skipping. Preview: ${preview}" >&2
    fi
  fi
  echo "[]"
}

merge_findings() {
  local incoming="$1"
  if jq -s '.[0] + .[1]' "$FINDINGS_JSON_FILE" <(echo "$incoming") > "${FINDINGS_JSON_FILE}.tmp" 2>/dev/null; then
    mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
  else
    local count
    count=$(echo "$incoming" | jq 'length' 2>/dev/null || echo "?")
    echo "WARNING: Failed to merge findings JSON; skipping batch of ${count} finding(s)." >&2
    rm -f "${FINDINGS_JSON_FILE}.tmp"
  fi
}

# Apply declarative suppressions from suppressions.json.
# Runs after all findings are merged, before confidence filter and dedup.
# Suppressed findings are removed from FINDINGS_JSON_FILE and logged to stderr.
apply_suppressions() {
  local suppressions_file="${SCRIPT_DIR}/suppressions.json"

  if [[ ! -f "$suppressions_file" ]]; then
    return 0
  fi

  if ! jq -e 'type == "array"' "$suppressions_file" > /dev/null 2>&1; then
    echo "WARNING: suppressions.json is not a valid JSON array; skipping suppression filter." >&2
    return 0
  fi

  local result
  result=$(jq --slurpfile rules "$suppressions_file" '
    ($rules[0]) as $active_rules |

    def rule_matches(rule):
      (if rule.match.code then
        (.finding | startswith(rule.match.code))
      else true end)
      and
      (if rule.match.file then
        (.file // "" | contains(rule.match.file))
      else true end)
      and
      (if rule.match.line then
        (.line == rule.match.line)
      else true end)
      and
      (if rule.match.pattern then
        (.finding | test(rule.match.pattern; "i"))
      else true end);

    def find_rule_id:
      . as $f |
      ([$active_rules[] | select(. as $r | $f | rule_matches($r))][0].id // "?");

    def is_suppressed:
      . as $finding |
      any($active_rules[]; . as $rule | $finding | rule_matches($rule));

    {
      kept: [.[] | select(is_suppressed | not)],
      suppressed: [.[] | select(is_suppressed) | . + {suppression_id: find_rule_id}]
    }
  ' "$FINDINGS_JSON_FILE") || {
    echo "WARNING: apply_suppressions jq failed; skipping suppression filter." >&2
    return 0
  }

  if echo "$result" | jq '.kept' > "${FINDINGS_JSON_FILE}.tmp"; then
    mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
  else
    echo "WARNING: apply_suppressions failed to write filtered findings; keeping original." >&2
    rm -f "${FINDINGS_JSON_FILE}.tmp"
  fi

  local suppressed_count
  suppressed_count=$(echo "$result" | jq '.suppressed | length')
  SUPPRESSED_COUNT=$suppressed_count

  if [[ "$suppressed_count" -gt 0 ]]; then
    echo "Suppressed findings: ${suppressed_count}" >&2
    echo "$result" | jq -r '.suppressed[] | "  SUPPRESSED [\(.suppression_id)] \(.file // "?"):\(.line // "?") — \(.finding | .[0:80])"' >&2
  fi
}

for agent_output in "${AGENT_OUTPUTS[@]}"; do
  AGENT_JSON=$(extract_findings "$agent_output")
  if [[ "$AGENT_JSON" != "[]" ]]; then
    merge_findings "$AGENT_JSON"
  fi
done

# Merge shellcheck findings
if [[ "$SHELLCHECK_JSON" != "[]" ]]; then
  merge_findings "$SHELLCHECK_JSON"
fi

# Apply declarative suppressions (won't-fix / false positives)
SUPPRESSED_COUNT=0
apply_suppressions

# Filter out findings below confidence threshold (75)
PRE_FILTER_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
if jq '[.[] | select((.confidence // 0) >= 75)]' "$FINDINGS_JSON_FILE" > "${FINDINGS_JSON_FILE}.tmp"; then
  mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
else
  echo "WARNING: Confidence filter jq failed; keeping all findings unfiltered." >&2
  rm -f "${FINDINGS_JSON_FILE}.tmp"
fi
POST_FILTER_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
if [[ "$PRE_FILTER_COUNT" -ne "$POST_FILTER_COUNT" ]]; then
  echo "Filtered findings: ${PRE_FILTER_COUNT} → ${POST_FILTER_COUNT} (confidence >= 75)" >&2
fi

# Deduplicate findings on same file:line (keep highest severity)
if jq '
  def sev_rank: if . == "Critical" then 4 elif . == "High" then 3
    elif . == "Medium" then 2 else 1 end;
  group_by(.file + ":" + (.line | tostring))
  | map(sort_by(.severity | sev_rank) | reverse | .[0])
' "$FINDINGS_JSON_FILE" > "${FINDINGS_JSON_FILE}.tmp"; then
  mv "${FINDINGS_JSON_FILE}.tmp" "$FINDINGS_JSON_FILE"
else
  echo "WARNING: Dedup jq failed; findings may contain duplicates." >&2
  rm -f "${FINDINGS_JSON_FILE}.tmp"
fi
DEDUP_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE")
echo "Total findings after dedup: ${DEDUP_COUNT}" >&2

# Build merged findings markdown from all agent outputs (strip json-findings blocks)
FINDINGS_CLEAN_FILE=$(mktemp_tracked /tmp/ai-review-findings-clean-XXXXXXXX.md)
: > "$FINDINGS_CLEAN_FILE"
for agent_output in "${AGENT_OUTPUTS[@]}"; do
  AGENT_CONTENT=$(sed '/```json-findings/,/```/d' "$agent_output")
  if [[ -n "$AGENT_CONTENT" ]]; then
    echo "$AGENT_CONTENT" >> "$FINDINGS_CLEAN_FILE"
    echo "" >> "$FINDINGS_CLEAN_FILE"
  fi
done

# Append failed agent notice if any agents failed
if [[ "${#FAILED_AGENTS[@]}" -gt 0 ]]; then
  echo "> **Note:** The following agents failed and their output is excluded: ${FAILED_AGENTS[*]}" >> "$FINDINGS_CLEAN_FILE"
fi

# Build token usage table as a collapsed details block
TOKEN_TABLE_FILE=$(mktemp_tracked /tmp/ai-review-token-table-XXXXXXXX.md)
if [[ "${#TOKEN_LOG[@]}" -gt 0 ]]; then
  {
    echo "<details>"
    echo "<summary>Token usage by agent</summary>"
    echo ""
    echo "| Agent | Input | Output | Total |"
    echo "|-------|------:|-------:|------:|"
    local_total_in=0
    local_total_out=0
    for entry in "${TOKEN_LOG[@]}"; do
      agent_name="${entry%%:*}"
      in_tok=$(echo "$entry" | grep -oE 'input=[0-9]+' | sed 's/input=//' || echo "0")
      out_tok=$(echo "$entry" | grep -oE 'output=[0-9]+' | sed 's/output=//' || echo "0")
      row_total=$(( in_tok + out_tok ))
      echo "| ${agent_name} | ${in_tok} | ${out_tok} | ${row_total} |"
      local_total_in=$(( local_total_in + in_tok ))
      local_total_out=$(( local_total_out + out_tok ))
    done
    echo "| **Total** | **${local_total_in}** | **${local_total_out}** | **$(( local_total_in + local_total_out ))** |"
    echo ""
    echo "</details>"
  } > "$TOKEN_TABLE_FILE"
fi

# ---------------------------------------------------------------------------
# Phase 3: Post to GitHub
# ---------------------------------------------------------------------------
echo "--- Posting to GitHub ---" >&2

# Pass failed agents to post-review.sh so it can avoid a false APPROVE.
# Use colon as delimiter (agent names never contain colons).
if [[ "${#FAILED_AGENTS[@]}" -gt 0 ]]; then
  export AI_REVIEW_FAILED_AGENTS
  AI_REVIEW_FAILED_AGENTS=$(IFS=:; echo "${FAILED_AGENTS[*]}")
fi

"${SCRIPT_DIR}/post-review.sh" \
  "$PR_NUMBER" \
  "$SUMMARY_FILE" \
  "$FINDINGS_CLEAN_FILE" \
  "$FINDINGS_JSON_FILE" \
  "$DIFF_FILE" \
  "$HEAD_SHA" \
  "$TOKEN_TABLE_FILE"

# ---------------------------------------------------------------------------
# Phase 4: Summary to step summary
# ---------------------------------------------------------------------------
if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "## AI PR Review Results"
    echo ""
    echo "**Mode:** ${REVIEW_MODE} | **Diff:** ${DIFF_LABEL}"
    echo "**Files:** ${FILE_COUNT}"
    echo "**Languages:** ${LANGUAGES:-none detected}"
    echo "**Agents:** ${AGENT_COUNT} finding agents"
    if [[ "${#FAILED_AGENTS[@]}" -gt 0 ]]; then
      echo "**Failed agents:** ${FAILED_AGENTS[*]}"
    fi
    echo ""
    FINDING_COUNT=$(jq 'length' "$FINDINGS_JSON_FILE" 2>/dev/null || echo "0")
    if [[ "${SUPPRESSED_COUNT:-0}" -gt 0 ]]; then
      echo "**Findings:** ${FINDING_COUNT} (${SUPPRESSED_COUNT} suppressed)"
    else
      echo "**Findings:** ${FINDING_COUNT}"
    fi
    echo ""
    if [[ "${#TOKEN_LOG[@]}" -gt 0 ]]; then
      echo "### Token Usage"
      echo ""
      echo "| Agent | Input | Output | Total |"
      echo "|-------|------:|-------:|------:|"
      local_total_in=0
      local_total_out=0
      for entry in "${TOKEN_LOG[@]}"; do
        agent_name="${entry%%:*}"
        in_tok=$(echo "$entry" | grep -oE 'input=[0-9]+' | sed 's/input=//' || echo "0")
        out_tok=$(echo "$entry" | grep -oE 'output=[0-9]+' | sed 's/output=//' || echo "0")
        row_total=$(( in_tok + out_tok ))
        echo "| ${agent_name} | ${in_tok} | ${out_tok} | ${row_total} |"
        local_total_in=$(( local_total_in + in_tok ))
        local_total_out=$(( local_total_out + out_tok ))
      done
      echo "| **Total** | **${local_total_in}** | **${local_total_out}** | **$(( local_total_in + local_total_out ))** |"
      echo ""
    fi
    echo "### Summary"
    cat "$SUMMARY_FILE"
  } >> "$GITHUB_STEP_SUMMARY"
fi

echo "=== AI PR Review complete ===" >&2
