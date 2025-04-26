package cmd

import (
	"os"

	// cobraはGo言語用のコマンドラインアプリケーションフレームワーク
	// 引数の処理やサブコマンドの実装を簡素化する
	"github.com/spf13/cobra"
	"github.com/yuta-otsubo/isucon-sutra/bench/internal/logger"
)

var rootCmd = &cobra.Command{
	Use:     "bench",
	Short:   "ISUCON14 benchmark",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logger.SetupGlobalLogger()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
