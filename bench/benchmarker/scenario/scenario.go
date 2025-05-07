package scenario

import (
	"context"
	"time"

	"github.com/isucon/isucandar"
	"github.com/isucon/isucandar/score"
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

	return &Scenario{
		target:           target,
		contestantLogger: contestantLogger,
		world:            w,
		worldCtx:         worldCtx,

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

	_, err = client.PostInitialize(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Load はシナリオのメイン処理を行う
func (s *Scenario) Load(ctx context.Context, step *isucandar.BenchmarkStep) error {
	// agent, err := verify.NewAgent(s.target, s.contestantLogger)
	// if err != nil {
	// 	s.contestantLogger.Error("Failed to create agent", zap.Error(err))
	// 	return err
	// }

	// if err := agent.Run(); err != nil {
	// 	s.contestantLogger.Error("Failed to run agent", zap.Error(err))
	// 	return err
	// }
	//w, err := worker.NewWorker(func(ctx context.Context, _ int) {
	//	agent, err := verify.NewAgent(s.target, s.contestantLogger)
	//	if err != nil {
	//		s.contestantLogger.Error("Failed to create agent", zap.Error(err))
	//		return
	//	}
	//
	//	if err := agent.Run(); err != nil {
	//		s.contestantLogger.Error("Failed to run agent", zap.Error(err))
	//	}
	//}, worker.WithMaxParallelism(10))
	//if err != nil {
	//	return err
	//}
	//
	//w.Process(ctx)

	go func() {
		for req := range s.completedRequestChan {
			s.contestantLogger.Info("request completed", zap.Any("request", req))
			step.AddScore(score.ScoreTag("completed_request"))
		}
	}()

	region := s.world.Regions[1]
	for range 1 {
		_, err := s.world.CreateChair(s.worldCtx, &world.CreateChairArgs{
			Region:            region,
			InitialCoordinate: world.RandomCoordinateOnRegion(region),
			WorkTime:          world.NewInterval(world.ConvertHour(0), world.ConvertHour(23)),
		})
		if err != nil {
			return err
		}
	}
	for range 1 {
		u, err := s.world.CreateUser(s.worldCtx, &world.CreateUserArgs{Region: region})
		if err != nil {
			return err
		}
		u.State = world.UserStateActive
	}

	// TODO webapp側でマッチングさせる
	// go func() {
	// 	for id := range s.requestQueue {
	// 		matched := false
	// 		for _, chair := range s.world.ChairDB.Iter() {
	// 			if chair.State == world.ChairStateActive && !chair.ServerRequestID.Valid {
	// 				if f, ok := s.chairNotificationReceiverMap.Get(chair.ServerID); ok {
	// 					f(&world.ChairNotificationEventMatched{ServerRequestID: id})
	// 				}
	// 				matched = true
	// 				break
	// 			}
	// 		}
	// 		if !matched {
	// 			s.requestQueue <- id
	// 		}
	// 	}
	// }()

	for now := range world.ConvertHour(24 * 14) {
		s.world.Tick(s.worldCtx)

		if now%world.ConvertHour(1) == 0 {
			s.contestantLogger.Info("tick", zap.Int("time", now/world.ConvertHour(1)))
		}
	}

	return nil
}

// Validation はシナリオの結果検証処理を行う
func (s *Scenario) Validation(ctx context.Context, step *isucandar.BenchmarkStep) error {
	return nil
}
