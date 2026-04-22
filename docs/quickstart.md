# Quick Start

## 1. 运行本地自测

```bash
./scripts/smoke.sh
```

## 2. 设置当前目标 provider 环境变量

```bash
export OPENAI_API_KEY="<your key>"
export BASE_URL="https://vibediary.app/api/v1"
export MODEL="gpt-5.4"
```

## 3. 查看内置 probe case

```bash
go run ./cmd/provider-probe -list-cases
```

## 4. 跑 starter benchmark

```bash
./scripts/run_eino_starter.sh
```

## 5. 跑粗粒度参水监测

```bash
./scripts/run_eino_monitoring.sh
```

## 6. 跑自定义回归数据集

```bash
go run ./cmd/provider-probe -config examples/custom-dataset-template.json
```

## 7. 对比两次运行

```bash
./scripts/compare_reports.sh artifacts/a.json artifacts/b.json
```

## 8. 查看历史趋势

```bash
./scripts/history_summary.sh artifacts
./scripts/history_summary.sh artifacts benchmark
```

## 9. 生成历史 Markdown / HTML 汇总

```bash
./scripts/render_history.sh artifacts artifacts/history-summary
./scripts/render_history.sh artifacts artifacts/history-benchmark-summary benchmark
```

## 10. 做一次密钥泄漏审计

```bash
./scripts/audit_secrets.sh
```

## 11. 转换外部 benchmark 样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/commonsenseqa.mapping.json \
  --input benchmarks/source-samples/commonsenseqa-sample.jsonl \
  --output /tmp/commonsenseqa.converted.jsonl

python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/webqa.mapping.json \
  --input benchmarks/source-samples/webqa-sample.jsonl \
  --output /tmp/webqa.converted.jsonl

python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/mmlu-pro.mapping.json \
  --input benchmarks/source-samples/mmlu-pro-sample.jsonl \
  --output /tmp/mmlu-pro.converted.jsonl
```
