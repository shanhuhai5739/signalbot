---
name: signalbot
description: 量化行情分析工具，对 BTC、黄金（XAUUSD）等标的计算 RSI、MACD、布林带、EMA、ATR、成交量六类技术指标，输出结构化 JSON 行情报告，可据此生成行情分析推文或做出操作建议。
metadata: {"openclaw": {"emoji": "📊", "requires": {"bins": ["signalbot"]}, "install": [{"id": "download-mac-arm64", "kind": "download", "url": "https://github.com/shanhuhai5739/signalbot/releases/latest/download/signalbot-darwin-arm64", "os": ["darwin"]}, {"id": "download-linux", "kind": "download", "url": "https://github.com/shanhuhai5739/signalbot/releases/latest/download/signalbot-linux-amd64", "os": ["linux"]}]}}
---

# Signalbot 量化行情分析技能

## 何时使用

当用户提出以下类型的请求时，主动调用此技能：

- 询问 BTC、比特币、黄金、XAUUSD 等标的的行情、走势、涨跌
- 需要技术指标分析（RSI、MACD、布林带、均线、ATR、成交量）
- 需要生成行情分析推文、日报、周报
- 需要判断当前趋势方向、支撑阻力位、操作建议（BUY/SELL/HOLD）
- 用户要求"分析一下今天行情"、"BTC 现在多空如何"、"黄金值得买吗"等

## 调用方式

使用 `exec` 工具（或 `bash`）运行以下命令，将 stdout 捕获为 JSON：

```bash
signalbot analyze --asset <标的> --timeframe <周期>
```

**注意**：命令会在 stderr 输出两行进度提示（可忽略），JSON 报告输出到 stdout。

### --asset 合法值

| 值 | 含义 |
|---|---|
| `BTC` | 比特币 (BTCUSDT) |
| `XAUUSD` | 黄金 (XAUUSDT) |
| `ETH` | 以太坊 (ETHUSDT) |
| `SOL` | Solana (SOLUSDT) |
| `BNB` | 币安币 (BNBUSDT) |

也可直接填写 Binance 符号，如 `BTCUSDT`。

### --timeframe 合法值

| 值 | 适用场景 |
|---|---|
| `1h` | 短线日内分析 |
| `4h` | 中线波段分析（**推荐默认**） |
| `1d` | 长线趋势分析 |
| `15m` | 超短线 |

### 可选参数

- `--limit <数量>`：拉取 K 线数量，默认 200，最多 1500
- `--output <文件路径>`：将 JSON 写入文件而非 stdout

## JSON 输出字段解读

拿到 JSON 后，按以下规则解读各字段，再生成自然语言分析：

### price（最新价格快照）
- `current`：当前收盘价
- `change_pct`：较上一根 K 线的涨跌幅（%）

### indicators.rsi
- `value`：RSI 值（0–100）
- `signal`：
  - `overbought`（≥70）→ 超买，注意回调
  - `bullish`（55–69）→ 强势区间
  - `neutral`（46–54）→ 中性观望
  - `bearish`（31–45）→ 弱势区间
  - `oversold`（≤30）→ 超卖，关注反弹

### indicators.macd
- `histogram > 0` → 多头动能增强
- `histogram < 0` → 空头动能增强
- `cross: "golden_cross"` → 金叉（强看涨信号）
- `cross: "death_cross"` → 死叉（强看跌信号）

### indicators.bollinger
- `percent_b`：价格在布林带中的位置（0=下轨，0.5=中轨，1=上轨）
- `position: "upper_zone"` → 强势突破区，`"lower_zone"` → 弱势区
- `width`：带宽越窄代表价格越压缩，往往是行情突破前兆

### indicators.ema
- `alignment: "strongly_bullish"` → EMA9>21>50>200，四线多头完全排列
- `alignment: "bullish"` → EMA9>21>50，短中期多头
- `alignment: "bearish"` / `"strongly_bearish"` → 空头排列

### indicators.atr
- `pct`：ATR 占当前价格的百分比
- `regime: "low_volatility"` → 市场压缩，突破在即；`"high_volatility"` → 波动剧烈，注意止损

### indicators.volume
- `ratio`：当前成交量 / 20日均量
- `signal: "high_volume"` 或 `"above_average"` → 成交量放大确认行情
- `signal: "low_volume"` → 量能不足，信号可信度下降

### analysis（综合研判）
- `trend`：`bullish` / `neutral` / `bearish`
- `strength`：`strong` / `moderate` / `weak`
- `signal`：`BUY` / `SELL` / `HOLD`
- `confidence`：0–100 置信度，基于多指标共振程度
- `score`：原始评分（-8 到 +8），负数越大越空头
- `key_levels.support`：近期支撑价位列表（从近到远）
- `key_levels.resistance`：近期阻力价位列表（从近到远）

## 推文生成指南

拿到 JSON 数据后，按以下结构生成中文行情分析推文（约 200–280 字）：

```
📊 $<标的> | <时间周期>行情分析

📈/📉 趋势：<多头/空头/震荡>（<strength>）

关键指标：
• RSI(<value>)：<信号描述>
• MACD：<histogram方向 + 是否金叉/死叉>
• EMA排列：<描述>
• 布林带：<价格位置描述>

🎯 关键价位：
支撑：$<support[0]> / $<support[1]>
阻力：$<resistance[0]> / $<resistance[1]>

⚡️ ATR波动：<regime>（<pct>%）
📊 成交量：<signal描述>

🔖 操作建议：<BUY/SELL/HOLD>（置信度 <confidence>%）
<简要理由，1–2句>

#BTC #Bitcoin #行情分析 #量化交易
```

## 多标的分析示例

如需同时分析 BTC 和黄金，依次运行两条命令，再合并分析：

```bash
signalbot analyze --asset BTC --timeframe 4h
signalbot analyze --asset XAUUSD --timeframe 1d
```

## 常见问题处理

- **命令不存在**：提示用户先 `go build -o signalbot .` 编译，或将二进制加入 PATH
- **数据不足**：XAUUSD 流动性较低，建议使用 `1d` 周期；BTC 各周期均可
- **网络超时**：建议设置 `BINANCE_BASE_URL` 环境变量切换为代理地址
