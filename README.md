# provider-probe

一个基于 **Eino** 的 LLM 评测/探针工具。

它现在同时覆盖两层能力：

1. **probe cases**：快速探测 provider 的格式稳定性、逻辑、长上下文、tool calling
2. **benchmark-style datasets**：用统一 JSONL 标准接入通用 QA、WebQA、逻辑题、脑筋急转弯、工具调用等 eval 数据集

## 当前定位

它不是纯 A/B 工具，也不是完整 leaderboard 框架；它更像：

- 一个可复现的 **Eino-based eval runner**
- 一个 provider 稳定性/退化探针
- 一个可扩展的内部回归集执行器

## 核心能力

- 基于 `github.com/cloudwego/eino` + `github.com/cloudwego/eino-ext/components/model/openai`
- 支持 OpenAI-compatible Base URL / API key / model 通过环境变量注入
- 保留原有 probe cases
- 新增 dataset-backed benchmark case 支持
- 支持 `exact_match / regex_match / multiple_choice / tool_call` evaluator
- 输出 JSON / Markdown / HTML 报告
- 支持 compare / history 汇总（provider / benchmark / provider-benchmark）
- 支持自定义内部 JSONL 数据集接入

## 推荐环境变量

```bash
export OPENAI_API_KEY="<your key>"
export BASE_URL="https://vibediary.app/api/v1"
export MODEL="gpt-5.4"
```

> 不要把真实 key 写进配置文件、README、脚本或提交记录。

---

## 快速开始

### 1. 本地自测

```bash
./scripts/smoke.sh
```

### 2. 查看内置 probe case

```bash
go run ./cmd/provider-probe -list-cases
```

### 3. 跑 Eino starter benchmark

```bash
./scripts/run_eino_starter.sh
```

它会读取：

- `OPENAI_API_KEY`
- `BASE_URL`
- `MODEL`

并执行 `examples/eino-benchmark-starter.json` 中的 starter benchmark。

### 4. 跑粗粒度参水监测配置

```bash
./scripts/run_eino_monitoring.sh
```

它会执行一个更轻量的监测矩阵，目标是快速发现**严重参水/明显降级**，而不是追求和官方 benchmark 高度对齐。

### 5. 跑单 provider 配置

```bash
go run ./cmd/provider-probe -config examples/eino-benchmark-starter.json
```

### 6. 跑自定义内部数据集

```bash
go run ./cmd/provider-probe -config examples/custom-dataset-template.json
```

### 7. 对比两次运行

```bash
./scripts/compare_reports.sh artifacts/run-a.json artifacts/run-b.json
```

### 8. 查看历史趋势

```bash
./scripts/history_summary.sh artifacts
./scripts/history_summary.sh artifacts benchmark
```

---

## 当前支持的 probe case

- `exact_json`
- `exact_line`
- `logic_filter`
- `chinese_compact`
- `nested_json_schema`
- `response_format_json_schema`
- `go_snippet_output`
- `tool_call_echo`
- `long_context_needle_small`
- `long_context_needle_medium`
- `long_context_needle_large`

## 当前附带的 starter benchmark-style datasets

- `benchmarks/starter/commonsenseqa-starter.jsonl`
- `benchmarks/starter/mmlu-pro-starter.jsonl`
- `benchmarks/starter/gpqa-starter.jsonl`
- `benchmarks/starter/logiqa-starter.jsonl`
- `benchmarks/starter/bbh-logical-deduction-starter.jsonl`
- `benchmarks/starter/bbh-tracking-objects-starter.jsonl`
- `benchmarks/starter/webqa-starter.jsonl`
- `benchmarks/starter/ruler-retrieval-starter.jsonl`
- `benchmarks/starter/brainteaser-zh-starter.jsonl`
- `benchmarks/starter/bfcl-style-tool-starter.jsonl`

这些 starter 集的用途是：

- 验证框架接入
- 做小规模回归
- 给 provider 一个初级能力基线
- 为后续接更大公开 benchmark 打样

