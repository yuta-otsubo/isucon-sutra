package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"sync"
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
	// TimeOfDay 仮想世界の1日の時刻
	TimeOfDay int
	// Regions 地域
	Regions map[int]*Region
	// UserDB 全ユーザーDB
	UserDB *GenericDB[UserID, *User]
	// ChairDB 全椅子DB
	ChairDB *GenericDB[ChairID, *Chair]
	// RequestDB 全リクエストDB
	RequestDB *RequestDB
	// RootRand ルートの乱数生成器
	RootRand *rand.Rand
	// CompletedRequestChan 完了したリクエストのチャンネル
	CompletedRequestChan chan *Request

	tickTimeout   time.Duration
	timeoutTicker *time.Ticker
}

func NewWorld(tickTimeout time.Duration, completedRequestChan chan *Request) *World {
	region := &Region{
		RegionWidth:   30,
		RegionHeight:  30,
		RegionOffsetX: 0,
		RegionOffsetY: 0,
	}
	return &World{
		Regions:   map[int]*Region{1: region},
		UserDB:    NewGenericDB[UserID, *User](),
		ChairDB:   NewGenericDB[ChairID, *Chair](),
		RequestDB: NewRequestDB(),
		// TODO シードをどうする
		RootRand:             random.NewLockedRand(rand.NewPCG(0, 0)),
		CompletedRequestChan: completedRequestChan,
		tickTimeout:          tickTimeout,
		timeoutTicker:        time.NewTicker(1 * time.Hour),
	}
}

func (w *World) Tick(ctx *Context) {
	var wg sync.WaitGroup

	w.timeoutTicker.Reset(w.tickTimeout)

	for _, c := range w.ChairDB.Iter() {
		// 前のTickの処理が完了していない椅子は完了するまで新しい時間はスキップする
		if c.TickCompleted() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := c.Tick(ctx)
				if err != nil {
					// TODO: エラーペナルティ
					log.Println(err)
				}
			}()
		}
	}
	for _, u := range w.UserDB.Iter() {
		// 前のTickの処理が完了していないユーザーは完了するまで新しい時間はスキップする
		if u.TickCompleted() {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := u.Tick(ctx)
				if err != nil {
					// TODO: エラーペナルティ
					log.Println(err)
				}
			}()
		}
	}

	select {
	case <-concurrent.WaitChan(&wg):
		// タイムアウトする前に完了

	case <-w.timeoutTicker.C:
		timeoutChair := 0
		timeoutUser := 0
		for _, c := range w.ChairDB.Iter() {
			if !c.TickCompleted() {
				timeoutChair++
			}
		}
		for _, u := range w.UserDB.Iter() {
			if !u.TickCompleted() {
				timeoutUser++
			}
		}
		if timeoutUser > 0 || timeoutChair > 0 {
			// タイムアウト数計算途中に完了した場合はタイムアウトしなかった扱いにする
			log.Printf("tick timeout (time: %d, timeout users: %d, timeout chairs: %d)", w.Time, timeoutUser, timeoutChair)
		}
	}

	w.Time++
	w.TimeOfDay = int(w.Time % LengthOfDay)
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

	res, err := ctx.client.RegisterUser(ctx, &RegisterUserRequest{
		UserName:    registeredData.UserName,
		FirstName:   registeredData.FirstName,
		LastName:    registeredData.LastName,
		DateOfBirth: registeredData.DateOfBirth,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterUser, err)
	}

	u := &User{
		ServerID:       res.ServerUserID,
		Region:         args.Region,
		State:          UserStateInactive,
		RegisteredData: registeredData,
		AccessToken:    res.AccessToken,
		Rand:           random.CreateChildRand(w.RootRand),
	}
	u.tickDone.Store(true)
	return w.UserDB.Create(u), nil
}

type CreateChairArgs struct {
	// Region 椅子を配置する地域
	Region *Region
	// InitialCoordinate 椅子の初期位置
	InitialCoordinate Coordinate
	// WorkTime 稼働時間
	WorkTime Interval[int]
}

// CreateChair 仮想世界に椅子を作成する
func (w *World) CreateChair(ctx *Context, args *CreateChairArgs) (*Chair, error) {
	registeredData := RegisteredChairData{
		UserName:    random.GenerateUserName(),
		FirstName:   random.GenerateFirstName(),
		LastName:    random.GenerateLastName(),
		DateOfBirth: random.GenerateDateOfBirth(),
		// TODO model, noの扱い
		ChairModel: "ISU_X",
		ChairNo:    fmt.Sprintf("%d", rand.Uint32()),
	}

	res, err := ctx.client.RegisterChair(ctx, &RegisterChairRequest{
		UserName:    registeredData.UserName,
		FirstName:   registeredData.FirstName,
		LastName:    registeredData.LastName,
		DateOfBirth: registeredData.DateOfBirth,
		ChairModel:  registeredData.ChairModel,
		ChairNo:     registeredData.ChairNo,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterChair, err)
	}

	c := &Chair{
		ServerID:       res.ServerUserID,
		Region:         args.Region,
		Current:        args.InitialCoordinate,
		Speed:          2, // TODO 速度どうする
		State:          ChairStateInactive,
		WorkTime:       args.WorkTime,
		RegisteredData: registeredData,
		AccessToken:    res.AccessToken,
		Rand:           random.CreateChildRand(w.RootRand),
	}
	c.tickDone.Store(true)
	return w.ChairDB.Create(c), nil
}
