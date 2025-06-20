package world

import (
	"log/slog"
	"math"
	"math/rand/v2"
	"sync/atomic"
	"time"

	"github.com/yuta-otsubo/isucon-sutra/bench/internal/concurrent"
	"github.com/yuta-otsubo/isucon-sutra/bench/internal/random"
)

const (
	// LengthOfMinute 仮想世界の1分の長さ
	LengthOfMinute = 1 // 1Tickが1分
	// LengthOfHour 仮想世界の1時間の長さ
	LengthOfHour = LengthOfMinute * 60
	// LengthOfDay 仮想世界の1日の長さ
	LengthOfDay = LengthOfHour * 24
)

type World struct {
	// Time 仮想世界開始からの経過時間
	Time int64
	// Regions 地域
	Regions []*Region
	// UserDB 全ユーザーDB
	UserDB *GenericDB[UserID, *User]
	// ProviderDB 全プロバイダーDB
	ProviderDB *GenericDB[ProviderID, *Provider]
	// ChairDB 全椅子DB
	ChairDB *GenericDB[ChairID, *Chair]
	// RequestDB 全リクエストDB
	RequestDB *RequestDB
	// PaymentDB 支払い結果DB
	PaymentDB *PaymentDB
	// Client webappへのクライアント
	Client WorldClient
	// RootRand ルートの乱数生成器
	RootRand *rand.Rand
	// CompletedRequestChan 完了したリクエストのチャンネル
	CompletedRequestChan chan *Request
	// ErrorCounter エラーカウンター
	ErrorCounter *ErrorCounter

	tickTimeout        time.Duration
	timeoutTicker      *time.Ticker
	prevTimeout        bool
	criticalErrorCh    chan error
	waitingTickCount   atomic.Int32
	userIncrease       float64
	chairIncreaseSales int64
	increasingChairs   atomic.Int64

	// contestantLogger 競技者向けに出力されるロガー
	contestantLogger *slog.Logger

	// TimeoutTickCount タイムアウトしたTickの累計数
	TimeoutTickCount int
}

func NewWorld(tickTimeout time.Duration, completedRequestChan chan *Request, client WorldClient, contestantLogger *slog.Logger) *World {
	return &World{
		Regions: []*Region{
			NewRegion("A", 0, 0, 100, 100),
			NewRegion("B", 300, 300, 100, 100),
		},
		UserDB:               NewGenericDB[UserID, *User](),
		ProviderDB:           NewGenericDB[ProviderID, *Provider](),
		ChairDB:              NewGenericDB[ChairID, *Chair](),
		RequestDB:            NewRequestDB(),
		PaymentDB:            NewPaymentDB(),
		Client:               client,
		RootRand:             random.NewLockedRand(rand.NewPCG(0, 0)),
		CompletedRequestChan: completedRequestChan,
		ErrorCounter:         NewErrorCounter(),
		tickTimeout:          tickTimeout,
		timeoutTicker:        time.NewTicker(tickTimeout),
		criticalErrorCh:      make(chan error),
		userIncrease:         5,
		chairIncreaseSales:   10000,
		contestantLogger:     contestantLogger,
	}
}

func (w *World) Tick(ctx *Context) error {
	if !w.prevTimeout {
		// 前回タイムアウトしなかったら地域毎に増加させる
		for _, region := range w.Regions {
			increase := int(math.Round(w.userIncrease * (float64(region.UserSatisfactionScore()) / 5)))
			if increase > 0 {
				w.contestantLogger.Info("Region内の評判を元にUserが増加します", slog.String("region", region.Name), slog.Int("increase", increase))
				for range increase {
					w.waitingTickCount.Add(1)
					go func() {
						defer w.waitingTickCount.Add(-1)
						_, err := w.CreateUser(ctx, &CreateUserArgs{Region: region})
						if err != nil {
							w.handleTickError(err)
						}
					}()
				}
			}
		}
	}

	for _, p := range w.ProviderDB.Iter() {
		increase := p.TotalSales.Load()/w.chairIncreaseSales - int64(p.ChairDB.Len()) + 10 - w.increasingChairs.Load()
		if increase > 0 {
			w.contestantLogger.Info("一定の売上が立ったためProviderのChairが増加します", slog.Int("id", int(p.ID)), slog.Int64("increase", increase))
			w.increasingChairs.Add(increase)
			for range increase {
				w.waitingTickCount.Add(1)
				go func() {
					defer func() {
						w.waitingTickCount.Add(-1)
						w.increasingChairs.Add(-1)
					}()
					_, err := w.CreateChair(ctx, &CreateChairArgs{
						Provider:          p,
						InitialCoordinate: RandomCoordinateOnRegionWithRand(p.Region, p.Rand),
					})
					if err != nil {
						w.handleTickError(err)
					}
				}()
			}
		}
	}

	for _, c := range w.ChairDB.Iter() {
		w.waitingTickCount.Add(1)
		go func() {
			defer w.waitingTickCount.Add(-1)
			err := c.Tick(ctx)
			if err != nil {
				w.handleTickError(err)
			}
		}()
	}
	for _, u := range w.UserDB.Iter() {
		w.waitingTickCount.Add(1)
		go func() {
			defer w.waitingTickCount.Add(-1)
			err := u.Tick(ctx)
			if err != nil {
				w.handleTickError(err)
			}
		}()
	}
	for _, p := range w.ProviderDB.Iter() {
		w.waitingTickCount.Add(1)
		go func() {
			defer w.waitingTickCount.Add(-1)
			err := p.Tick(ctx)
			if err != nil {
				w.handleTickError(err)
			}
		}()
	}

	select {
	// クリティカルエラーが発生
	case err := <-w.criticalErrorCh:
		return err

	// タイムアウト
	case <-w.timeoutTicker.C:
		if w.waitingTickCount.Load() > 0 {
			// タイムアウトまでにエンティティの行動が全て完了しなかった
			w.TimeoutTickCount++
			w.prevTimeout = true
		} else {
			w.prevTimeout = false
		}
	}

	w.Time++
	return nil
}

