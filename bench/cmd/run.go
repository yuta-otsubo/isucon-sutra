package cmd

import (
	"context"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/worker"
	"github.com/spf13/cobra"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp"
	"github.com/yuta-otsubo/isucon-sutra/bench/internal/logger"
	"go.uber.org/zap"
)

// ベンチマークターゲット
var target string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a benchmark",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := zap.L()
		defer l.Sync()

		contestantLogger, err := logger.CreateContestantLogger()
		if err != nil {
			l.Error("Failed to create contestant logger", zap.Error(err))
			return err
		}

		b, err := isucandar.NewBenchmark(
			isucandar.WithoutPanicRecover(),
			isucandar.WithLoadTimeout(1*time.Second),
		)
		if err != nil {
			l.Error("Failed to create benchmark", zap.Error(err))
			return err
		}

		b.Prepare(func(ctx context.Context, step *isucandar.BenchmarkStep) error {
			client, err := webapp.NewClient(webapp.ClientConfig{
				TargetBaseURL:         target,
				DefaultClientTimeout:  5 * time.Second,
				ClientIdleConnTimeout: 10 * time.Second,
				InsecureSkipVerify:    true,
				ContestantLogger:      contestantLogger,
			})
			if err != nil {
				return err
			}

			_, err = client.PostInitialize(ctx)
			if err != nil {
				return err
			}

			return nil
		})
		b.Load(func(ctx context.Context, step *isucandar.BenchmarkStep) error {
			client, err := webapp.NewClient(webapp.ClientConfig{
				TargetBaseURL:         target,
				DefaultClientTimeout:  5 * time.Second,
				ClientIdleConnTimeout: 10 * time.Second,
				InsecureSkipVerify:    true,
				ContestantLogger:      contestantLogger,
			})
			if err != nil {
				return err
			}

			w, err := worker.NewWorker(func(ctx context.Context, _ int) {
				err = client.GetPing(ctx)
				if err != nil {
					step.AddError(err)
					return
				}
				step.AddScore("ping")
			}, worker.WithMaxParallelism(10))
			w.Process(ctx)

			return nil
		})

		l.Info("benchmark started")
		result := b.Start(context.Background())
		result.Score.Set("ping", 1)

		l.Info("benchmark finished", zap.Int64("score", result.Score.Total()))
		return nil
	},
}

func init() {
	runCmd.Flags().StringVar(&target, "target", "http://localhost:8080", "benchmark target")
	rootCmd.AddCommand(runCmd)
}
