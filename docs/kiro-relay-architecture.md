# Kiro 中转架构交接文档

## 1. 背景

本项目（provider-probe）用于评测 LLM API 中转服务的质量。当前主要对接的中转是 **Kiro Gateway**，一个基于 AWS Q Developer / Kiro IDE 凭证的 OpenAI 兼容网关。

---

## 2. 系统架构

### 2.1 整体拓扑

```
┌─────────────────────┐
│  provider-probe     │  Go CLI 评测工具
│  (本项目)            │
└──────────┬──────────┘
           │ HTTP POST /chat/completions
           │ Authorization: Bearer <PROXY_API_KEY>
           ▼
┌─────────────────────┐
│  Kiro Gateway       │  Python FastAPI 服务
│  localhost:58575    │  (kiro-gateway 项目)
└──────────┬──────────┘
           │ AWS Q Developer API
           │ (使用 Kiro IDE 凭证)
           ▼
┌─────────────────────┐
│  实际 LLM 后端       │  Claude / DeepSeek / Qwen 等
└─────────────────────┘
```

### 2.2 组件说明

| 组件 | 位置 | 作用 |
| --- | --- | --- |
| provider-probe | `/root/llm-provider-probe` | 评测工具，发送测试请求并验证响应 |
| kiro-gateway | `/root/kiro-gateway` | OpenAI 兼容网关，转发请求到 AWS Q API |
| Kiro IDE 凭证 | `~/.aws/sso/cache/` 或 `credentials.json` | 认证凭证来源 |

### 2.3 支持的模型

通过 Kiro Gateway 可访问的模型（`GET /v1/models`）：

| 模型 ID | 说明 |
| --- | --- |
| `claude-sonnet-4.5` | Claude Sonnet 4.5 |
| `claude-sonnet-4.6` | Claude Sonnet 4.6 |
| `deepseek-3.2` | DeepSeek 3.2 |
| `claude-haiku-4.5` | Claude Haiku（轻量） |
| `auto-kiro` | 自动选择 |

---

## 3. 通信协议

### 3.1 请求

```http
POST http://localhost:58575/v1/chat/completions
Authorization: Bearer <PROXY_API_KEY>
Content-Type: application/json

{
  "model": "claude-sonnet-4.5",
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ],
  "temperature": 0,
  "tools": [...],
  "tool_choice": {"function": {"name": "xxx"}}
}
```

### 3.2 响应

```json
{
  "id": "msg-xxx",
  "object": "chat.completion",
  "created": 1778392811,
  "model": "claude-sonnet-4.5",
  "choices": [{
    "index": 0,
    "message": {"role": "assistant", "content": "..."},
    "finish_reason": "stop"
  }],
  "usage": {"prompt_tokens": 39, "completion_tokens": 105, "total_tokens": 144}
}
```

### 3.3 关键字段

| 字段 | 用途 |
| --- | --- |
| `response.model` | 实际执行的模型名，用于检测模型替换 |
| `finish_reason` | 正常应为 `stop`，异常可能是 `length` 或空 |
| `usage` | Token 消耗统计 |

---

## 4. 代码架构

### 4.1 核心模块

```
internal/
├── config/config.go      配置解析
│   - ProviderConfig 结构体
│   - ResolvedBaseURL() / ResolvedModel() / ResolvedAPIKey()
│   - 支持直接值和环境变量两种注入方式
│
├── provider/openai.go    API 客户端
│   - 基于 eino-ext/components/model/openai
│   - Client.Do(ctx, Request) → Response
│   - WithResponseMessageModifier 拦截原始响应提取 returned_model
│
├── runner/runner.go      执行引擎
│   - Run(ctx, Config) → RunResult
│   - 遍历 providers × cases × repeats
│   - 自动脱敏所有输出
│
├── suite/default.go      内置 probe case
│   - 11 个内置测试用例
│   - 每个 case 包含 prompt + evaluator
│
├── dataset/dataset.go    JSONL 数据集加载
│   - 从 benchmarks/starter/*.jsonl 加载
│
└── report/report.go      报告生成
    - JSON / Markdown / HTML 三种格式
    - BenchmarkSummary + StarterBaselineBand
```

### 4.2 请求流程

```
main.go
  → loadConfig() 解析配置
  → runner.Run(ctx, cfg)
    → 对每个 provider:
      → provider.NewClient(providerCfg)
      → 对每个 case:
        → suite.BuildMany(caseCfg) 构建请求
        → client.Do(ctx, req) 发送请求
        → built.Evaluate(resp) 验证响应
      → summarizeProvider() 汇总结果
    → attachCrossProviderSignals() 跨 provider 对比
  → 输出 JSON/Markdown/HTML 报告
```

### 4.3 注水检测逻辑

