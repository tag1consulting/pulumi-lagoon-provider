#!/usr/bin/env bash
#
# bedrock-call.sh — Call the Tag1 Bedrock proxy and extract the response text.
#
# Usage:
#   ./bedrock-call.sh <model_id> <system_prompt_file> <user_message_file> [max_tokens]
#
# Environment:
#   BEDROCK_API_URL  — Bedrock proxy base URL (e.g., https://openwebui-proxy.tag1.io/bedrock)
#   BEDROCK_API_KEY  — Bearer token for authentication
#
# Output:
#   Writes the assistant's response text to stdout.
#   Exits non-zero on API errors with diagnostics on stderr.

set -euo pipefail

MODEL_ID="${1:?Usage: bedrock-call.sh <model_id> <system_prompt_file> <user_message_file> [max_tokens]}"
SYSTEM_PROMPT_FILE="${2:?Missing system prompt file}"
USER_MESSAGE_FILE="${3:?Missing user message file}"
MAX_TOKENS="${4:-4096}"

: "${BEDROCK_API_URL:?BEDROCK_API_URL is required}"
: "${BEDROCK_API_KEY:?BEDROCK_API_KEY is required}"

# Build the request JSON using jq to safely handle prompt content
SYSTEM_PROMPT=$(cat "$SYSTEM_PROMPT_FILE")
USER_MESSAGE=$(cat "$USER_MESSAGE_FILE")

REQUEST_BODY=$(jq -n \
  --arg system "$SYSTEM_PROMPT" \
  --arg user "$USER_MESSAGE" \
  --argjson max_tokens "$MAX_TOKENS" \
  '{
    anthropic_version: "bedrock-2023-05-31",
    system: $system,
    messages: [{ role: "user", content: $user }],
    max_tokens: $max_tokens,
    temperature: 0.3
  }')

# URL-encode the model ID (replace special chars used in Bedrock model IDs)
ENCODED_MODEL_ID=$(printf '%s' "$MODEL_ID" | jq -sRr @uri)
URL="${BEDROCK_API_URL}/model/${ENCODED_MODEL_ID}/invoke"

RESPONSE_FILE=$(mktemp /tmp/bedrock-response-XXXXXXXX.json)
trap 'rm -f "$RESPONSE_FILE"' EXIT

# Use -s (silent, no progress bar) but NOT -f (fail) so we always capture
# the response body for diagnostics on error.
HTTP_CODE=$(curl -s -w "%{http_code}" -o "$RESPONSE_FILE" \
  --max-time 180 \
  -X POST "$URL" \
  -H "Authorization: Bearer ${BEDROCK_API_KEY}" \
  -H "Content-Type: application/json" \
  -d "$REQUEST_BODY") || {
    HTTP_CODE="${HTTP_CODE:-000}"
    echo "ERROR: Bedrock API request failed (HTTP ${HTTP_CODE})" >&2
    if [[ -f "$RESPONSE_FILE" ]]; then
      echo "Response body:" >&2
      cat "$RESPONSE_FILE" >&2
    fi
    exit 1
  }

if [[ "$HTTP_CODE" -lt 200 || "$HTTP_CODE" -ge 300 ]]; then
  echo "ERROR: Bedrock API returned HTTP ${HTTP_CODE}" >&2
  cat "$RESPONSE_FILE" >&2
  exit 1
fi

# Extract the response text from the Anthropic Messages format
RESPONSE_TEXT=$(jq -r '.content[0].text // empty' "$RESPONSE_FILE" 2>/dev/null)

if [[ -z "$RESPONSE_TEXT" ]]; then
  echo "ERROR: Could not extract response text from API response" >&2
  echo "Raw response:" >&2
  cat "$RESPONSE_FILE" >&2
  exit 1
fi

printf '%s\n' "$RESPONSE_TEXT"
