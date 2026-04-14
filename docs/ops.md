# 运维建议

## 环境变量

推荐使用环境变量注入：

- `MODELSCOPE_API_KEY`
- `OPENAI_API_KEY`

不要把真实 key 写入配置文件并提交。
建议使用仓库外或未纳入版本控制的 `.env.local` / secret manager 注入。

## 敏感信息保护

当前工具会在报告生成阶段自动对以下内容做脱敏：

- `Authorization` / `X-API-Key` / `Cookie` / `Set-Cookie` 等敏感响应头
- `Bearer ...` 形式的长 token
- `ms-...` / `sk-...` 这类常见供应商 key 形态
- 常见 `token=` / `api_key=` / `secret=` 形式的长值

这意味着：

- 报告与 history 文件默认更安全
- 但**源环境变量、shell 历史、外部日志系统**仍需你自己保护

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

示例已改成 `source ./.env.local` 的形式，避免把 token 直接写在 crontab 行里。

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
- 不要把真实 token 粘贴进 issue、README、commit message、CI 明文变量输出
