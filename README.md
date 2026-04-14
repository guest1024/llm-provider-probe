# provider-probe

一个用于判断当前 LLM provider 是否“参水”的 Go 版探针工具。

## 目标

它不是看模型“像不像”，而是输出一条证据链：

- 返回模型标识是否稳定
- 格式约束是否稳定
- 推理/指令跟随是否稳定
- 长上下文检索是否早衰
- 多 provider 同题 A/B 下是否明显落后

## 设计原则

- **快速部署**：纯 Go 标准库，无第三方依赖
- **快速测试**：内置默认测试集，可直接跑
- **易配置**：支持命令行单 provider 模式和 JSON 配置文件模式
- **易对比**：支持 Markdown/HTML 报告、compare 和 history 汇总
- **易扩展**：provider 层抽象为 OpenAI-compatible HTTP adapter；后续如果要接 Eino，可在 `internal/provider` 增加 adapter，不影响 suite 和 report

> 当前 MVP 没有强制引入 Eino，是为了减少依赖和部署复杂度；如果后续你要接多模型 SDK、Tool/Agent 编排或更复杂的评测流，再把 Eino 接到 provider adapter 层最合适。

---

## 快速开始

### 1. 一键自测

```bash
./scripts/smoke.sh
```

### 2. 单 provider 快速跑

```bash
go run ./cmd/provider-probe \
  -base-url https://api.openai.com/v1 \
  -model gpt-4.1-mini \
  -api-key-env OPENAI_API_KEY \
  -repeat 2 \
  -out artifacts/quick.json
```

### 3. 多 provider A/B 跑

```bash
export TARGET_PROVIDER_API_KEY=xxx
export OPENAI_API_KEY=xxx

go run ./cmd/provider-probe -config examples/config.json
```

### 4. 用脚本跑 ModelScope / Qwen

```bash
export MODELSCOPE_API_KEY=你的_key
./scripts/run_modelscope.sh --repeat 2
```

### 5. 查看内置测试项

```bash
go run ./cmd/provider-probe -list-cases
```

### 6. 对比两次报告

```bash
./scripts/compare_reports.sh artifacts/run-a.json artifacts/run-b.json
```

### 7. 查看历史趋势

```bash
./scripts/history_summary.sh artifacts
```

history 输出会附带：
- latest score / suspicion
- volatility
- 自动 verdict

### 8. 渲染历史汇总文件

```bash
./scripts/render_history.sh artifacts artifacts/history-summary
```

### 9. 跑官方 baseline A/B

```bash
export MODELSCOPE_API_KEY=xxx
export OPENAI_API_KEY=xxx
./scripts/run_ab.sh examples/modelscope-vs-openai-template.json ab-check
```

### 10. 只跑指定能力 case

```bash
go run ./cmd/provider-probe \
  -config examples/modelscope-qwen.json \
  -cases exact_json,nested_json_schema,go_snippet_output
```

---

## 输出解释

工具会输出：

- `score`：综合分，0~100
- `suspicion`：`low | medium | high`
- `warnings`：疑点说明
- `*.md`：自动生成的人类可读结论报告
- `*.html`：可直接打开的 HTML 报告
- `error_runs`：本轮中出现错误响应的次数
- `history verdict`：历史汇总里根据波动/错误/高风险次数给出的自动结论
- 报告中的敏感 header / 常见 key 形态会自动脱敏
- 每次 case 的：
  - `latency_ms`
  - `status_code`
  - `returned_model`
  - `finish_reason`
  - `prompt_tokens / completion_tokens / total_tokens`
  - `response_headers`
  - `raw_response_snippet`

### Suspicion 解释

- `low`：暂时没有明显证据说明 provider 在降配/混模
- `medium`：存在一致性问题，需要扩大样本复测
- `high`：有较强证据显示输出质量、长上下文或模型标识存在明显异常

### CI / 自动告警

可以加 `-fail-on`：

```bash
go run ./cmd/provider-probe -config examples/modelscope-qwen.json -fail-on high
```

当任一 provider 的 `suspicion` 达到阈值时，进程会返回非 0。

---

## 建议测试流程

### 最小可用版

1. 跑默认 8 个 case
2. 每个 case 跑 3 次
3. target provider 和 official baseline 同时跑
4. 如果 target 明显落后 15 分以上，再扩到：
   - `long_context_needle_large`
   - 不同时段重复跑
   - 增加自定义 case

### 时段降配检测

当前工具负责“单次矩阵采样”。
要检查高峰期动态降配，建议用 cron/CI 在以下时段重复执行并比较 `artifacts/*.json`：

- 凌晨
- 上午
- 晚高峰
- 周末高峰

也可以直接用脚本批量打标签：

```bash
./scripts/timeslot_batch.sh --config examples/modelscope-qwen.json --repeat 1
```

---

## 配置文件说明

见 `examples/config.json`。

重点字段：

- `providers[].base_url`：API 基础地址
- `providers[].model`：模型名
- `providers[].api_key_env`：从环境变量取 key
- `providers[].headers`：附加 header
- `providers[].extra_body`：给请求体附加 provider 私有参数
- `suite.cases[].params.approx_tokens`：长上下文 case 近似 token 数
- `run.output`：JSON 报告输出路径；对应 Markdown/HTML 会默认落到同名前缀

---

## 脚本

- `scripts/run_probe.sh`：通用单次运行入口
- `scripts/run_modelscope.sh`：ModelScope/Qwen 快捷入口
- `scripts/compare_reports.sh`：报告对比
- `scripts/history_summary.sh`：历史趋势汇总
- `scripts/render_history.sh`：历史趋势落盘为 md/html
- `scripts/run_ab.sh`：target vs baseline 一键 A/B
- `scripts/timeslot_batch.sh`：按时段标签批跑
- `scripts/audit_secrets.sh`：扫描仓库中疑似密钥/Token 形态
- `scripts/smoke.sh`：本地/CI 自检

---

## 文档

- `docs/quickstart.md`
- `docs/methodology.md`
- `docs/modelscope.md`
- `docs/ops.md`
- `docs/security.md`
- `examples/openai-baseline-template.json`
- `examples/modelscope-vs-openai-template.json`

---

## Makefile

常用入口：

```bash
make test
make list-cases
make history
make history-files
make audit-secrets
```

---

## 新增内置 case

- `chinese_compact`：中文严格单行输出
- `nested_json_schema`：嵌套 JSON 结构遵循
- `response_format_json_schema`：基于 `response_format=json_schema` 的严格结构输出
- `go_snippet_output`：Go 代码阅读/执行结果推断
- `tool_call_echo`：函数/工具调用返回能力探针

---

## 后续扩展建议

### 1. 自定义测试集

可扩展 `internal/suite`：

- 中文格式约束
- 代码修复小题
- JSON schema 模式
- Tool calling / function calling
- vision case

### 2. Eino 接入点

如需接 Eino，建议做法：

- 在 `internal/provider` 新增 `eino_adapter.go`
- 保持 `Run(ctx, cfg)` 和 suite evaluator 不变
- 用配置决定走 `openai-compatible` 还是 `eino`

### 3. 报表聚合

当前已支持：

- `compare` 子命令
- `history` 子命令
- 自动输出 markdown/html 结论

下一步可以继续加：

- 历史运行趋势图
- 不同时段漂移分析
- 更细粒度的 case 维度趋势
