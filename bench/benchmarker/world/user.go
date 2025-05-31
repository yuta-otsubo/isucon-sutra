package world

import (
	"fmt"
	"log"
	"math/rand/v2"
	"slices"
	"sync/atomic"

	"github.com/yuta-otsubo/isucon-sutra/bench/internal/concurrent"
)

type UserState int

const (
	UserStateInactive UserState = iota
	UserStatePaymentMethodsNotRegister
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
	// PaymentToken 支払いトークン
	PaymentToken string
	// RequestHistory リクエスト履歴
	RequestHistory []*Request
	// TotalEvaluation 完了したリクエストの平均評価
	TotalEvaluation int
	// NotificationConn 通知ストリームコネクション
	NotificationConn NotificationStream
	// notificationQueue 通知キュー。毎Tickで最初に処理される
	notificationQueue chan NotificationEvent

	// Rand 専用の乱数
	Rand *rand.Rand
	// tickDone 行動が完了しているかどうか
	tickDone atomic.Bool
}

type RegisteredUserData struct {
	UserName    string
	FirstName   string
	LastName    string
	DateOfBirth string
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
	if !u.tickDone.CompareAndSwap(true, false) {
		return nil
	}
	defer func() {
		if !u.tickDone.CompareAndSwap(false, true) {
			panic("2重でUserのTickが終了した")
		}
	}()

	// 通知キューを順番に処理する
	for event := range concurrent.TryIter(u.notificationQueue) {
		err := u.HandleNotification(event)
		if err != nil {
			return err
		}
	}

	switch {
	// 支払いトークンが未登録
	case u.State == UserStatePaymentMethodsNotRegister:
		// トークン登録を試みる
		err := ctx.client.RegisterPaymentMethods(ctx, u)
		if err != nil {
			return WrapCodeError(ErrorCodeFailedToRegisterPaymentMethods, err)
		}

		// 成功したのでアクティブ状態にする
		u.State = UserStateActive

	// 進行中のリクエストが存在
	case u.Request != nil:
		switch u.Request.Statuses.User {
		case RequestStatusMatching:
			// マッチングされるまで待機する
			// 一向にマッチングされない場合は、このユーザーの行動はハングする
			break

		case RequestStatusDispatching:
			// 椅子が到着するまで待つ
			// 一向に到着しない場合は、このユーザーの行動はハングする
			break

		case RequestStatusDispatched:
			// 椅子が出発するのを待つ
			// 一向に到着しない場合は、このユーザーの行動はハングする
			break

		case RequestStatusCarrying:
			// 椅子が到着するのを待つ
			// 一向に到着しない場合は、このユーザーの行動はハングする
			break

		case RequestStatusArrived:
			// 送迎の評価及び支払いがまだの場合は行う
			if !u.Request.Evaluated {
				score := u.Request.CalculateEvaluation().Score()
				err := ctx.client.SendEvaluation(ctx, u.Request, score)
				if err != nil {
					return WrapCodeError(ErrorCodeFailedToEvaluate, err)
				}

				// サーバーが評価を受理したので完了状態になるのを待機する
				u.Request.CompletedAt = ctx.world.Time
				u.Request.Statuses.Desired = RequestStatusCompleted
				u.Request.Evaluated = true
				if requests := len(u.RequestHistory); requests == 1 {
					u.Region.TotalEvaluation.Add(int32(score))
				} else {
					u.Region.TotalEvaluation.Add(int32((u.TotalEvaluation+score)/requests - u.TotalEvaluation/(requests-1)))
				}
				u.TotalEvaluation += score
				u.Request.Chair.Provider.TotalSales.Add(int64(u.Request.Fare()))
				ctx.world.CompletedRequestChan <- u.Request
			}

		case RequestStatusCompleted:
			// 進行中のリクエストが無い状態にする
			u.Request = nil

		case RequestStatusCanceled:
			// サーバー側でリクエストがキャンセルされた
			u.Request.Statuses.Desired = RequestStatusCanceled

			// 進行中のリクエストが無い状態にする
			u.Request = nil
			return CodeError(ErrorCodeRequestCanceledByServer)
		}

	// 進行中のリクエストは存在しないが、ユーザーがアクティブ状態
	case u.Request == nil && u.State == UserStateActive:
		if u.NotificationConn == nil {
			// 通知コネクションが無い場合は繋いでおく
			conn, err := ctx.client.ConnectUserNotificationStream(ctx, u, func(event NotificationEvent) {
				if !concurrent.TrySend(u.notificationQueue, event) {
					log.Printf("通知受け取りチャンネルが詰まってる: user server id: %s", u.ServerID)
					u.notificationQueue <- event
				}
			})
			if err != nil {
				return WrapCodeError(ErrorCodeFailedToConnectNotificationStream, err)
			}
			u.NotificationConn = conn
		}

		if count := len(u.RequestHistory); (count == 1 && u.TotalEvaluation <= 1) || float64(u.TotalEvaluation)/float64(count) <= 2 {
			// 初回利用で評価1なら離脱
			// 2回以上利用して平均評価が2以下の場合は離脱
			u.State = UserStateInactive
			u.NotificationConn.Close()
			u.NotificationConn = nil
			break
		}

		// リクエストを作成する
		// TODO 作成する条件・頻度
		err := u.CreateRequest(ctx)
		if err != nil {
			return err
		}

	// 離脱ユーザーは何もしない
	case u.State == UserStateInactive:
		break
	}
	return nil
}

