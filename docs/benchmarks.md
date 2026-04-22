# Benchmark Coverage

## Current layers

### 1. Probe cases (handcrafted)
- exact_json
- exact_line
- logic_filter
- chinese_compact
- nested_json_schema
- response_format_json_schema
- go_snippet_output
- tool_call_echo
- long_context_needle_*

### 2. Starter benchmark-style datasets
- `commonsenseqa-starter`: common-sense MCQ smoke subset
- `mmlu-pro-starter`: broad knowledge / academic QA smoke subset
- `gpqa-starter`: harder science/general-professional QA smoke subset
- `logiqa-starter`: logic reasoning MCQ subset
- `bbh-logical-deduction-starter`: BBH 逻辑推理子集 starter
- `bbh-tracking-objects-starter`: BBH 对象追踪子集 starter
- `webqa-starter`: offline web-grounded QA subset based on provided snippets
- `ruler-retrieval-starter`: long-context retrieval / needle lookup starter
- `brainteaser-zh-starter`: Chinese brainteaser / 急转弯 subset
- `bfcl-style-tool-starter`: function-calling starter subset

These starter sets are intentionally small. They are for:
- framework validation
- regression testing
- provider smoke checks
- internal baseline calibration
- severe-regression / “参水” monitoring

They are **not** a claim of official leaderboard parity.

每次 provider report 会额外输出 benchmark summary：

- attempts / passes / errors
- pass rate
- avg score
- starter baseline band（weak / acceptable / strong）

## Additional public benchmarks to extend next
- MMLU-Pro / GPQA full official splits: larger knowledge / science coverage
- LogiQA 2.0 / more BBH subsets: stronger logic coverage
- WebWalkerQA / browser-style WebQA: live web grounding
- BFCL full / multi-turn tool evals: deeper tool-calling quality
- RULER larger contexts / InfiniteBench: longer context retrieval and synthesis

## Starter baseline bands
For the included starter subsets, treat these as rough bands rather than strict leaderboard claims:

| Benchmark | Weak | Acceptable | Strong |
| --- | --- | --- | --- |
| commonsenseqa-starter | < 50% | 50%~79% | >= 80% |
| mmlu-pro-starter | < 45% | 45%~74% | >= 75% |
| gpqa-starter | < 35% | 35%~64% | >= 65% |
| logiqa-starter | < 40% | 40%~69% | >= 70% |
| bbh-logical-deduction-starter | < 40% | 40%~69% | >= 70% |
| bbh-tracking-objects-starter | < 40% | 40%~69% | >= 70% |
| webqa-starter | < 50% | 50%~79% | >= 80% |
| ruler-retrieval-starter | < 50% | 50%~79% | >= 80% |
| brainteaser-zh-starter | < 40% | 40%~69% | >= 70% |
| bfcl-style-tool-starter | < 50% | 50%~89% | >= 90% |

这些 band 的主要用途是：

- 看是否**大致正常**
- 看是否出现**明显塌陷**
- 给“是否可能严重参水”提供一个粗粒度信号

不建议把它们当作精确能力刻度。

## Suggested default env for the current target provider
Use environment variables instead of storing secrets in config files:

```bash
export OPENAI_API_KEY="<your key>"
export BASE_URL="https://vibediary.app/api/v1"
export MODEL="gpt-5.4"
```

Then run:

```bash
./scripts/run_eino_starter.sh
```
