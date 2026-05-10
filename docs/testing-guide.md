# 测试指南

## 1. 测试体系概览

```
测试层次：
├── 单元测试          make test（Go test，无需网络）
├── 冒烟测试          make smoke（编译 + 单元测试 + 自检）
├── 快速监测          make eino-monitoring（14 case，~30s）
├── Starter 评测      make eino-starter（含所有 starter dataset）
└── 全量评测          ./scripts/run_deepseek_full_eval.sh（68 case，~10min）
```

---

## 2. 单元测试

```bash
make test
# 或
go test ./...
```

覆盖范围：
- `internal/config/` — 配置解析、环境变量解析、校验
- `internal/provider/` — API 客户端（mock HTTP server）
- `internal/runner/` — 脱敏逻辑
- `internal/report/` — 报告渲染
- `internal/suite/` — Case 构建
- `cmd/provider-probe/` — CLI 参数解析

无需网络连接，无需 API key。

---

## 3. 集成测试（需要 API）

### 3.1 前置条件

```bash
# 确保 kiro-gateway 运行中
ps aux | grep "main.py --port 58575"

# 设置环境变量
export OPENAI_API_KEY=$(grep "PROXY_API_KEY" /root/kiro-gateway/.env | cut -d'"' -f2)
export BASE_URL="http://localhost:58575/v1"
export MODEL="deepseek-3.2"
```

### 3.2 快速监测（推荐日常使用）

```bash
./scripts/run_eino_monitoring.sh
```

| 项目 | 值 |
| --- | --- |
| Case 数量 | 14 |
| 预计耗时 | 30-60 秒 |
| 配置文件 | `examples/eino-monitoring-minimal.json` |
| 覆盖能力 | JSON格式、逻辑推理、长上下文、常识、问答、检索、中文、Tool Calling |

### 3.3 Starter Benchmark

```bash
./scripts/run_eino_starter.sh
```

| 项目 | 值 |
| --- | --- |
| Case 数量 | ~30 |
| 预计耗时 | 2-5 分钟 |
| 配置文件 | `examples/eino-benchmark-starter.json` |
| 覆盖能力 | 所有 starter 数据集 |

### 3.4 全量评测

```bash
./scripts/run_deepseek_full_eval.sh
```

| 项目 | 值 |
| --- | --- |
| Case 数量 | 68 |
| 预计耗时 | 8-15 分钟 |
| 配置文件 | `examples/deepseek-full-eval.json` |
| 覆盖能力 | 全部内置 case + 全部 dataset（含 real 高难度） |

---

## 4. 解读报告

### 4.1 关键指标

| 指标 | 含义 | 正常范围 |
| --- | --- | --- |
| Score | 综合得分（0-100） | ≥ 75 |
| Pass Rate | 通过率 | ≥ 70% |
| Suspicion | 注水嫌疑等级 | low |
| Error Runs | 请求失败数 | 0 |
| Returned Models | 响应中的模型名 | 应与请求一致 |

### 4.2 Suspicion 等级

| 等级 | 触发条件 | 含义 |
| --- | --- | --- |
| low | 无警告 | 正常 |
| medium | 1+ 警告或 1 error | 需关注 |
| high | 3+ 警告或 watermark suspected | 高度可疑 |

### 4.3 Watermark Suspected

当某个 benchmark 的 pass_rate < reference_score × 80% 时触发。

**注意**：不一定是真正注水，可能原因：
- evaluator 过严（如 webqa 的 exact_match）
- 模型 API 限制（如 deepseek reasoner 不支持 tool_choice）
- 题目本身极难（如 GPQA Diamond）

需要结合失败 case 的具体 `expected` vs `actual` 判断。

### 4.4 Starter Baseline Band

| Band | 含义 | 行动 |
| --- | --- | --- |
| strong | 通过率高于阈值 | 正常 |
| acceptable | 通过率在中间 | 观察 |
| weak | 通过率低于阈值 | 需排查 |

---

## 5. 添加新 Case

### 5.1 添加内置 Probe Case

编辑 `internal/suite/default.go`，在 `cases` map 中添加：