```go
// 1. returned_model 检测
if len(summary.DistinctReturnedModels) > 1 {
    // 警告：模型不稳定
}

// 2. Reference Score 检测
if passRate < referenceScore * 0.8 {
    // 注水嫌疑
}

// 3. Starter Baseline Band
// weak < acceptable < strong
```

---

## 5. 运维操作

### 5.1 启动 Kiro Gateway

```bash
cd /root/kiro-gateway
source .venv/bin/activate
python main.py --port 58575
```

### 5.2 运行评测

```bash
cd /root/llm-provider-probe

# 读取 gateway 的 API key
export OPENAI_API_KEY=$(grep "PROXY_API_KEY" /root/kiro-gateway/.env | cut -d'"' -f2)
export BASE_URL="http://localhost:58575/v1"
export MODEL="deepseek-3.2"

# 快速监测
./scripts/run_eino_monitoring.sh

# 全量评测
./scripts/run_deepseek_full_eval.sh
```

### 5.3 查看报告

```bash
# 最新报告
ls -lt artifacts/*.md | head -3

# 历史趋势
make history
```

### 5.4 定期监测（cron）

```bash
# 每天 9:00 跑一次监测
0 9 * * * cd /root/llm-provider-probe && \
  export OPENAI_API_KEY="xxx" && \
  export BASE_URL="http://localhost:58575/v1" && \
  export MODEL="deepseek-3.2" && \
  ./scripts/run_eino_monitoring.sh
```

---

## 6. 评测结论（2026-05-10）

### 6.1 测试矩阵

| 模型 | 测试集 | 总分 | 通过率 | 注水 |
| --- | --- | --- | --- | --- |
| claude-sonnet-4.5 | 全量 68 case | 80.9 | 64.7% | ❌ 无 |
| deepseek-3.2 | 全量 68 case | 83.8 | 76.5% | ❌ 无 |

### 6.2 各能力维度

| 能力 | Claude | DeepSeek | 说明 |
| --- | --- | --- | --- |
| 常识推理 | 100% | 100% | 两者均 strong |
| 逻辑推理 | 100% | 100% | 两者均 strong |
| 长上下文 (4K/12K/32K) | 100% | 100% | 两者均 strong |
| JSON 格式 | 100% | 100% | 两者均 strong |
| Tool Calling | 50% | 100% | DeepSeek 更好 |
| GPQA Diamond (极难) | 33% | 58% | 题目本身极难 |
| MMLU-Pro Real (难) | 45% | 70% | DeepSeek 更强 |

### 6.3 已知问题

| 问题 | 原因 | 影响 |
| --- | --- | --- |
| webqa 全部 0% | evaluator 用 exact_match，模型输出完整句子 | 误报，不影响注水判定 |
| Claude chinese_compact 失败 | 安全过滤拒绝执行 | Claude 特有 |
| DeepSeek 幻觉 web_search | 模型尝试调用不存在的工具 | 部分题目无法评分 |

---

## 7. 对接新 Provider

### 7.1 前提条件

新 provider 必须兼容 OpenAI Chat Completions API：
- `POST /v1/chat/completions`
- Bearer Token 认证
- 标准 JSON 请求/响应格式

### 7.2 步骤

1. 设置环境变量：

```bash
export BASE_URL="https://new-provider.com/v1"
export MODEL="target-model"
export OPENAI_API_KEY="new-key"
```

2. 冒烟测试：

```bash
./provider-probe -base-url "$BASE_URL" -model "$MODEL" \
  -api-key-env "OPENAI_API_KEY" -cases "exact_json,logic_filter"
```

3. 全量评测：

```bash
./scripts/run_deepseek_full_eval.sh
```

4. 检查报告中的 `returned_model` 和 `suspicion`。

### 7.3 判断标准

| 信号 | 正常 | 异常 |
| --- | --- | --- |
| returned_model | 与请求一致 | 返回不同模型名 |
| 核心 case 通过率 | ≥ 80% | < 60% |
| 延迟 | 1-20s | < 100ms（缓存）或 > 60s |
| error_runs | 0 | ≥ 2 |

---

## 8. 源码索引

| 文件 | 职责 | 修改频率 |
| --- | --- | --- |
| `cmd/provider-probe/main.go` | CLI 入口、参数解析 | 低 |
| `internal/config/config.go` | 配置结构体 | 低 |
| `internal/provider/openai.go` | API 通信核心 | 低 |
| `internal/runner/runner.go` | 执行引擎、脱敏 | 中 |
| `internal/suite/default.go` | 内置 case 定义 | 高（新增 case） |
| `internal/dataset/dataset.go` | 数据集加载 | 低 |
| `internal/report/report.go` | 报告渲染 | 中 |
| `examples/*.json` | 运行配置模板 | 高 |
| `benchmarks/starter/*.jsonl` | 评测数据集 | 中（新增数据） |
| `scripts/*.sh` | 运行脚本 | 中 |
