# 粗粒度参水监测指南

这个项目当前更适合做 **粗粒度、低成本、可重复** 的参水监测，目标不是学术 leaderboard，而是：

- 快速发现模型/Provider 是否明显变差
- 快速发现格式遵循、逻辑、长上下文、tool calling 是否塌陷
- 在不做复杂 A/B 的情况下，给出一个足够稳定的告警信号

## 推荐使用场景

如果你关心的是：

- “今天是不是明显参水了？”
- “最近是否开始严重偷懒/降智？”
- “接口虽然还能用，但实际能力是不是掉了很多？”

那么建议优先跑：

```bash
./scripts/run_eino_monitoring.sh
```

对应配置：

- `examples/eino-monitoring-minimal.json`

它比完整 starter benchmark 更轻、更快，更适合日常巡检。

## 监测内容

这个最小监测配置包含：

1. **格式与基础推理 probe**
   - `exact_json`
   - `logic_filter`
2. **长上下文检索**
   - `long_context_needle_small`
   - `ruler_retrieval` starter 子集
3. **通用知识 / 科学问答**
   - `commonsenseqa`
   - `gpqa`
4. **WebQA / grounded QA**
   - `webqa`
5. **中文轻量异常检测**
   - `cn_brainteaser`
6. **tool calling**
   - `bfcl_style`

## 判读原则

这里不追求“和官方成绩高度一致”，只要满足下面几点，就已经足够拿来做参水监测：

- 同一模型多次运行结果不要大幅抖动
- 大多数 benchmark 不要掉到 `weak`
- probe 不要出现连续失败
- history 中不要持续出现 `high`

### 一个实用的粗判标准

如果出现下面任一情况，就值得怀疑 provider 明显参水：

- `exact_json` / `logic_filter` / `tool calling` 明显失败
- `webqa`、`commonsenseqa`、`ruler_retrieval` 同时掉到弱档
- `gpqa` 从可接受档掉到弱档，并且连续多次如此
- `history -group-by benchmark` 里多个 benchmark 同时转成高风险

## 为什么这里要“放宽”

因为这个项目解决的是 **严重参水监测**，不是论文复现。

所以这里有意采用：

- 小样本 starter 子集
- 粗粒度 baseline band
- “差不多即可”的预期区间

只要能稳定地区分：

- 正常
- 明显变弱
- 很可能严重参水

就已经达到主要目标。

## 推荐日常操作

### 1. 每天/每个时段跑一轮

```bash
./scripts/run_eino_monitoring.sh
```

### 2. 看 provider 总览

```bash
./scripts/history_summary.sh artifacts
```

### 3. 看 benchmark 维度退化位置

```bash
./scripts/history_summary.sh artifacts benchmark
```

### 4. 如果发现异常，再跑完整 starter 集

```bash
./scripts/run_eino_starter.sh
```

## 建议告警思路

适合接到 cron 或外部巡检里的简化逻辑：

- 单次 `run_eino_monitoring.sh` 返回高风险
- 或最近 3 次里有 2 次 `high`
- 或 benchmark history 里有 2 个以上 benchmark 落入明显弱档

则发告警。

## 不建议过度解读的地方

下面这些情况，不建议直接判定为参水：

- 单个 benchmark 小幅波动
- GPQA 这种较难集合偶发掉分
- 单次样本命中偏差

建议看：

- 是否连续出现
- 是否多个 benchmark 一起变差
- 是否基础 probe 也同时出问题
