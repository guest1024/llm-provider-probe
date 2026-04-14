.PHONY: test list-cases smoke history history-files modelscope ab

test:
	go test ./...

list-cases:
	go run ./cmd/provider-probe -list-cases

smoke:
	./scripts/smoke.sh

history:
	./scripts/history_summary.sh artifacts

history-files:
	./scripts/render_history.sh artifacts artifacts/history-summary

modelscope:
	./scripts/run_modelscope.sh --repeat 1

ab:
	./scripts/run_ab.sh examples/modelscope-vs-openai-template.json ab-check
