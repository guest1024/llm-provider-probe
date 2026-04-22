# 判定方法

## 总体思路

本工具不再只依赖单次 A/B，而是结合两类信号：

1. **Probe signal**：格式、指令遵循、逻辑、长上下文、tool calling 稳定性
2. **Benchmark signal**：通用 QA、逻辑题、WebQA、脑筋急转弯、工具调用数据集表现

## 当前 probe case

- `exact_json`
- `exact_line`
- `logic_filter`
- `chinese_compact`
- `nested_json_schema`
- `response_format_json_schema`
- `go_snippet_output`
- `tool_call_echo`
- `long_context_needle_small`
- `long_context_needle_medium`
- `long_context_needle_large`

## 当前 starter benchmark coverage

- `commonsenseqa-starter`
- `mmlu-pro-starter`
- `gpqa-starter`
- `logiqa-starter`
- `bbh-logical-deduction-starter`
- `bbh-tracking-objects-starter`
- `webqa-starter`
- `ruler-retrieval-starter`
- `brainteaser-zh-starter`
- `bfcl-style-tool-starter`

## 当前 evaluator

- `exact_match`
- `regex_match`
- `multiple_choice`
- `tool_call`

## 为什么不是只做 A/B

只做 A/B 有几个问题：

- 很难知道差异来自哪种能力
- 不能稳定复现具体失败样本
- 对内部专用数据集不友好
- 对 WebQA / 逻辑 / tool calling 这类垂直能力不够细

所以更合理的方式是：

- 用 probe 看稳定性和协议行为
- 用 benchmark 看能力矩阵
- 用 history 看时段波动和漂移
- 用 benchmark history 看具体测试集的退化位置，而不是只看 provider 总分

## 为什么这里不用太严

这个项目当前主要目标是：

- 发现明显参水
- 发现明显退化
- 发现“还能回答但实际能力已经塌了”的情况

所以这里采用的是：

- starter 子集
- 粗粒度 baseline band
- 宽松但稳定的阈值

重点不是复现学术分数，而是把**严重异常**尽快捞出来。

## Suspicion 解释

### low
- 错误较少
- 关键 probe 稳定
- starter benchmark 没有明显塌陷

### medium
- 有错误返回
- 某些能力面不稳定
- benchmark/probe 某个维度明显变弱

### high
- 多次错误返回
- 多个能力面持续失稳
- 同批对比显著落后
- benchmark pass rate 或总分明显偏低

## 最佳实践

1. 先跑 starter benchmark + probe
2. 再加内部回归集
3. 再做多时段重跑
4. 最后再看是否需要对外 baseline A/B
