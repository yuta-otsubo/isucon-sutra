package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"sync"

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
}

func (w *World) Tick(ctx *Context) {
	var wg sync.WaitGroup

	for _, c := range w.ChairDB.Iter() {
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
	for _, u := range w.UserDB.Iter() {
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

	wg.Wait()

	w.Time++
	w.TimeOfDay = int(w.Time % LengthOfDay)
}

// UpdateRequestChairStatus 椅子が認識しているリクエストのステータスを変更する
func (w *World) UpdateRequestChairStatus(chairID ChairID, status RequestStatus) error {
	chair := w.ChairDB.Get(chairID)
	return chair.ChangeRequestStatus(status)
}

// UpdateRequestUserStatus ユーザーが認識しているリクエストのステータスを変更する
func (w *World) UpdateRequestUserStatus(userID UserID, status RequestStatus) error {
	user := w.UserDB.Get(userID)
	return user.ChangeRequestStatus(status)
}

// AssignRequest 椅子にリクエストを割り当てる
func (w *World) AssignRequest(chairID ChairID, serverRequestID string) error {
	chair := w.ChairDB.Get(chairID)
	return chair.AssignRequest(serverRequestID)
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

	return w.UserDB.Create(&User{
		ServerID:       res.ServerUserID,
		Region:         args.Region,
		State:          UserStateInactive,
		RegisteredData: registeredData,
		AccessToken:    res.AccessToken,
	}), nil
}

type CreateChairArgs struct {
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

	return w.ChairDB.Create(&Chair{
		ServerID:       res.ServerUserID,
		Current:        args.InitialCoordinate,
		Speed:          2, // TODO 速度どうする
		State:          ChairStateInactive,
		WorkTime:       args.WorkTime,
		RegisteredData: registeredData,
		AccessToken:    res.AccessToken,
	}), nil
}
