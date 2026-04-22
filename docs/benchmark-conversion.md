# Benchmark Conversion Guide

本项目内部统一使用 **provider-probe JSONL** 作为评测样本格式。

为了方便把公开 benchmark 或内部 CSV/JSON/JSONL 数据转进来，仓库提供：

- `scripts/convert_dataset.py`
- `benchmarks/mappings/*.mapping.json`
- `benchmarks/source-samples/*`

## 转换命令

### CommonsenseQA 风格样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/commonsenseqa.mapping.json \
  --input benchmarks/source-samples/commonsenseqa-sample.jsonl \
  --output /tmp/commonsenseqa.converted.jsonl
```

### LogiQA 风格样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/logiqa.mapping.json \
  --input benchmarks/source-samples/logiqa-sample.jsonl \
  --output /tmp/logiqa.converted.jsonl
```

### WebQA 风格样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/webqa.mapping.json \
  --input benchmarks/source-samples/webqa-sample.jsonl \
  --output /tmp/webqa.converted.jsonl
```

### MMLU-Pro / GPQA 风格样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/mmlu-pro.mapping.json \
  --input benchmarks/source-samples/mmlu-pro-sample.jsonl \
  --output /tmp/mmlu-pro.converted.jsonl

python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/gpqa.mapping.json \
  --input benchmarks/source-samples/gpqa-sample.jsonl \
  --output /tmp/gpqa.converted.jsonl
```

### RULER 风格 retrieval 样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/ruler.mapping.json \
  --input benchmarks/source-samples/ruler-sample.jsonl \
  --output /tmp/ruler.converted.jsonl
```

### 中文脑筋急转弯 CSV 样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/brainteaser-zh.mapping.json \
  --input benchmarks/source-samples/brainteaser-zh-sample.csv \
  --output /tmp/brainteaser-zh.converted.jsonl
```

### BFCL 风格 tool-calling 样本

```bash
python3 scripts/convert_dataset.py \
  --mapping benchmarks/mappings/bfcl-style-tool.mapping.json \
  --input benchmarks/source-samples/bfcl-style-tool-sample.jsonl \
  --output /tmp/bfcl-style-tool.converted.jsonl
```

## mapping 文件字段

| 字段 | 说明 |
| --- | --- |
| `input_format` | `jsonl` / `json` / `csv` |
| `benchmark` | 输出 benchmark 名 |
| `split` | 输出 split |
| `category` | 输出 category |
| `evaluator` | 默认 evaluator |
| `field_map` | 原始字段 -> 标准字段映射；除了 `id/prompt` 外，其它字段会按目标字段名直接拷贝，可用于 `tools` / `tool_choice` / `expected_tool_calls` |
| `choice_fields` | 多选项映射 |
| `prompt_template` | 用模板直接拼 prompt |
| `acceptable_answer_fields` | 可接受答案字段列表 |
| `metadata_fields` | 额外 metadata 映射 |
| `defaults` | 输出时附加的默认字段 |

## 当前提供的 mapping 模板

- `benchmarks/mappings/commonsenseqa.mapping.json`
- `benchmarks/mappings/logiqa.mapping.json`
- `benchmarks/mappings/webqa.mapping.json`
- `benchmarks/mappings/mmlu-pro.mapping.json`
- `benchmarks/mappings/gpqa.mapping.json`
- `benchmarks/mappings/bbh-logical-deduction.mapping.json`
- `benchmarks/mappings/ruler.mapping.json`
- `benchmarks/mappings/brainteaser-zh.mapping.json`
- `benchmarks/mappings/bfcl-style-tool.mapping.json`
- `benchmarks/mappings/custom-regression.mapping.json`

## 建议流程

1. 先用 source sample 小样本验证转换结果
2. 检查输出 JSONL 是否满足 `docs/custom-dataset-standard.md`
3. 用 `examples/custom-dataset-template.json` 或新 config 跑一轮 dry run
4. 再接入更大完整数据集
