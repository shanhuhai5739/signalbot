package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"signalbot/config"
	"signalbot/internal/analysis"
	"signalbot/internal/data"
)

func main() {
	if err := rootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func rootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "signalbot",
		Short: "量化行情分析工具 — 输出 BTC / 黄金等标的的技术指标 JSON 报告",
		Long: `signalbot 通过 Binance 公共 API 获取 K 线数据，
计算 RSI、MACD、布林带、EMA、ATR、成交量等技术指标，
并输出结构化 JSON 报告供 LLM 或下游工具消费。

数据源：Binance 公共 API（无需 API Key）
  现货：BTC、ETH、SOL、BNB 等
  期货：XAUUSD（黄金，USD-M 永续合约）`,
	}
	root.AddCommand(analyzeCmd())
	root.AddCommand(multiCmd())
	return root
}

// analyzeCmd 分析单个标的单个时间周期
func analyzeCmd() *cobra.Command {
	var (
		asset     string
		timeframe string
		limit     int
		output    string
	)

	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "分析指定标的单个时间周期，输出技术指标 JSON 报告",
		Example: `  # 分析 BTC 4小时行情，输出到终端
  signalbot analyze --asset BTC --timeframe 4h

  # 分析黄金日线行情，保存到文件
  signalbot analyze --asset XAUUSD --timeframe 1d --output report.json

  # 拉取更多 K 线（默认 200 根）
  signalbot analyze --asset BTC --timeframe 1h --limit 500`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			if limit <= 0 {
				limit = cfg.DefaultLimit
			}
			if limit > 1500 {
				limit = 1500
			}
			const minRequired = 210
			if limit < minRequired {
				limit = minRequired
			}

			provider := data.NewProvider(cfg)
			ctx := context.Background()

			fmt.Fprintf(os.Stderr, "正在获取 %s %s K线数据 (%d 根)...\n",
				strings.ToUpper(asset), timeframe, limit)

			candles, err := provider.FetchKlines(ctx, asset, timeframe, limit)
			if err != nil {
				return fmt.Errorf("获取行情失败: %w", err)
			}

			if len(candles) < 50 {
				return fmt.Errorf(
					"数据不足: 收到 %d 根 K线，至少需要 50 根。请检查标的名称和时间周期是否正确",
					len(candles),
				)
			}

			fmt.Fprintf(os.Stderr, "收到 %d 根 K线，正在计算指标...\n", len(candles))

			rep := analysis.Analyze(strings.ToUpper(asset), timeframe, candles)

			if output != "" {
				if err := rep.Save(output); err != nil {
					return fmt.Errorf("保存报告失败: %w", err)
				}
				fmt.Fprintf(os.Stderr, "报告已保存到: %s\n", output)
				return nil
			}

			return rep.WriteJSON(os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&asset, "asset", "a", "", "标的资产，如 BTC、XAUUSD（必填）")
	cmd.Flags().StringVarP(&timeframe, "timeframe", "t", "4h",
		"K线时间周期: 1m 5m 15m 30m 1h 2h 4h 6h 8h 12h 1d 3d 1w 1M")
	cmd.Flags().IntVarP(&limit, "limit", "l", 0,
		"获取 K线数量，默认由 DEFAULT_LIMIT 环境变量控制（默认 200，最多 1500）")
	cmd.Flags().StringVarP(&output, "output", "o", "",
		"JSON 输出文件路径，不指定则打印到 stdout")

	_ = cmd.MarkFlagRequired("asset")
	return cmd
}

// multiCmd 同时分析单个标的的 1h/4h/1d/1w/1M 五个时间维度
func multiCmd() *cobra.Command {
	var (
		asset  string
		output string
	)

	cmd := &cobra.Command{
		Use:   "multi",
		Short: "多周期综合分析：并发分析 1h/4h/日线/周线/月线（年线趋势参考）五个时间维度",
		Example: `  # 分析 BTC 所有时间维度
  signalbot multi --asset BTC

  # 分析黄金多周期，保存到文件
  signalbot multi --asset XAUUSD --output gold_multi.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			provider := data.NewProvider(cfg)
			ctx := context.Background()

			fmt.Fprintf(os.Stderr, "正在并发拉取 %s 五个时间维度数据（1h/4h/1d/1w/1M）...\n",
				strings.ToUpper(asset))

			rep := analysis.AnalyzeMulti(ctx, strings.ToUpper(asset), provider)

			fmt.Fprintf(os.Stderr, "分析完成，共 %d 个时间维度\n", len(rep.Timeframes))

			if output != "" {
				if err := rep.Save(output); err != nil {
					return fmt.Errorf("保存报告失败: %w", err)
				}
				fmt.Fprintf(os.Stderr, "报告已保存到: %s\n", output)
				return nil
			}

			return rep.WriteJSON(os.Stdout)
		},
	}

	cmd.Flags().StringVarP(&asset, "asset", "a", "", "标的资产，如 BTC、XAUUSD（必填）")
	cmd.Flags().StringVarP(&output, "output", "o", "",
		"JSON 输出文件路径，不指定则打印到 stdout")

	_ = cmd.MarkFlagRequired("asset")
	return cmd
}
