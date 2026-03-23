# signalbot

BTC、黄金等标的的量化行情分析 CLI 工具。

通过 **Binance 公共 Klines API**（无需 API Key）获取 K 线数据，计算 **10 类技术指标**，输出结构化 **JSON 报告**，可直接供 LLM（如 OpenClaw、GPT）读取并生成行情分析推文。

**支持的指标：** RSI · MACD · 布林带 · EMA · ATR · 成交量/OBV · 顾比均线(GMMA) · 斐波那契回撤 · 锚定VWAP · 固定范围成交量分布(VPVR)

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
# 分析 BTC 4小时行情（含全部 10 类指标）
./signalbot analyze --asset BTC --timeframe 4h

# 分析黄金日线行情，结果写入文件
./signalbot analyze --asset XAUUSD --timeframe 1d --output gold_report.json

# 多周期综合分析（并发 1h/4h/1d/1w/1M）
./signalbot multi --asset BTC

# 拉取更多 K 线（顾比均线 EMA377 需要 377+ 根）
./signalbot analyze --asset BTC --timeframe 1d --limit 500

# 管道给 LLM CLI（以 ollama 为例）
./signalbot analyze --asset BTC --timeframe 4h | \
  ollama run llama3 "根据以下行情 JSON，以专业交易员视角撰写一条行情分析推文："
```

---

## CLI 参数

### `analyze` — 单周期分析

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--asset` | `-a` | **必填** | 标的资产：`BTC`、`XAUUSD`、`ETH`、`SOL` 等 |
| `--timeframe` | `-t` | `4h` | K 线周期：`1m` `5m` `15m` `30m` `1h` `2h` `4h` `6h` `8h` `12h` `1d` `3d` `1w` `1M` |
| `--limit` | `-l` | `200` | 获取 K 线数量（最多 1500，自动补足至 210 以保证 EMA200 收敛）|
| `--output` | `-o` | stdout | JSON 输出文件路径，不填则打印到 stdout |

### `multi` — 多周期综合分析

```bash
./signalbot multi --asset BTC
```

并发拉取并分析 **1h / 4h / 1d / 1w / 1M** 五个时间维度，返回含跨周期 `summary` 的综合报告。

| 参数 | 简写 | 默认值 | 说明 |
|------|------|--------|------|
| `--asset` | `-a` | **必填** | 标的资产 |
| `--output` | `-o` | stdout | JSON 输出文件路径 |

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

