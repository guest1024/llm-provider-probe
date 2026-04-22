# 运维建议

## 环境变量

推荐使用环境变量注入：

- `OPENAI_API_KEY`
- `BASE_URL`
- `MODEL`
- 可选：`MODELSCOPE_API_KEY`

不要把真实 key 写入配置文件并提交。

## 推荐启动方式

```bash
./scripts/run_eino_starter.sh
```

## Makefile

```bash
make test
make list-cases
make eino-starter
make eino-monitoring
make history
make benchmark-history
make history-files
make benchmark-history-files
make audit-secrets
```

## 定时任务

参考：`ops/cron.example`

建议至少覆盖：

- 凌晨
- 上午
- 晚高峰
- 周末高峰

## 结果判读

如果 history 中出现：

- `Score avg/min/max` 波动很大
- `Latest pass rate` / `Avg pass rate` 明显下滑
- `high` 次数较多
- `Total error runs` 持续上升
- 某个 benchmark 持续低于 starter baseline band

则说明 provider 稳定性或能力存在明显问题。

如果你的主要目标是监测**严重参水**，建议优先跑：

```bash
./scripts/run_eino_monitoring.sh
```

这套配置更轻量，也更适合做 cron / 日常巡检。

## 安全建议

- 定期轮换 API key
- 为 key 设置最小权限
- 不要把包含 token 的命令写入公共日志
- 不要把真实 token 粘贴进 issue、README、commit message、CI 明文变量输出

## 数据集转换

如果要接公开 benchmark 原始样本，先用 `scripts/convert_dataset.py` 转成项目 JSONL，再加入配置。