func (u *User) CreateRequest(ctx *Context) error {
	if u.Request != nil {
		panic("ユーザーに進行中のリクエストがあるのにも関わらず、リクエストを新規作成しようとしている")
	}

	// TODO 目的地の決定方法をランダムじゃなくする
	pickup := RandomCoordinateOnRegionWithRand(u.Region, u.Rand)
	dest := RandomCoordinateAwayFromHereWithRand(pickup, u.Rand.IntN(100)+5, u.Rand)

	req := &Request{
		User:             u,
		PickupPoint:      pickup,
		DestinationPoint: dest,
		RequestedAt:      ctx.world.Time,
		Statuses: RequestStatuses{
			Desired: RequestStatusMatching,
			Chair:   RequestStatusMatching,
			User:    RequestStatusMatching,
		},
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

func (u *User) ChangeRequestStatus(status RequestStatus, serverRequestID string) error {
	request := u.Request
	if request == nil {
		return WrapCodeError(ErrorCodeUserNotRequestingButStatusChanged, fmt.Errorf("user server id: %s, got: %v", u.ServerID, status))
	}
	request.Statuses.RLock()
	defer request.Statuses.RUnlock()
	if status != RequestStatusCanceled && request.Statuses.User != status && request.Statuses.Desired != status {
		// キャンセル以外で、現在認識しているユーザーの状態で無いかつ、想定状態ではない状態に遷移しようとしている場合
		if request.Statuses.User == RequestStatusMatching && request.Statuses.Desired == RequestStatusDispatched && status == RequestStatusDispatching {
			// ユーザーにDispatchingが送られる前に、椅子が到着している場合があるが、その時にDispatchingを受け取ることを許容する
		} else if request.Statuses.User == RequestStatusDispatched && request.Statuses.Desired == RequestStatusArrived && status == RequestStatusCarrying {
			// もう到着しているが、ユーザー側の通知が遅延していて、DISPATCHED状態からまだCARRYINGに遷移してないときは、CARRYINGを許容する
		} else if request.Statuses.Desired == RequestStatusDispatched && request.Statuses.User == RequestStatusDispatched && status == RequestStatusCarrying {
			// FIXME: 出発リクエストを送った後、ベンチマーカーのDesiredステータスの変更を行う前に通知が届いてしまうことがある
		} else if status == RequestStatusCompleted {
			// 履歴を見て、過去扱っていたRequestに向けてのCOMPLETED通知であれば無視する
			for _, r := range slices.Backward(u.RequestHistory) {
				if r.ServerID == serverRequestID && r.Statuses.Desired == RequestStatusCompleted {
					r.Statuses.User = RequestStatusCompleted
					return nil
				}
			}
			return WrapCodeError(ErrorCodeUnexpectedUserRequestStatusTransitionOccurred, fmt.Errorf("request server id: %v, expect: %v, got: %v (current: %v)", request.ServerID, request.Statuses.Desired, status, request.Statuses.User))
		} else {
			return WrapCodeError(ErrorCodeUnexpectedUserRequestStatusTransitionOccurred, fmt.Errorf("request server id: %v, expect: %v, got: %v (current: %v)", request.ServerID, request.Statuses.Desired, status, request.Statuses.User))
		}
	}
	request.Statuses.User = status
	return nil
}

func (u *User) HandleNotification(event NotificationEvent) error {
	switch data := event.(type) {
	case *UserNotificationEventDispatching:
		err := u.ChangeRequestStatus(RequestStatusDispatching, data.ServerRequestID)
		if err != nil {
			return err
		}
	case *UserNotificationEventDispatched:
		err := u.ChangeRequestStatus(RequestStatusDispatched, data.ServerRequestID)
		if err != nil {
			return err
		}
	case *UserNotificationEventCarrying:
		err := u.ChangeRequestStatus(RequestStatusCarrying, data.ServerRequestID)
		if err != nil {
			return err
		}
	case *UserNotificationEventArrived:
		err := u.ChangeRequestStatus(RequestStatusArrived, data.ServerRequestID)
		if err != nil {
			return err
		}
	case *UserNotificationEventCompleted:
		err := u.ChangeRequestStatus(RequestStatusCompleted, data.ServerRequestID)
		if err != nil {
			return err
		}
	case *UserNotificationEventCanceled:
		err := u.ChangeRequestStatus(RequestStatusCanceled, data.ServerRequestID)
		if err != nil {
			return err
		}
	}
	return nil
}
