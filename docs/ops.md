# 运维建议

## 环境变量

推荐使用环境变量注入：

- `MODELSCOPE_API_KEY`
- `OPENAI_API_KEY`

不要把真实 key 写入配置文件并提交。

## Makefile

可直接使用：

```bash
make test
make list-cases
make history
make history-files
```

如果已经设置好 `MODELSCOPE_API_KEY`：

```bash
make modelscope
```

如果同时设置了 `MODELSCOPE_API_KEY` 与 `OPENAI_API_KEY`：

```bash
make ab
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
- `high` 次数较多
- `Total error runs` 持续上升

则说明 provider 稳定性存在明显问题。

## 安全建议

- 定期轮换 API key
- 为 key 设置最小权限
- 不要把包含 token 的命令写入公共日志