type CreateUserArgs struct {
	// Region ユーザーを配置する地域
	Region *Region
}

// CreateUser 仮想世界にユーザーを作成する
func (w *World) CreateUser(ctx *Context, args *CreateUserArgs) (*User, error) {
	registeredData := RegisteredUserData{
		UserName:    random.GenerateUserName(),
		FirstName:   random.GenerateFirstName(),
		LastName:    random.GenerateLastName(),
		DateOfBirth: random.GenerateDateOfBirth(),
	}

	res, err := w.Client.RegisterUser(ctx, &RegisterUserRequest{
		UserName:    registeredData.UserName,
		FirstName:   registeredData.FirstName,
		LastName:    registeredData.LastName,
		DateOfBirth: registeredData.DateOfBirth,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterUser, err)
	}

	u := &User{
		ServerID:          res.ServerUserID,
		Region:            args.Region,
		State:             UserStatePaymentMethodsNotRegister,
		RegisteredData:    registeredData,
		PaymentToken:      random.GeneratePaymentToken(),
		Client:            res.Client,
		Rand:              random.CreateChildRand(w.RootRand),
		notificationQueue: make(chan NotificationEvent, 500),
	}
	w.PaymentDB.PaymentTokens.Set(u.PaymentToken, u)
	result := w.UserDB.Create(u)
	result.Region.UsersDB.Set(result.ID, u)
	return result, nil
}

type CreateProviderArgs struct {
	// Region 椅子を配置する地域
	Region *Region
}

// CreateProvider 仮想世界に椅子のプロバイダーを作成する
func (w *World) CreateProvider(ctx *Context, args *CreateProviderArgs) (*Provider, error) {
	registeredData := RegisteredProviderData{
		Name: random.GenerateProviderName(),
	}

	res, err := w.Client.RegisterProvider(ctx, &RegisterProviderRequest{
		Name: registeredData.Name,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterProvider, err)
	}

	p := &Provider{
		ServerID:       res.ServerProviderID,
		Region:         args.Region,
		ChairDB:        concurrent.NewSimpleMap[ChairID, *Chair](),
		RegisteredData: registeredData,
		Client:         res.Client,
		Rand:           random.CreateChildRand(w.RootRand),
	}
	return w.ProviderDB.Create(p), nil
}

type CreateChairArgs struct {
	// Provider 椅子のプロバイダー
	Provider *Provider
	// InitialCoordinate 椅子の初期位置
	InitialCoordinate Coordinate
}

// CreateChair 仮想世界に椅子を作成する
func (w *World) CreateChair(ctx *Context, args *CreateChairArgs) (*Chair, error) {
	registeredData := RegisteredChairData{
		Name: random.GenerateChairName(),
		// TODO modelの扱い
		Model: random.GenerateChairModel(),
	}

	res, err := args.Provider.Client.RegisterChair(ctx, args.Provider, &RegisterChairRequest{
		Name:  registeredData.Name,
		Model: registeredData.Model,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterChair, err)
	}

	c := &Chair{
		ServerID:          res.ServerUserID,
		Region:            args.Provider.Region,
		Provider:          args.Provider,
		Location:          ChairLocation{Initial: args.InitialCoordinate},
		Speed:             2, // TODO 速度どうする
		State:             ChairStateInactive,
		RegisteredData:    registeredData,
		Client:            res.Client,
		Rand:              random.CreateChildRand(args.Provider.Rand),
		notificationQueue: make(chan NotificationEvent, 500),
	}
	result := w.ChairDB.Create(c)
	result.Provider.ChairDB.Set(result.ID, c)
	return result, nil
}

func (w *World) handleTickError(err error) {
	if errs, ok := UnwrapMultiError(err); ok {
		for _, err2 := range errs {
			w.handleTickError(err2)
		}
	} else if IsCriticalError(err) {
		_ = w.ErrorCounter.Add(err)
		w.criticalErrorCh <- err
	} else {
		w.contestantLogger.Error("エラーが発生しました", slog.String("error", err.Error()))
		if err2 := w.ErrorCounter.Add(err); err2 != nil {
			w.criticalErrorCh <- err2
		}
	}
}

func (w *World) RestTicker() {
	w.timeoutTicker.Reset(w.tickTimeout)
}
