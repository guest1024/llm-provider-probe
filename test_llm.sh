#!/usr/bin/env bash
set -euo pipefail

: "${BASE_URL:=https://vibediary.app/api/v1}"
: "${MODEL:=gpt-5.4}"

if ! command -v jq >/dev/null 2>&1; then
    echo "❌ 错误: 未找到 jq。请先安装 jq。"
    exit 1
fi

if [ -z "${OPENAI_API_KEY:-}" ]; then
    echo "❌ 错误: 请先设置环境变量 OPENAI_API_KEY"
    exit 1
fi

PROMPT=${1:-"用一句话概括一下什么是大语言模型。"}

echo -e "👤 You: $PROMPT\n"
echo -n "🤖 AI ($MODEL): "

STREAM_SUCCESS=false
RESPONSE_BUFFER=""

while IFS= read -r line; do
    if [[ "$line" == data:* ]]; then
        STREAM_SUCCESS=true
        json="${line#data: }"

        if [[ "$json" == "[DONE]" ]]; then
            break
        fi

        content=$(echo "$json" | jq -j -r '.choices[0].delta.content // empty' 2>/dev/null)
        err_msg=$(echo "$json" | jq -r '.error.message // empty' 2>/dev/null)
        if [ -n "$err_msg" ]; then
            echo -e "\n\n⚠️ [流中断异常]: $err_msg"
            break
        fi

        echo -n "$content"
    elif [[ -n "$line" ]]; then
        RESPONSE_BUFFER="${RESPONSE_BUFFER}${line}"
    fi
done < <(curl -s -N "${BASE_URL}/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{
    "model": '"$(printf '%s' "$MODEL" | jq -R -s '.')"',
    "messages": [
      {
        "role": "user",
        "content": '"$(printf '%s' "$PROMPT" | jq -R -s '.')"'
      }
    ],
    "stream": true
  }')

echo -e "\n"

if [ "$STREAM_SUCCESS" = false ] && [ -n "$RESPONSE_BUFFER" ]; then
    NON_STREAM_CONTENT=$(echo "$RESPONSE_BUFFER" | jq -r '.choices[0].message.content // empty' 2>/dev/null)
    if [ -n "$NON_STREAM_CONTENT" ]; then
        echo -e "ℹ️  提示: 当前接口忽略了流式参数，已自适应提取完整文本。"
        echo -e "$NON_STREAM_CONTENT\n"
        exit 0
    fi

    ERROR_MSG=$(echo "$RESPONSE_BUFFER" | jq -r '.error.message // .error // .message // empty' 2>/dev/null)

    echo -e "❌ [流式请求失败]"
    if [ -n "$ERROR_MSG" ]; then
        echo -e "📄 错误详情: $ERROR_MSG\n"
    else
        echo -e "📄 原始网关响应: $RESPONSE_BUFFER\n"
    fi

    echo -e "🔄 正在尝试降级，发起非流式请求 (stream: false)...\n"

    FALLBACK_RESP=$(curl -s "${BASE_URL}/chat/completions" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $OPENAI_API_KEY" \
      -d '{
        "model": '"$(printf '%s' "$MODEL" | jq -R -s '.')"',
        "messages": [
          {
            "role": "user",
            "content": '"$(printf '%s' "$PROMPT" | jq -R -s '.')"'
          }
        ],
        "stream": false
      }')

    FALLBACK_ERROR=$(echo "$FALLBACK_RESP" | jq -r '.error.message // empty' 2>/dev/null)
    FALLBACK_CONTENT=$(echo "$FALLBACK_RESP" | jq -r '.choices[0].message.content // empty' 2>/dev/null)

    if [ -n "$FALLBACK_CONTENT" ]; then
        echo -e "✅ [降级成功] \n$FALLBACK_CONTENT\n"
    elif [ -n "$FALLBACK_ERROR" ]; then
        echo -e "❌ [降级再次失败] 错误详情: $FALLBACK_ERROR\n"
    else
        echo -e "❌ [降级遇到未知异常] 响应: $FALLBACK_RESP\n"
    fi
fi
