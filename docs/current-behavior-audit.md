# 当前行为审计（2026-04-14）

## 审计对象

`provider-probe` 当前工作树与最近一轮真实探针结果。

## 当前实现状态

当前工具已经具备以下能力：

- 单 provider / 多 provider 配置运行
- JSON / Markdown / HTML 报告输出
- history 汇总与自动 verdict
- compare 对比与 case 级分数变化
- 时段批跑与 A/B 脚本
- OpenAI-compatible 响应解析
- `response_format=json_schema` 探针
- tool calling 探针
- 长上下文 needle 检索

## 当前新鲜验证证据

- `go test ./...`：PASS
- `go build ./...`：PASS
- `bash -n scripts/*.sh`：PASS
- `go run ./cmd/provider-probe -list-cases`：PASS
- `go run ./cmd/provider-probe history -dir artifacts -provider modelscope-qwen-397b`：PASS

## 当前被审计的 provider 行为

基于现有 `artifacts/` 历史记录，ModelScope Qwen 目标项的当前汇总为：

- Reports: `9`
- Score avg/min/max: `58.4 / 25.0 / 100.0`
- Latest score/suspicion: `40.0 / high`
- Volatility: `75.0`
- Suspicion medium/high counts: `3 / 5`
- Total error runs: `15`
- Verdict: `高波动/高风险，优先做官方基线 A/B 并持续监控`

## 审计结论

1. **当前 probe 本身可用且验证通过**，没有发现新的本地构建/测试问题。
2. **外部 provider 稳定性波动明显**，不是单次偶发：历史分数跨度极大，并多次出现高风险判词。
3. **最关键异常仍是 `HTTP 200 + empty choices payload`**，这更像网关/后端退化或不稳定，而不是 probe 解析错误。
4. **下一步最有价值动作**是补齐官方 baseline A/B 的真实运行证据，而不是继续扩展更多本地功能。

## 审计建议

- 尽快使用 `examples/modelscope-vs-openai-template.json` 跑一轮真实 A/B。
- 按凌晨 / 上午 / 晚高峰 / 周末高峰继续做定时采样。
- 将 history verdict 作为告警输入，而不是单看单次 score。
