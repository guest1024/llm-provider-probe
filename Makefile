.PHONY: test list-cases smoke history benchmark-history history-files benchmark-history-files modelscope ab audit-secrets eino-starter eino-monitoring convert-samples

test:
	go test ./...

list-cases:
	go run ./cmd/provider-probe -list-cases

smoke:
	./scripts/smoke.sh

history:
	./scripts/history_summary.sh artifacts

benchmark-history:
	./scripts/history_summary.sh artifacts benchmark

history-files:
	./scripts/render_history.sh artifacts artifacts/history-summary

benchmark-history-files:
	./scripts/render_history.sh artifacts artifacts/history-benchmark-summary benchmark

modelscope:
	./scripts/run_modelscope.sh --repeat 1

ab:
	./scripts/run_ab.sh examples/modelscope-vs-openai-template.json ab-check

eino-starter:
	./scripts/run_eino_starter.sh

eino-monitoring:
	./scripts/run_eino_monitoring.sh

convert-samples:
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/commonsenseqa.mapping.json --input benchmarks/source-samples/commonsenseqa-sample.jsonl --output /tmp/commonsenseqa.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/logiqa.mapping.json --input benchmarks/source-samples/logiqa-sample.jsonl --output /tmp/logiqa.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/webqa.mapping.json --input benchmarks/source-samples/webqa-sample.jsonl --output /tmp/webqa.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/mmlu-pro.mapping.json --input benchmarks/source-samples/mmlu-pro-sample.jsonl --output /tmp/mmlu-pro.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/gpqa.mapping.json --input benchmarks/source-samples/gpqa-sample.jsonl --output /tmp/gpqa.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/bbh-logical-deduction.mapping.json --input benchmarks/source-samples/bbh-logical-deduction-sample.jsonl --output /tmp/bbh-logical-deduction.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/ruler.mapping.json --input benchmarks/source-samples/ruler-sample.jsonl --output /tmp/ruler.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/brainteaser-zh.mapping.json --input benchmarks/source-samples/brainteaser-zh-sample.csv --output /tmp/brainteaser-zh.converted.jsonl
	python3 scripts/convert_dataset.py --mapping benchmarks/mappings/bfcl-style-tool.mapping.json --input benchmarks/source-samples/bfcl-style-tool-sample.jsonl --output /tmp/bfcl-style-tool.converted.jsonl

audit-secrets:
	./scripts/audit_secrets.sh
