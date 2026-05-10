# provider-probe

LLM 中转服务质量评测工具。检测 API 中转商是否存在模型替换（注水）、能力退化、响应异常等问题。

基于 [Eino](https://github.com/cloudwego/eino) 框架，兼容所有 OpenAI API 格式的 provider。

## 它能做什么

- 检测中转服务是否偷换模型（通过 `returned_model` 字段）
- 评估模型核心能力：推理、常识、长上下文、Tool Calling
- 基于 reference score 自动判定注水嫌疑
- 输出 JSON / Markdown / HTML 报告
- 支持历史趋势对比和多 provider A/B 测试

## 快速开始

### 1. 设置环境变量

```bash
export OPENAI_API_KEY="<your-api-key>"
export BASE_URL="<provider-base-url>"   # 如 http://localhost:58575/v1
export MODEL="<model-name>"             # 如 claude-sonnet-4.5
```

### 2. 运行评测

```bash
# 快速监测（14 case，~30s）
./scripts/run_eino_monitoring.sh

# 全量评测（68 case，~10min）
./scripts/run_deepseek_full_eval.sh

# Starter benchmark（含所有 starter dataset）
./scripts/run_eino_starter.sh
```

### 3. 查看报告

运行后在 `artifacts/` 目录生成：
- `*.json` — 结构化数据
- `*.md` — Markdown 可读报告
- `*.html` — 浏览器可视化报告

## 项目结构

```
cmd/provider-probe/       CLI 入口
internal/
  config/                 配置解析（环境变量 + JSON）
  provider/               Eino OpenAI 客户端（API 通信层）
  runner/                 执行引擎（遍历 case、收集结果、脱敏）
  suite/                  内置 probe case 定义
  dataset/                JSONL 数据集加载器
  report/                 报告渲染（JSON/Markdown/HTML）
benchmarks/
  starter/                内置评测数据集（JSONL）
  custom/                 自定义数据集示例
  mappings/               外部 benchmark 转换映射
examples/                 运行配置模板
scripts/                  运行/对比/历史脚本
docs/                     详细文档
artifacts/                评测报告输出目录
```

## 常用命令

| 命令 | 说明 |
| --- | --- |
| `make test` | 运行单元测试 |
| `make eino-monitoring` | 快速监测 |
| `make eino-starter` | Starter benchmark |
| `make history` | 查看历史趋势 |
| `make audit-secrets` | 密钥泄漏审计 |
| `./provider-probe -list-cases` | 列出内置 probe case |

## 配置方式

### 环境变量模式（推荐）

```bash
export OPENAI_API_KEY="sk-xxx"
export BASE_URL="http://localhost:58575/v1"
export MODEL="deepseek-3.2"
./scripts/run_eino_monitoring.sh
```

### 配置文件模式

```bash
./provider-probe -config examples/deepseek-full-eval.json
```

### CLI 单次模式

```bash
./provider-probe \
  -base-url "http://localhost:58575/v1" \
  -model "deepseek-3.2" \
  -api-key-env "OPENAI_API_KEY" \
  -cases "exact_json,logic_filter"
```

## 内置 Probe Case

| Case | 类别 | 检测目标 |
| --- | --- | --- |
| `exact_json` | 格式 | JSON 输出精确性 |
| `exact_line` | 格式 | 单行精确输出 |
| `logic_filter` | 推理 | 条件过滤逻辑 |
| `chinese_compact` | 中文 | 中文紧凑输出 |
| `nested_json_schema` | 格式 | 嵌套 JSON Schema |
| `go_snippet_output` | 代码 | Go 代码片段 |
| `tool_call_echo` | 工具 | Tool Calling |
| `long_context_needle_*` | 长上下文 | Needle-in-Haystack |

## Benchmark 数据集

| 数据集 | 样本数 | 评估能力 |
| --- | --- | --- |
| commonsenseqa | 3 | 常识推理 |
| mmlu-pro (starter) | 3 | 多学科知识 |
| mmlu-pro (real) | 20 | 多学科知识（高难度） |
| gpqa (starter) | 3 | 研究生级问答 |
| gpqa-diamond (real) | 12 | 研究生级问答（高难度） |
| logiqa | 3 | 逻辑推理 |
| bbh-logical-deduction | 2 | 逻辑演绎 |
| bbh-tracking-objects | 2 | 对象追踪 |
| webqa | 3 | 基于文本的问答 |
| ruler-retrieval | 3 | 长上下文检索 |
| brainteaser-zh | 3 | 中文脑筋急转弯 |
| bfcl-style-tool | 2 | Function Calling |

## 注水检测原理

1. **returned_model 检测**：请求 model A，响应返回 model B → 模型替换
2. **Reference Score 阈值**：pass_rate < 参考分 × 80% → 注水嫌疑
3. **Starter Baseline Band**：将通过率分为 weak / acceptable / strong

详见 `docs/watermark-detection.md`。

## 文档索引

| 文档 | 内容 |
| --- | --- |
| [交接文档](docs/kiro-relay-architecture.md) | Kiro 中转架构、通信协议、评测结论 |
| [测试指南](docs/testing-guide.md) | 如何运行测试、添加 case、解读报告 |
| [注水检测](docs/watermark-detection.md) | 检测原理与阈值说明 |
| [自定义数据集](docs/custom-dataset-standard.md) | JSONL 数据集接入规范 |
| [Benchmark 转换](docs/benchmark-conversion.md) | 外部数据集转换方法 |
| [监测用法](docs/monitoring.md) | 粗粒度定期监测配置 |
| [安全规范](docs/security.md) | 密钥管理与脱敏 |

## 安全

- 所有报告输出自动脱敏（API key、Bearer token）
- 不要将密钥写入配置文件或提交到仓库
- 使用环境变量注入所有敏感信息
