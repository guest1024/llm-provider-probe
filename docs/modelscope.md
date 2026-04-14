# ModelScope 实战

## 配置

已内置示例：`examples/modelscope-qwen.json`

目标模型：

- Base URL: `https://api-inference.modelscope.cn/v1`
- Model: `Qwen/Qwen3.5-397B-A17B`

## 运行

```bash
export MODELSCOPE_API_KEY=你的key
./scripts/run_modelscope.sh --repeat 2
```

## 时段批跑

```bash
export MODELSCOPE_API_KEY=你的key
./scripts/timeslot_batch.sh --config examples/modelscope-qwen.json --repeat 1
```

默认时段标签：

- 凌晨
- 上午
- 晚高峰
- 周末高峰

> 这个脚本不会自动等待到这些真实时间点；它的作用是帮助你给不同批次打标签。真正定时执行建议配合 cron / CI scheduler。

## 和官方 baseline 做 A/B

```bash
export MODELSCOPE_API_KEY=你的key
export OPENAI_API_KEY=你的key
./scripts/run_ab.sh examples/modelscope-vs-openai-template.json ab-check
```

## 建议观察项

1. `returned_model` 是否稳定
2. 是否出现 `provider returned 200 but no choices payload`
3. `logic_filter` 是否稳定
4. `long_context_needle_*` 是否随时段退化
5. Markdown 报告中的 warnings 是否持续出现
6. `history_summary.sh` 中 avg/min/max 是否波动过大

## 生产建议

- 为 API key 设置最小权限并定期轮换
- 不要把 key 写进配置文件提交到仓库
- 用环境变量注入，例如 `MODELSCOPE_API_KEY`
