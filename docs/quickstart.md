# Quick Start

## 1. 运行本地自测

```bash
./scripts/smoke.sh
```

这会执行：

- `go test ./...`
- `go run ./cmd/provider-probe -list-cases`
- 如果存在历史报告，则执行一次 compare 自检
- 如果环境变量 `MODELSCOPE_API_KEY` 已设置，则额外跑一轮真实 ModelScope 探针

## 2. 跑单次探针

```bash
./scripts/run_probe.sh --config examples/config.json --label baseline
```

输出会落到：

- `artifacts/<label>-<timestamp>.json`
- `artifacts/<label>-<timestamp>.md`
- `artifacts/<label>-<timestamp>.html`

## 3. 跑 ModelScope / Qwen

```bash
export MODELSCOPE_API_KEY=你的key
./scripts/run_modelscope.sh --repeat 2
```

## 4. 对比两次运行

```bash
./scripts/compare_reports.sh artifacts/a.json artifacts/b.json
```

## 5. 查看历史趋势

```bash
./scripts/history_summary.sh artifacts
```

历史汇总会直接给出：

- latest score / suspicion
- volatility
- 自动 verdict

## 6. 生成历史 Markdown / HTML 汇总

```bash
./scripts/render_history.sh artifacts artifacts/history-summary
```

## 7. 跑 target vs baseline A/B

```bash
export MODELSCOPE_API_KEY=你的key
export OPENAI_API_KEY=你的key
./scripts/run_ab.sh examples/modelscope-vs-openai-template.json ab-check
```

## 8. 只跑特定 case

```bash
go run ./cmd/provider-probe \
  -config examples/modelscope-qwen.json \
  -cases exact_json,nested_json_schema,go_snippet_output
```
