# 注水检测原理与使用指南

## 什么是"注水"

"注水"（watermarking / model substitution）指 LLM Provider 在 API 接口背后悄悄替换或降级实际运行的模型，使得：

- 对外宣称的模型版本与实际运行的模型不符
- 实际能力明显低于宣称版本
- 用户付费购买高端模型，实际得到的是低端模型的输出

注水通常不会导致接口报错，而是表现为**能力下降**：逻辑推理变弱、知识覆盖变窄、长上下文检索失败、工具调用不稳定等。

---

## 检测原理

本工具采用**单模型绝对阈值**检测方法，不需要同时运行两个模型做 A/B 对比。

### 核心逻辑

对每个 benchmark，配置一个**最高水准参考分**（`reference_score`），代表该 benchmark 上顶级模型（如 GPT-4o、Claude 3.5 Sonnet）的预期 pass rate。

运行时，如果被测模型的实际 pass rate 低于参考分的 **80%**，则判定该 benchmark 存在注水嫌疑：

```
watermark_suspected = (actual_pass_rate < reference_score × 0.8)
```

**示例：**

| Benchmark | 参考分 | 80% 阈值 | 实际得分 | 结论 |
| --- | ---: | ---: | ---: | --- |
| mmlu_pro | 75% | 60% | 45% | ⚠️ 注水嫌疑 |
| commonsenseqa | 90% | 72% | 80% | ✅ 正常 |
| gpqa | 65% | 52% | 40% | ⚠️ 注水嫌疑 |

### 为什么是 -20%

- 顶级模型在 starter 子集上的表现本身有一定随机性（小样本波动）
- 允许 ±10% 的正常波动空间
- 低于 80% 意味着差距已经超出正常波动范围，属于**显著退化**
- 这个阈值是保守的：宁可漏报，不要误报

### 单模型检测的优势

- **不需要同时运行两个模型**：只需要一次运行，对比历史参考分
- **可重复**：每次运行结果可以和同一参考分对比，形成时序趋势
- **可自定义**：可以针对自己的业务场景配置参考分
- **低成本**：不需要维护"对照组"模型的 API 访问

---

## 配置方法

在运行配置文件的 `run` 字段中加入 `reference_scores`：

```json
{
  "run": {
    "repeats": 1,
    "temperature": 0,
    "reference_scores": {
      "commonsenseqa": 0.90,
      "mmlu_pro":      0.75,
      "gpqa":          0.65,
      "logiqa":        0.75,
      "bbh_logical_deduction":         0.90,
      "bbh_tracking_shuffled_objects": 0.85,
      "webqa":          0.85,
      "ruler_retrieval": 0.90,
      "cn_brainteaser": 0.75,
      "bfcl_style":     0.95
    }
  }
}
```

`reference_scores` 的 key 是 benchmark 名称（与 JSONL 数据集中的 `benchmark` 字段一致），value 是 0.0–1.0 的 pass rate。

**不配置 `reference_scores` 时**，工具仍然正常运行，只是不做注水检测，退化为原有的 starter band 判断。

---

## 内置参考分说明

`examples/eino-benchmark-starter.json` 和 `examples/eino-monitoring-minimal.json` 中已内置参考分，基于以下依据：

| Benchmark | 参考分 | 依据 |
| --- | ---: | --- |
| commonsenseqa | 90% | GPT-4o 在 CommonsenseQA 上约 90%+ |
| mmlu_pro | 75% | GPT-4o 在 MMLU-Pro 上约 72–78% |
| gpqa | 65% | GPT-4o 在 GPQA 上约 53–65%（diamond 子集更难） |
| logiqa | 75% | 顶级模型在 LogiQA 上约 70–80% |
| bbh_logical_deduction | 90% | GPT-4o 在 BBH 上约 85–95% |
| bbh_tracking_shuffled_objects | 85% | GPT-4o 在 BBH 上约 80–90% |
| webqa | 85% | 顶级模型在 WebQA 上约 80–90% |
| ruler_retrieval | 90% | 顶级模型在长上下文检索上约 85–95% |
| cn_brainteaser | 75% | 顶级模型在中文脑筋急转弯上约 70–80% |
| bfcl_style | 95% | GPT-4o 在 BFCL 上约 90–95% |

> 这些参考分基于 starter 子集（小样本），不是官方 leaderboard 成绩。
> 如果你的目标模型不是顶级模型，应该根据实际情况调低参考分。

---

## 报告中的注水信号

### JSON 报告

每个 `benchmark_summaries` 条目会包含：

```json
{
  "benchmark": "mmlu_pro",
  "pass_rate": 0.40,
  "reference_score": 0.75,
  "watermark_suspected": true
}
```

### Markdown / HTML 报告

