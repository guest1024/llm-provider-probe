# Custom Dataset Integration Standard

The project accepts dataset-backed eval cases through JSONL files.
Each line is one sample.

## Minimum fields

```json
{
  "id": "sample-001",
  "benchmark": "team_regression",
  "split": "dev",
  "category": "policy_qa",
  "prompt": "What is the SLA for P1 incidents?",
  "evaluator": "exact_match",
  "expected": "15 minutes"
}
```

## Supported sample fields

| Field | Required | Description |
| --- | --- | --- |
| `id` | recommended | Stable sample id for regression diffing |
| `benchmark` | recommended | Benchmark/group name |
| `split` | optional | e.g. `dev`, `test`, `starter` |
| `category` | optional | Capability slice |
| `system_prompt` | optional | System instruction |
| `context` | optional | Evidence/snippets/context prepended before prompt |
| `prompt` | yes | User-facing question/instruction |
| `choices` | optional | Multiple-choice options |
| `expected` | optional | Canonical expected answer |
| `acceptable_answers` | optional | Alternate exact answers |
| `regex` | optional | Regex for `regex_match` |
| `evaluator` | yes | `exact_match`, `regex_match`, `multiple_choice`, `tool_call` |
| `required_substrings` | optional | Substrings that must appear in answer |
| `tools` | optional | OpenAI-compatible tool schema array |
| `tool_choice` | optional | OpenAI-compatible tool choice |
| `expected_tool_calls` | optional | Expected tool name + argument subset |
| `metadata` | optional | Free-form custom metadata |

## Evaluator guidance

### `exact_match`
Use for short factual answers or policy values.

### `regex_match`
Use for stable formats like IDs, dates, slugs, and templated output.

### `multiple_choice`
Use for QA/logic tasks with A/B/C/D labels.
Expected values can be either the label (`"B"`) or an acceptable answer alias.

### `tool_call`
Use when the response should be a tool call instead of plain text.
`expected_tool_calls[0].arguments` is treated as a required subset of the generated arguments.

## Example command

```bash
go run ./cmd/provider-probe -config examples/custom-dataset-template.json
```

## Reference example
See:
- `benchmarks/custom/example-company-regression.jsonl`
- `examples/custom-dataset-template.json`
