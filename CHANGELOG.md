# CHANGELOG

## 2026-05-10

### 修复 Evaluator 误报

**问题**：之前的评测中 webqa 和 bfcl_style 报告 `watermark_suspected`，实际是 evaluator 配置问题而非真正注水。

**修复内容**：

1. **webqa evaluator 改为 contains 匹配**
   - 之前：`exact_match` 要求精确匹配单词，模型输出完整句子被判 0 分
   - 之后：`contains` 只要回答中包含正确答案即判 pass
   - 效果：webqa 0% → 100%

2. **新增 stripHallucinationMarkers 预处理**
   - DeepSeek 模型有时会幻觉输出 `<web_search>` 或 `<｜DSML｜function_calls` 标记
   - 在 exactMatchEvaluator、multipleChoiceEvaluator、containsEvaluator 中统一 strip
   - 效果：减少因幻觉标记导致的误判

3. **新增 containsEvaluator**
   - 支持 `"evaluator": "contains"` 类型
   - 适用于答案正确但模型倾向输出完整句子的场景

**验证结果**（deepseek-3.2 via Kiro Gateway）：
- Score: 85.7 → 92.9
- Pass Rate: 78.6% → 92.9%
- webqa: 0% → 100%
- bfcl_style: 0% → 100%

### 新增全量评测配置

- `examples/deepseek-full-eval.json` — 68 case 全量评测
- `scripts/run_deepseek_full_eval.sh` — 一键运行脚本

### 文档重写

- `README.md` — 重写，结构清晰
- `docs/kiro-relay-architecture.md` — 交接文档（架构、协议、运维、评测结论）
- `docs/testing-guide.md` — 测试指南（运行方法、报告解读、添加 case）