它们**不是**官方 leaderboard 成绩替代。

---

## 自定义 dataset 接入

项目支持通过 `suite.cases[].dataset` 指向 JSONL 数据集。

最小配置示例：

```json
{
  "name": "company-regression",
  "enabled": true,
  "dataset": {
    "path": "benchmarks/custom/example-company-regression.jsonl",
    "name": "company_regression",
    "split": "dev"
  }
}
```

支持的 evaluator：

- `exact_match`
- `regex_match`
- `multiple_choice`
- `tool_call`

详细字段说明见：

- `docs/custom-dataset-standard.md`
- `benchmarks/custom/example-company-regression.jsonl`
- `examples/custom-dataset-template.json`

## Benchmark 转换工具

仓库提供原始 benchmark -> 本项目 JSONL 的转换脚本：

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/commonsenseqa.mapping.json \
  --input benchmarks/source-samples/commonsenseqa-sample.jsonl \
  --output /tmp/commonsenseqa.converted.jsonl
```

更多说明见：

- `docs/benchmark-conversion.md`
- `benchmarks/mappings/*.mapping.json`
- `benchmarks/source-samples/*`

当前附带的 mapping 样例包括：

- `commonsenseqa`
- `mmlu-pro`
- `gpqa`
- `logiqa`
- `webqa`
- `ruler-retrieval`
- `bbh-logical-deduction`
- `brainteaser-zh`
- `bfcl-style-tool`
- `custom-regression`

---

## Starter baseline 建议

见 `docs/benchmarks.md`。

当前仓库给的是 **starter baseline band**，不是学术 leaderboard：

- `commonsenseqa-starter`
- `mmlu-pro-starter`
- `gpqa-starter`
- `logiqa-starter`
- `webqa-starter`
- `ruler-retrieval-starter`
- `cn_brainteaser`
- `bfcl-style-tool-starter`

用于把模型表现粗分为：

- weak
- acceptable
- strong

这些 band 现在也会出现在 provider report 的 benchmark summary 中，方便直接看“当前模型对这个 starter 子集处在哪个初级档位”。

如果你的目标只是发现**严重参水**，优先看：

- 是否掉到 `weak`
- 是否多个 benchmark 同时变弱
- 是否 probe 也同时失败

不需要追求和预期分数完全严丝合缝。

---

## 输出解释

每轮运行会输出：

- `score`
- `suspicion`
- `warnings`
- `*.json`
- `*.md`
- `*.html`
- 每个 sample/case 的：
  - `benchmark`
  - `split`
  - `sample_id`
  - `latency_ms`
  - `status_code`
  - `returned_model`
  - `finish_reason`
  - `prompt_tokens / completion_tokens / total_tokens`
- `raw_response_snippet`
- benchmark 维度汇总：
  - `attempts / passes / errors`
  - `pass_rate`
  - `avg_score`
  - `starter_baseline_band`

---

## 目录说明

- `internal/provider`：Eino-based model invocation
- `internal/suite`：built-in probe cases + dataset case expansion
- `internal/dataset`：JSONL dataset loader
- `internal/report`：报告结构与渲染
- `benchmarks/starter`：starter benchmark fixtures
- `benchmarks/custom`：自定义数据集范例
- `examples/*.json`：运行配置模板
- `docs/benchmarks.md`：benchmark 覆盖与 starter baseline
- `docs/custom-dataset-standard.md`：自定义数据集接入规范
- `docs/monitoring.md`：粗粒度参水监测用法

---

## 常用命令

```bash
make test
make list-cases
make eino-starter
make eino-monitoring
make history
make benchmark-history
make history-files
make benchmark-history-files
make audit-secrets
```

---

## 后续扩展建议

优先顺序：

1. 接更完整的公开 benchmark 转换脚本
2. 增加更多 evaluator（如 citation / evidence span / partial credit）
3. 增加 live web browsing / browser-agent WebQA
4. 增加 multi-step tool execution eval