Benchmark Summary 表格中会出现 `Ref Score` 和 `Watermark?` 列：

| Benchmark | Pass Rate | Ref Score | Watermark? |
| --- | ---: | ---: | --- |
| mmlu_pro | 40.0% | 75.0% | ⚠️ YES |
| commonsenseqa | 85.0% | 90.0% | |

### Suspicion 升级

只要有任意一个 benchmark 触发 `watermark_suspected`，provider 的 `suspicion` 会直接升级为 `high`，并在 warnings 中输出具体原因：

```
benchmark mmlu_pro pass rate 40.0% is below reference 75.0% × 80% = 60.0% — watermark suspected
```

---

## 自定义题目检测注水

除了使用内置 starter benchmark，你也可以用**自己的业务题目**来检测注水。这对于以下场景特别有用：

- 你有一批已知答案的内部题目
- 你想检测模型在特定领域（如法律、医疗、代码）的能力退化
- 你想用更难的题目来区分高端模型和低端模型

### 步骤

**1. 准备 JSONL 数据集**

每行一个题目，格式参考 `docs/custom-dataset-standard.md`：

```jsonl
{"id": "q001", "benchmark": "my_domain", "split": "dev", "prompt": "...", "expected": "...", "evaluator": "exact_match"}
{"id": "q002", "benchmark": "my_domain", "split": "dev", "prompt": "...", "expected": "...", "evaluator": "multiple_choice", "choices": ["A. ...", "B. ...", "C. ...", "D. ..."]}
```

**2. 在顶级模型上跑一次，记录 pass rate**

```bash
go run ./cmd/provider-probe \
  -base-url https://api.openai.com/v1 \
  -model gpt-4o \
  -api-key-env OPENAI_API_KEY \
  -config examples/custom-dataset-template.json
```

记录输出中 `my_domain` 的 pass rate，例如 `0.85`。

**3. 把这个 pass rate 作为参考分写入配置**

```json
{
  "run": {
    "reference_scores": {
      "my_domain": 0.85
    }
  },
  "suite": {
    "cases": [
      {
        "name": "my-domain-check",
        "enabled": true,
        "dataset": {
          "path": "benchmarks/custom/my-domain.jsonl",
          "name": "my_domain",
          "split": "dev"
        }
      }
    ]
  }
}
```

**4. 对被测 provider 运行**

```bash
go run ./cmd/provider-probe -config my-watermark-check.json
```

如果被测 provider 的 pass rate 低于 `0.85 × 0.8 = 0.68`，则触发注水告警。

### 选题建议

好的注水检测题目应该：

- **有明确的正确答案**（适合 exact_match 或 multiple_choice）
- **顶级模型能稳定答对**（pass rate ≥ 80%）
- **低端模型明显答不好**（pass rate ≤ 50%）
- **不依赖实时信息**（避免因知识截止日期导致误判）
- **覆盖多个能力维度**（推理、知识、代码、工具调用）

---

## 与 A/B 对比的关系

本工具**支持但不强制** A/B 对比。

| 方法 | 优点 | 缺点 |
| --- | --- | --- |
| 单模型 + 参考分 | 不需要对照组，成本低，可随时运行 | 参考分需要提前标定 |
| A/B 对比 | 不需要预设参考分，直接对比 | 需要同时维护两个 provider 的访问 |

推荐工作流：

1. 先用顶级模型（如 GPT-4o）跑一次，记录各 benchmark 的 pass rate 作为参考分
2. 日常监测只跑被测 provider，对比参考分
3. 发现异常时，再临时做一次 A/B 对比确认

---

## 常见问题

**Q: 参考分应该用哪个模型的成绩？**

用你认为"正常水准"的模型。如果你购买的是 GPT-4o 级别的服务，就用 GPT-4o 的成绩作为参考分。如果你购买的是 GPT-4o-mini 级别，就用 GPT-4o-mini 的成绩。

**Q: 小样本 starter 集的结果可靠吗？**

单次运行有随机性，建议：
- 多次运行取平均（`repeats: 3`）
- 结合 history 趋势判断，不要只看单次结果
- 多个 benchmark 同时触发才更可信

**Q: 触发了注水告警，一定是注水吗？**

不一定。可能的原因：
- 参考分设置过高
- 小样本随机波动
- 模型本身在该 benchmark 上表现不稳定
- 网络/超时导致部分请求失败

建议：扩大样本量、多次重跑、结合 probe case 结果综合判断。

**Q: 没有触发告警，一定没有注水吗？**

不一定。注水检测有盲区：
- 如果注水幅度小于 20%，不会触发
- 如果参考分本身设置偏低，阈值也会偏低
- 某些能力维度没有被 benchmark 覆盖

建议定期更新参考分，并覆盖更多能力维度。
