# 安全与密钥处理

## 原则

- 不把真实 token 写进源码、配置模板、README、工单、commit message
- 运行时通过环境变量或受保护的 `.env.local` 注入
- 报告文件默认做敏感信息脱敏
- `artifacts/` 与 `.omx/` 不纳入版本控制

## 推荐做法

1. 从 `.env.example` 复制一份本地私有文件，例如 `.env.local`
2. 仅在本机或受控 CI secret store 中保存真实密钥
3. 使用脚本前先 `source ./.env.local`
4. 对外分享报告前，优先使用本工具生成的脱敏版报告

## 自动审计

运行：

```bash
./scripts/audit_secrets.sh
```

它会扫描仓库中的常见 token 形态，例如：

- `ms-...`
- `sk-...`
- `Bearer ...`

默认排除：

- `.git/`
- `.omx/`
- `artifacts/`
- 明确标注为假数据的测试夹具

## 说明

如果你曾经在终端、对话、工单或日志里暴露过真实 token，**仅仅修改仓库文件并不等于安全**。
仍然建议：

- 立即轮换 token
- 检查 CI 日志
- 检查 shell history
- 检查共享文档/聊天记录
