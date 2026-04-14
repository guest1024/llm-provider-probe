# 判定方法

## 目标

本工具不是直接断言“这个 provider 一定换了模型”，而是输出一条可追溯证据链：

1. 返回模型标识是否稳定
2. 格式约束是否能稳定满足
3. 指令跟随/推理是否稳定
4. 长上下文检索是否早衰
5. 是否存在明显的网关异常（如 200 但 `choices: null`）

## 当前内置 case

- `exact_json`：严格 JSON 输出
- `exact_line`：严格单行输出
- `logic_filter`：多条件逻辑过滤
- `chinese_compact`：中文严格单行输出
- `nested_json_schema`：嵌套 JSON 结构遵循
- `response_format_json_schema`：JSON Schema 严格结构输出
- `go_snippet_output`：Go 代码结果推断
- `tool_call_echo`：工具/函数调用能力
- `long_context_needle_small`：小规模长上下文检索
- `long_context_needle_medium`：中规模长上下文检索
- `long_context_needle_large`：大规模长上下文检索（默认关闭）

## Suspicion 解释

### low

- 没有明显错误返回
- 结构化输出与逻辑题稳定
- 长上下文检索大体正常

### medium

- 存在单次或少量错误返回
- 某一类能力偶发失稳
- 需要扩大样本与时段继续观察

### high

满足以下任一倾向：

- 多次错误返回
- 多个核心能力持续失稳
- 同批 A/B 相比显著落后
- 出现较频繁的 `choices: null` / usage 全 0 / 异常空响应

## 特殊异常：HTTP 200 但无 choices

如果 provider 返回 HTTP 200，但响应体中：

- `choices` 为空或 `null`
- `usage` 为 0
- 无法产出可评估内容

则应视为**有效结果缺失**，而不是“模型答错了”。
这类问题更偏向：

- 网关/路由异常
- 推理任务未完成却被包装成成功
- 下游服务退化或瞬时故障

## 最佳实践

1. 先跑默认矩阵 2~3 次
2. 再用 `timeslot_batch.sh` 在不同时段重跑
3. 最好增加一个官方直连 baseline
4. 看 trend，不看单次偶发结果

## 建议的 baseline

如果你怀疑某个 provider 在“参水”，最有力的方法仍然是同题 A/B：

- target provider
- 官方直连 baseline

仓库里已经提供模板：

- `examples/openai-baseline-template.json`
- `examples/modelscope-vs-openai-template.json`
