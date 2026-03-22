# signalbot

BTC、黄金等标的的量化行情分析 CLI 工具。

通过 **Binance 公共 Klines API**（无需 API Key）获取 K 线数据，计算六类技术指标，输出结构化 **JSON 报告**，可直接供 LLM（如 OpenClaw、GPT）读取并生成行情分析推文。

---

## 构建

```bash
# 需要 Go 1.21+
git clone <repo>
cd signalbot
go build -o signalbot .
```

---

## 快速开始

```bash
# 分析 BTC 4小时行情
./signalbot analyze --asset BTC --timeframe 4h

# 分析黄金日线行情，结果写入文件
./signalbot analyze --asset XAUUSD --timeframe 1d --output gold_report.json

# 拉取更多 K 线
./signalbot analyze --asset BTC --timeframe 1h --limit 500

# 管道给 LLM CLI（以 claude 为例）
./signalbot analyze --asset BTC --timeframe 4h | \
  claude "根据以下行情 JSON，以专业交易员视角撰写一条行情分析推文，包含关键价位和操作建议："
```

---

## CLI 参数

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--asset` | `-a` | **必填** | 标的资产：`BTC`、`XAUUSD`、`ETH`、`SOL` 等 |
| `--timeframe` | `-t` | `4h` | K 线周期：`1m` `5m` `15m` `30m` `1h` `2h` `4h` `6h` `8h` `12h` `1d` `3d` `1w` `1M` |
| `--limit` | `-l` | `200` | 获取 K 线数量（最多 1500，自动补足至 210 以保证 EMA200 收敛） |
| `--output` | `-o` | stdout | JSON 输出文件路径，不填则打印到 stdout |

---

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `BINANCE_BASE_URL` | `https://api.binance.com` | Binance API 地址，可替换为代理 |
| `HTTP_TIMEOUT_SEC` | `15` | HTTP 请求超时（秒） |
| `DEFAULT_LIMIT` | `200` | 默认 K 线数量 |

```bash
# 使用代理
BINANCE_BASE_URL=https://your-proxy.com ./signalbot analyze --asset BTC --timeframe 4h
```

---

## 支持标的

| 友好名称 | Binance 符号 |
|----------|-------------|
| `BTC` | BTCUSDT |
| `XAUUSD` | XAUUSDT |
| `ETH` | ETHUSDT |
| `SOL` | SOLUSDT |
| `BNB` | BNBUSDT |

> 也可直接输入 Binance 符号，如 `--asset BTCUSDT`

---

## JSON 输出结构

```jsonc
{
  "asset": "BTC",
  "timeframe": "4h",
  "timestamp": "2026-03-22T08:00:00Z",
  "price": {
    "current": 87234.5,
    "open": 86500.0,
    "high": 88000.0,
    "low": 86200.0,
    "change_pct": 0.85      // 与上一根收盘价的涨跌幅 (%)
  },
  "indicators": {
    "rsi": {
      "value": 62.5,
      "signal": "bullish"   // overbought | bullish | neutral | bearish | oversold
    },
    "macd": {
      "macd_line": 245.3,
      "signal_line": 198.7,
      "histogram": 46.6,
      "trend": "bullish",   // bullish | bearish | neutral
      "cross": ""           // golden_cross | death_cross | ""（空表示无交叉）
    },
    "bollinger": {
      "upper": 89500.0,
      "middle": 87000.0,    // SMA(20)
      "lower": 84500.0,
      "width": 5000.0,
      "percent_b": 0.55,    // 0=下轨, 0.5=中轨, 1=上轨
      "position": "middle", // above_upper | upper_zone | middle | lower_zone | below_lower
      "signal": "neutral"
    },
    "ema": {
      "ema9": 87100.0,
      "ema21": 86500.0,
      "ema50": 85000.0,
      "ema200": 82000.0,
      "alignment": "strongly_bullish", // strongly_bullish | bullish | neutral | bearish | strongly_bearish
      "signal": "bullish"
    },
    "atr": {
      "value": 1850.0,
      "pct": 2.12,          // ATR 占当前价格的百分比
      "regime": "normal"    // low_volatility | normal | high_volatility
    },
    "volume": {
      "current": 1250000.0,
      "sma20": 980000.0,
      "ratio": 1.28,        // 当前成交量 / SMA20
      "obv": 987654321.0,   // On-Balance Volume（累计）
      "signal": "above_average" // high_volume | above_average | normal | below_average | low_volume
    }
  },
  "analysis": {
    "trend": "bullish",     // bullish | neutral | bearish
    "strength": "moderate", // strong | moderate | weak | none
    "signal": "BUY",        // BUY | SELL | HOLD
    "confidence": 62,       // 0–100，基于多指标共振程度
    "score": 5,             // 原始评分（透明度，-8 到 +8）
    "key_levels": {
      "support": [85000.0, 83200.0, 82000.0],     // 最近 3 个支撑位
      "resistance": [88000.0, 90000.0, 92500.0]   // 最近 3 个阻力位
    }
  }
}
```

---

## 指标说明

### EMA 多头排列（Alignment）
- `strongly_bullish`：EMA9 > EMA21 > EMA50 > EMA200，四线多头完全排列
- `bullish`：EMA9 > EMA21 > EMA50，短中期多头
- `bearish` / `strongly_bearish`：反向排列

### RSI 信号
| 范围 | 信号 | 含义 |
|------|------|------|
| ≥ 70 | overbought | 超买，注意回调风险 |
| 55–69 | bullish | 强势区间 |
| 46–54 | neutral | 中性 |
| 31–45 | bearish | 弱势区间 |
| ≤ 30 | oversold | 超卖，关注反弹机会 |

### MACD Cross
- `golden_cross`：MACD 线上穿信号线（直方柱由负转正）— 强看涨信号
- `death_cross`：MACD 线下穿信号线 — 强看跌信号

### ATR 波动率机制（Regime）
- `low_volatility`：ATR < 1% 价格，市场盘整压缩
- `normal`：1%–3.5%
- `high_volatility`：ATR > 3.5% 价格，波动剧烈

### 综合评分（Score）
各指标独立打分后求和（-8 到 +8），映射为 `confidence` (0–100%) 和 `signal`：
- score ≥ 5 → **BUY** (strong)
- score 2–4 → **BUY** (moderate)
- score ±1 → **HOLD**
- score ≤ -5 → **SELL** (strong)

---

## 项目结构

```
signalbot/
├── main.go                      # CLI 入口
├── config/config.go             # 环境变量配置
└── internal/
    ├── data/
    │   ├── types.go             # Candle 结构体及提取函数
    │   └── binance.go           # Binance Klines REST 客户端
    ├── indicators/
    │   ├── helpers.go           # 公共工具：Round, SMA, StdDev
    │   ├── ema.go               # EMA 9/21/50/200
    │   ├── rsi.go               # RSI(14) Wilder 平滑
    │   ├── macd.go              # MACD(12,26,9) + 金叉/死叉
    │   ├── bollinger.go         # Bollinger Bands(20,2)
    │   ├── atr.go               # ATR(14) Wilder 平滑
    │   └── volume.go            # 成交量 SMA + OBV
    ├── analysis/
    │   └── engine.go            # 多指标评分聚合 + 支撑阻力位计算
    └── report/
        └── report.go            # JSON 输出结构体
```