| 友好名称 | Binance 符号 | 数据源 |
|----------|-------------|--------|
| `BTC` | BTCUSDT | 现货 |
| `XAUUSD` | XAUUSDT | USD-M 期货 |
| `ETH` | ETHUSDT | 现货 |
| `SOL` | SOLUSDT | 现货 |
| `BNB` | BNBUSDT | 现货 |

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
    "change_pct": 0.85
  },
  "indicators": {
    "rsi": { "value": 62.5, "signal": "bullish" },
    "macd": {
      "macd_line": 245.3, "signal_line": 198.7, "histogram": 46.6,
      "trend": "bullish", "cross": "golden_cross"
    },
    "bollinger": {
      "upper": 89500.0, "middle": 87000.0, "lower": 84500.0,
      "width": 5000.0, "percent_b": 0.55,
      "position": "middle", "signal": "neutral"
    },
    "ema": {
      "ema9": 87100.0, "ema21": 86500.0, "ema50": 85000.0, "ema200": 82000.0,
      "alignment": "strongly_bullish", "signal": "bullish"
    },
    "atr": { "value": 1850.0, "pct": 2.12, "regime": "normal" },
    "volume": {
      "current": 1250000.0, "sma20": 980000.0, "ratio": 1.28,
      "obv": 987654321.0, "signal": "above_average"
    },

    // ── 新增指标 ──────────────────────────────────────────────────

    "guppy": {
      // 短期组（投机/交易者）
      "ema3": 87800.0, "ema5": 87600.0, "ema8": 87300.0,
      "ema10": 87200.0, "ema13": 87000.0, "ema21": 86700.0,
      // 长期组（机构/投资者）
      "ema34": 85500.0, "ema55": 84000.0, "ema89": 82000.0,
      "ema144": 79000.0, "ema233": 0, "ema377": 0,  // 0 = 数据不足
      "short_min": 86700.0, "short_max": 87800.0,
      "long_min": 79000.0, "long_max": 85500.0,
      "gap_pct": 1.38,          // 短期组高于长期组的幅度(%)
      "alignment": "above_long",// above_long | crossing | below_long
      "signal": "bullish"       // bullish | compression | bearish
    },

    "fibonacci": {
      "swing_high": 97924.49, "swing_low": 60000.0, "range": 37924.49,
      "levels": [
        { "label": "0.0%",   "ratio": 0.0,   "price": 97924.49, "is_above": false },
        { "label": "23.6%",  "ratio": 0.236, "price": 88974.31, "is_above": false },
        { "label": "38.2%",  "ratio": 0.382, "price": 83437.33, "is_above": true  },
        { "label": "50.0%",  "ratio": 0.5,   "price": 78962.25, "is_above": true  },
        { "label": "61.8%",  "ratio": 0.618, "price": 74487.16, "is_above": true  },
        { "label": "78.6%",  "ratio": 0.786, "price": 68115.84, "is_above": true  },
        { "label": "100.0%", "ratio": 1.0,   "price": 60000.0,  "is_above": true  }
      ],
      "nearest_level": "38.2%",  // 当前价最近的水平
      "distance_pct": 0.42,      // 偏离该水平的 % 距离
      "signal": "at_resistance", // at_support | at_resistance | between_levels
      "direction": "upper_half"  // upper_half | lower_half（相对于摆动区间中点）
    },

    "vwap": {
      "value": 86500.0,          // 锚定 VWAP（最近50根K线）
      "upper_band1": 88200.0,    // VWAP + 1σ
      "lower_band1": 84800.0,    // VWAP − 1σ
      "upper_band2": 89900.0,    // VWAP + 2σ
      "lower_band2": 83100.0,    // VWAP − 2σ
      "std_dev": 1700.0,
      "deviation_pct": 0.85,     // 当前价偏离 VWAP 的 %
      "position": "above_vwap",  // above_band2|above_band1|above_vwap|below_vwap|below_band1|below_band2
      "signal": "bullish"        // overbought | bullish | bearish | oversold
    },

    "vpvr": {
      "poc": 85200.0,            // Point of Control（成交量最密集价位）
      "vah": 90500.0,            // Value Area High（70% 价值区上沿）
      "val": 78000.0,            // Value Area Low（70% 价值区下沿）
      "num_bins": 24,
      "bins": [                  // 24 个价格档，按成交量降序
        {
          "price_low": 84500.0, "price_high": 85900.0, "price_mid": 85200.0,
          "volume": 272908.2,
          "is_poc": true, "is_value_area": true
        }
        // ... 其余 23 档 ...
      ],
      "signal": "above_poc"      // above_vah|above_poc|at_poc|below_poc|below_val
    }
  },
  "analysis": {
    "trend": "bullish",
    "strength": "moderate",
    "signal": "BUY",
    "confidence": 62,
    "score": 5,
    "key_levels": {
      "support": [85000.0, 83200.0, 82000.0],
      "resistance": [88000.0, 90000.0, 92500.0]
    }
  }
}
```

---

## 指标说明

### 顾比均线 (GMMA)

两组 EMA 的相对位置反映不同周期资金的博弈：

| alignment | 含义 |
|-----------|------|
| `above_long` | 短期组整体高于长期组 → 趋势强劲，机构未阻击 |
| `crossing` | 两组交叉重叠 → 压缩期，趋势转换观察信号 |
| `below_long` | 短期组整体低于长期组 → 下行趋势，机构主导下行 |

> EMA233/EMA377 需要 233+/377+ 根 K 线，可用 `--limit 400` 补足

### 斐波那契回撤

基于回看 100 根 K 线的摆动高低点自动计算七个水平（0%→100%）。`signal` 在距某水平 ±1.5% 内触发：
- `at_support` → 贴近支撑位，关注反弹
- `at_resistance` → 贴近阻力位，关注回调

### 锚定 VWAP

以最近 50 根 K 线为锚定窗口，计算量价加权均价及 ±1σ/±2σ 通道。价格偏离 ±2σ 时为极端区域，均值回归概率较高。

### 固定范围成交量分布 (VPVR)

取最近 100 根 K 线、24 个价格档，按蜡烛图跨越区间**按比例**分配成交量：
- **POC**（Point of Control）：成交量最密集价位，最强支撑/阻力
- **VAH/VAL**：70% 价值区间上下边界
- 突破 VAH 且放量 → 强势信号；跌破 VAL 且放量 → 弱势信号

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
├── main.go                      # CLI 入口（analyze / multi 两个子命令）
├── config/config.go             # 环境变量配置
└── internal/
    ├── data/
    │   ├── types.go             # Candle 结构体及提取函数
    │   ├── binance.go           # Binance Klines REST 客户端（现货+期货）
    │   └── provider.go          # 统一数据提供者接口
    ├── indicators/
    │   ├── helpers.go           # 公共工具：Round, SMA, StdDev, SliceMin, SliceMax
    │   ├── ema.go               # EMA 9/21/50/200
    │   ├── rsi.go               # RSI(14) Wilder 平滑
    │   ├── macd.go              # MACD(12,26,9) + 金叉/死叉
    │   ├── bollinger.go         # Bollinger Bands(20,2)
    │   ├── atr.go               # ATR(14) Wilder 平滑
    │   ├── volume.go            # 成交量 SMA + OBV
    │   ├── guppy.go             # 顾比均线 GMMA（短期3/5/8/10/13/21，长期34/55/89/144/233/377）
    │   ├── fibonacci.go         # 斐波那契回撤（100根K线摆动高低点，7个水平）
    │   ├── vwap.go              # 锚定 VWAP + ±1σ/±2σ 标准差通道
    │   └── vpvr.go              # 固定范围成交量分布（24档，POC/VAH/VAL）
    ├── analysis/
    │   ├── engine.go            # 多指标评分聚合 + 支撑阻力位计算
    │   └── multi.go             # 多周期并发分析 + 跨周期信号汇总
    └── report/
        └── report.go            # JSON 输出结构体（含 MultiReport）
```
