package scenario

import (
	"context"
	"net/http"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/score"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp/api"
	"github.com/yuta-otsubo/isucon-sutra/bench/payment"
	"go.uber.org/zap"

	// "github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/scenario/agents/verify"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/scenario/worldclient"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/world"
)

// Scenario はシナリオを表す構造体
// 以下の関数を実装して b.AddScenario() で追加したあと b.Run() で実行される
// - Prepare(context.Context, *BenchmarkStep) error
//   - シナリオの初期化処理を行う
//   - Initialize した後に Validation を呼ぶことが多いっぽい
//   - Initialize -> Validation -> Initialize してもいいかも？（13とかはそうしてそう）
//
// - Load(context.Context, *BenchmarkStep) error
//   - シナリオのメイン処理を行う
//
// - Validation(context.Context, *BenchmarkStep) error
//   - シナリオの結果検証処理を行う
//   - 料金の整合性をみたいかも
type Scenario struct {
	target           string
	contestantLogger *zap.Logger
	world            *world.World
	worldCtx         *world.Context
	paymentServer    *payment.Server
	step             *isucandar.BenchmarkStep

	requestQueue         chan string // あんまり考えて導入してないです
	completedRequestChan chan *world.Request
}

func NewScenario(target string, contestantLogger *zap.Logger) *Scenario {
	requestQueue := make(chan string, 1000)
	completedRequestChan := make(chan *world.Request, 1000)
	w := world.NewWorld(30*time.Millisecond, completedRequestChan)
	worldClient := worldclient.NewWorldClient(context.Background(), w, webapp.ClientConfig{
		TargetBaseURL:         target,
		DefaultClientTimeout:  5 * time.Second,
		ClientIdleConnTimeout: 10 * time.Second,
		InsecureSkipVerify:    true,
		ContestantLogger:      contestantLogger,
	}, requestQueue, contestantLogger)
	worldCtx := world.NewContext(w, worldClient)
	paymentServer := payment.NewServer(w.PaymentDB, 300*time.Millisecond, 5)
	// TODO: サーバーハンドリング
	go func() {
		http.ListenAndServe(":12345", paymentServer)
	}()

	return &Scenario{
		target:           target,
		contestantLogger: contestantLogger,
		world:            w,
		worldCtx:         worldCtx,
		paymentServer:    paymentServer,

		requestQueue:         requestQueue,
		completedRequestChan: completedRequestChan,
	}
}

// Prepare はシナリオの初期化処理を行う
func (s *Scenario) Prepare(ctx context.Context, step *isucandar.BenchmarkStep) error {
	client, err := webapp.NewClient(webapp.ClientConfig{
		TargetBaseURL:         s.target,
		DefaultClientTimeout:  5 * time.Second,
		ClientIdleConnTimeout: 10 * time.Second,
		InsecureSkipVerify:    true,
		ContestantLogger:      s.contestantLogger,
	})
	if err != nil {
		return err
	}

	// TODO: 決済サーバーアドレス
	_, err = client.PostInitialize(ctx, &api.PostInitializeReq{PaymentServer: "http://localhost:12345"})
	if err != nil {
		return err
	}

	const (
		initialProvidersNum         = 5
		initialChairsNumPerProvider = 10
		initialUsersNum             = 10
	)

	for i := range initialProvidersNum {
		provider, err := s.world.CreateProvider(s.worldCtx, &world.CreateProviderArgs{
			Region: s.world.Regions[i%len(s.world.Regions)],
		})
		if err != nil {
			return err
		}

		for range initialChairsNumPerProvider {
			_, err := s.world.CreateChair(s.worldCtx, &world.CreateChairArgs{
				Provider:          provider,
				InitialCoordinate: world.RandomCoordinateOnRegion(provider.Region),
				WorkTime:          world.NewInterval(world.ConvertHour(0), world.ConvertHour(2000)),
			})
			if err != nil {
				return err
			}
		}
	}
	for i := range initialUsersNum {
		_, err := s.world.CreateUser(s.worldCtx, &world.CreateUserArgs{Region: s.world.Regions[i%len(s.world.Regions)]})
		if err != nil {
			return err
		}
	}

	return nil
}

// Load はシナリオのメイン処理を行う
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	go func() {
		for req := range s.completedRequestChan {
			s.contestantLogger.Info("request completed", zap.Stringer("request", req), zap.Stringer("eval", req.CalculateEvaluation()))
			step.AddScore(score.ScoreTag("completed_request"))
		}
	}()

	s.world.RestTicker()
LOOP:
	for {
		select {
		case <-ctx.Done():
			// 負荷走行終了
			break LOOP

		default:
			err := s.world.Tick(s.worldCtx)
			if err != nil {
				s.contestantLogger.Error("critical error", zap.Error(err))
				return err
			}

			if s.world.Time%world.LengthOfHour == 0 {
				s.contestantLogger.Info("tick",
					zap.Int64("ticks", s.world.Time),
					zap.Int("timeouts", s.world.TimeoutTickCount),
					zap.Float64("timeouts(%)", float64(s.world.TimeoutTickCount)/float64(s.world.Time)*100),
				)
			}
		}
	}

	return nil
}

// Validation はシナリオの結果検証処理を行う
func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	return nil
}