```go
"my_new_case": {
    Name:     "my_new_case",
    Category: "reasoning",
    Messages: []provider.Message{
        {Role: "system", Content: "You are a helpful assistant."},
        {Role: "user", Content: "What is 2+2? Answer with just the number."},
    },
    Evaluate: func(resp provider.Response) EvalResult {
        return EvalResult{
            Passed:   strings.TrimSpace(resp.Content) == "4",
            Score:    boolScore(strings.TrimSpace(resp.Content) == "4"),
            Expected: "4",
            Actual:   resp.Content,
        }
    },
},
```

然后在配置文件中启用：

```json
{"name": "my_new_case", "enabled": true}
```

### 5.2 添加 JSONL 数据集

创建 `benchmarks/starter/my-dataset.jsonl`：

```jsonl
{"id": "q1", "benchmark": "my_bench", "split": "starter", "category": "reasoning", "prompt": "What is the capital of France?", "expected": "Paris", "acceptable_answers": ["Paris"], "evaluator": "exact_match"}
{"id": "q2", "benchmark": "my_bench", "split": "starter", "category": "reasoning", "prompt": "2+2=?", "choices": ["A. 3", "B. 4", "C. 5"], "expected": "B", "acceptable_answers": ["B"], "evaluator": "multiple_choice"}
```

在配置文件中引用：

```json
{
  "name": "my-dataset",
  "enabled": true,
  "dataset": {
    "path": "benchmarks/starter/my-dataset.jsonl",
    "name": "my_bench",
    "split": "starter"
  }
}
```

### 5.3 Evaluator 类型

| Evaluator | 匹配方式 | 适用场景 |
| --- | --- | --- |
| `exact_match` | 精确匹配（trim 后） | 单词/短语答案 |
| `multiple_choice` | 提取选项字母 A/B/C/D | 选择题 |
| `tool_call` | 验证 tool call name + arguments | Function Calling |
| `contains` | 包含匹配 | 长文本中的关键信息 |

---

## 6. 配置文件结构

```json
{
  "providers": [{
    "name": "provider-name",
    "type": "eino_openai",
    "base_url_env": "BASE_URL",
    "model_env": "MODEL",
    "api_key_env": "OPENAI_API_KEY",
    "timeout_seconds": 120,
    "headers": {"X-Debug-Client": "provider-probe"}
  }],
  "run": {
    "repeats": 1,
    "temperature": 0,
    "capture_headers": true,
    "reference_scores": {
      "commonsenseqa": 0.90,
      "gpqa": 0.65
    }
  },
  "suite": {
    "cases": [
      {"name": "exact_json", "enabled": true},
      {"name": "my-dataset", "enabled": true, "dataset": {"path": "...", "name": "...", "split": "..."}}
    ]
  }
}
```

---

## 7. 历史趋势与对比

### 7.1 查看历史

```bash
# 按 provider 汇总
make history

# 按 benchmark 汇总
make benchmark-history

# 生成 Markdown/HTML 历史报告
make history-files
make benchmark-history-files
```

### 7.2 对比两次运行

```bash
./provider-probe compare -left artifacts/run-a.json -right artifacts/run-b.json
```

输出每个 provider 的 score/pass_rate/error 变化和 per-case delta。

---

## 8. CI 集成

```bash
# 在 CI 中运行，suspicion=high 时退出码 3
./provider-probe -config examples/eino-monitoring-minimal.json -fail-on high

# 检查退出码
if [ $? -eq 3 ]; then
  echo "⚠️ Provider quality degraded!"
fi
```

---

## 9. 故障排查

| 现象 | 可能原因 | 排查方法 |
| --- | --- | --- |
| 全部 timeout | Gateway 未启动 | `curl http://localhost:58575/v1/models` |
| HTTP 401 | API key 错误 | 检查 `OPENAI_API_KEY` 环境变量 |
| HTTP 400 | 模型不支持某参数 | 查看 error 字段详情 |
| returned_model 不一致 | 中转替换了模型 | 检查报告中 `distinct_returned_models` |
| pass_rate 突然下降 | 模型退化或中转问题 | 对比历史报告 `make history` |
| webqa 全部 0% | evaluator 过严 | 正常现象，模型回答正确但格式不符 |
