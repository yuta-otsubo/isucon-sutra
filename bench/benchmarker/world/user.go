package world

import "fmt"

type UserState int

const (
	UserStateInactive UserState = iota
	UserStateActive
)

type UserID int

type User struct {
	// ID ベンチマーカー内部ユーザーID
	ID UserID
	// ServerID サーバー上でのユーザーID
	ServerID string
	// Region ユーザーが居る地域
	Region *Region
	// State ユーザーの状態
	State UserState
	// Request 進行中の配椅子・送迎リクエスト
	Request *Request

	// RegisteredData サーバーに登録されているユーザー情報
	RegisteredData RegisteredUserData
	// AccessToken サーバーアクセストークン
	AccessToken string
	// RequestHistory リクエスト履歴
	RequestHistory []*Request
}

func (u *User) String() string {
	if u.Request != nil {
		return fmt.Sprintf("User{id=%d,totalReqs=%d,reqId=%d}", u.ID, len(u.RequestHistory), u.Request.ID)
	}
	return fmt.Sprintf("User{id=%d,totalReqs=%d}", u.ID, len(u.RequestHistory))
}

func (u *User) SetID(id UserID) {
	u.ID = id
}

func (u *User) Tick(ctx *Context) error {
	switch {
	// 進行中のリクエストが存在
	case u.Request != nil:
		switch u.Request.UserStatus {
		case RequestStatusMatching:
			// マッチングされるまで待機する
			// TODO: 待たされ続けた場合のキャンセル
			break

		case RequestStatusDispatching:
			// 椅子が到着するまで待つ
			// TODO: 椅子が一向に到着しない場合の対応
			break

		case RequestStatusDispatched:
			// 椅子が出発するのを待つ
			// TODO: 椅子が一向に出発しない場合の対応
			break

		case RequestStatusCarrying:
			// 椅子が到着するのを待つ
			// TODO: 椅子が一向に到着しない場合の対応
			break

		case RequestStatusArrived:
			// 送迎の評価を行う
			err := ctx.client.SendEvaluation(ctx, u.Request)
			if err != nil {
				return WrapCodeError(ErrorCodeFailedToEvaluate, err)
			}

			// サーバーが評価を受理したので完了状態にする
			u.Request.DesiredStatus = RequestStatusCompleted
			u.Request.UserStatus = RequestStatusCompleted

		case RequestStatusCompleted:
			// 進行中のリクエストが無い状態にする
			u.Request = nil

		case RequestStatusCanceled:
			// ここに分岐することはありえない
			panic("unexpected state")
		}

	// 進行中のリクエストは存在しないが、ユーザーがアクティブ状態
	case u.Request == nil && u.State == UserStateActive:
		// リクエストを作成する
		// TODO 作成する条件・頻度
		err := u.CreateRequest(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (u *User) CreateRequest(ctx *Context) error {
	if u.Request != nil {
		panic("ユーザーに進行中のリクエストがあるのにも関わらず、リクエストを新規作成しようとしている")
	}

	// TODO 目的地の決定方法をランダムじゃなくする
	pickup := RandomCoordinateOnRegionWithRand(u.Region, ctx.rand)
	dest := RandomCoordinateAwayFromHereWithRand(pickup, ctx.rand.IntN(100)+5, ctx.rand)

	req := &Request{
		User:             u,
		PickupPoint:      pickup,
		DestinationPoint: dest,
		RequestedAt:      ctx.world.Time,
		DesiredStatus:    RequestStatusMatching,
		ChairStatus:      RequestStatusMatching,
		UserStatus:       RequestStatusMatching,
	}
	res, err := ctx.client.SendCreateRequest(ctx, req)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToCreateRequest, err)
	}
	req.ServerID = res.ServerRequestID
	u.Request = req
	u.RequestHistory = append(u.RequestHistory, req)
	ctx.world.RequestDB.Create(req)
	return nil
}

func (u *User) ChangeRequestStatus(status RequestStatus) error {
	request := u.Request
	if request == nil {
		return CodeError(ErrorCodeUserNotRequestingButStatusChanged)
	}
	if request.DesiredStatus != status {
		return CodeError(ErrorCodeUnexpectedStatusTransitionOccurred)
	}
	request.UserStatus = status
	return nil
}

type RegisteredUserData struct {
	UserName    string
	FirstName   string
	LastName    string
	DateOfBirth string
}
