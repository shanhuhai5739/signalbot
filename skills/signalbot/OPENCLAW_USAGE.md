# SignalBot OpenClaw 使用指南

本文档面向使用者，说明如何在 OpenClaw 中安装和使用 signalbot skill。

---

## 前置条件

确保 `signalbot` 二进制已编译并在 PATH 中：

```bash
# 在 signalbot 项目目录下编译
cd /path/to/signalbot
go build -o signalbot .

# 移动到 PATH（任选其一）
cp signalbot /usr/local/bin/signalbot
# 或
export PATH="$PATH:/path/to/signalbot"
```

验证：
```bash
signalbot --help
```

---

## 安装 Skill

### 方式一：全局安装（推荐，所有 agent 共享）

```bash
mkdir -p ~/.openclaw/skills/signalbot
cp /path/to/signalbot/skills/signalbot/SKILL.md ~/.openclaw/skills/signalbot/
```

### 方式二：Workspace 安装（仅当前 agent）

将 `skills/signalbot/` 目录直接放入 OpenClaw workspace 的 `/skills/` 下：

```bash
cp -r /path/to/signalbot/skills/signalbot /your/openclaw/workspace/skills/
```

### 验证加载

重启 OpenClaw 或新建 session 后，执行：

```bash
openclaw skills list
```

输出中应包含 `signalbot 📊`。

---

## 直接对话调用

安装后，直接在 OpenClaw 聊天窗口中用自然语言提问即可，agent 会自动识别并调用 signalbot：

```
BTC 现在行情怎么样？
```

```
分析一下黄金日线，给出操作建议
```

```
同时分析 BTC 4小时和黄金日线，生成一条行情分析推文
```

```
BTC 和 XAUUSD 当前谁更强势？
```

---

## Cron Job 配置

### 每日早 8 点 BTC 行情日报

```bash
openclaw cron add --schedule "0 8 * * *" \
  --prompt "运行 signalbot analyze --asset BTC --timeframe 4h 获取行情数据，然后以专业量化交易员视角用中文生成一条行情分析推文，包含：当前趋势、关键支撑阻力位、技术指标信号摘要、操作建议（BUY/SELL/HOLD）和风险提示。推文长度控制在 250 字以内。" \
  --name "btc-daily-analysis"
```

### 每日早 8 点 BTC + 黄金双标的报告

```bash
openclaw cron add --schedule "0 8 * * *" \
  --prompt "依次运行以下两条命令：
1. signalbot analyze --asset BTC --timeframe 4h
2. signalbot analyze --asset XAUUSD --timeframe 1d

综合两份 JSON 数据，用中文生成一份「早间行情简报」，格式如下：
- 📊 市场概览（1句话）
- ₿ BTC 分析（趋势 + 关键价位 + 操作建议）
- 🥇 黄金分析（趋势 + 关键价位 + 操作建议）
- ⚠️ 风险提示（1句话）" \
  --name "morning-market-briefing"
```

### 每 4 小时行情监控

```bash
openclaw cron add --schedule "0 */4 * * *" \
  --prompt "运行 signalbot analyze --asset BTC --timeframe 4h，如果 analysis.signal 为 BUY 且 analysis.confidence >= 60，或 analysis.signal 为 SELL 且 analysis.confidence >= 60，则生成一条行情预警推文；否则回复「当前无明显信号，无需推送」。" \
  --name "btc-4h-alert"
```

### 查看和管理 Cron Job

```bash
# 查看所有 job
openclaw cron list

# 查看运行历史
openclaw cron history btc-daily-analysis

# 手动触发一次
openclaw cron run btc-daily-analysis

# 删除
openclaw cron remove btc-daily-analysis
```

---

## 常用 Prompt 模板

### 标准行情分析推文

```
分析 BTC 4小时行情，生成一条专业的中文行情分析推文，要求：
1. 包含 RSI、MACD、布林带、EMA 的信号解读
2. 列出最近 2–3 个支撑位和阻力位
3. 给出明确的操作建议（BUY/SELL/HOLD）和置信度
4. 末尾加上相关话题标签（#BTC #行情分析 等）
5. 总长度 250 字以内
```

### 突破预警

```
分析 BTC 4小时行情，判断是否存在以下信号：
- 价格突破布林带上轨（position = above_upper）且成交量放大（ratio > 1.5）
- MACD 金叉（cross = golden_cross）
- EMA 多头排列（alignment 含 bullish）

如果同时满足 2 个以上条件，生成一条突破预警推文；否则不输出任何内容。
```

### 多标的趋势对比

```
分别获取 BTC（4h）和 XAUUSD（1d）的行情数据，对比分析：
1. 哪个标的趋势更强？
2. 两者相关性如何（同涨同跌 or 背离）？
3. 基于当前指标，哪个更值得关注？
用简洁的中文表格或要点列表呈现。
```

### 周度行情总结

```
获取 BTC 日线行情（--timeframe 1d），结合指标数据生成本周行情总结：
- 本周价格区间
- 主要趋势和关键转折
- 下周重要支撑阻力位
- 综合操作建议
```

### 风险评估报告

```
获取 BTC 4小时行情，重点分析风险指标：
1. ATR 波动率机制（atr.regime）和具体数值
2. 布林带宽度（bollinger.width）是否异常收窄或扩张
3. RSI 是否处于超买/超卖极端区域
4. 成交量是否异常（ratio 极高或极低）
基于以上数据给出当前市场风险等级（低/中/高）和风险提示。
```

---

## 环境变量（可选）

如需在 OpenClaw 环境中配置，在 `~/.openclaw/openclaw.json` 中添加：

```json5
{
  "skills": {
    "entries": {
      "signalbot": {
        "enabled": true,
        "env": {
          "BINANCE_BASE_URL": "https://your-proxy.com",
          "HTTP_TIMEOUT_SEC": "20",
          "DEFAULT_LIMIT": "300"
        }
      }
    }
  }
}
```

---

## 故障排查

| 问题 | 原因 | 解决方案 |
|------|------|----------|
| `zsh: command not found: signalbot` | 二进制不在 PATH | 编译后 `cp signalbot /usr/local/bin/` |
| Skill 未被 OpenClaw 识别 | 未重启 session | 执行 `/new` 新建 session 或 `openclaw gateway restart` |
| 请求超时 | Binance 网络不通 | 设置 `BINANCE_BASE_URL` 为代理地址 |
| XAUUSD 数据异常 | 短周期流动性不足 | 改用 `--timeframe 1d` |
| 数据不足（< 50 根） | 标的名称错误 | 检查 `--asset` 参数是否正确 |
