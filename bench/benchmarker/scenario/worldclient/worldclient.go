package worldclient

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/guregu/null/v5"
	"github.com/samber/lo"

	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp/api"
	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/world"
)

type userClient struct {
	ctx    context.Context
	client *webapp.Client
}

type providerClient struct {
	ctx                context.Context
	client             *webapp.Client
	webappClientConfig webapp.ClientConfig
}

type chairClient struct {
	ctx    context.Context
	client *webapp.Client
}

type WorldClient struct {
	ctx                context.Context
	webappClientConfig webapp.ClientConfig
}

func NewWorldClient(ctx context.Context, webappClientConfig webapp.ClientConfig) *WorldClient {
	return &WorldClient{
		ctx:                ctx,
		webappClientConfig: webappClientConfig,
	}
}

func (c *WorldClient) RegisterUser(ctx *world.Context, data *world.RegisterUserRequest) (*world.RegisterUserResponse, error) {
	client, err := webapp.NewClient(c.webappClientConfig)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToCreateWebappClient, err)
	}

	response, err := client.AppPostRegister(c.ctx, &api.AppPostRegisterReq{
		Username:    data.UserName,
		Firstname:   data.FirstName,
		Lastname:    data.LastName,
		DateOfBirth: data.DateOfBirth,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterUser, err)
	}

	return &world.RegisterUserResponse{
		ServerUserID: response.ID,
		Client: &userClient{
			ctx:    c.ctx,
			client: client,
		},
	}, nil
}

func (c *WorldClient) RegisterProvider(ctx *world.Context, data *world.RegisterProviderRequest) (*world.RegisterProviderResponse, error) {
	client, err := webapp.NewClient(c.webappClientConfig)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToCreateWebappClient, err)
	}

	response, err := client.ProviderPostRegister(c.ctx, &api.OwnerPostRegisterReq{
		Name: data.Name,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterProvider, err)
	}

	return &world.RegisterProviderResponse{
		ServerProviderID:     response.ID,
		ChairRegisteredToken: response.ChairRegisterToken,
		Client: &providerClient{
			ctx:                c.ctx,
			client:             client,
			webappClientConfig: c.webappClientConfig,
		},
	}, nil
}

func (c *WorldClient) RegisterChair(ctx *world.Context, provider *world.Provider, data *world.RegisterChairRequest) (*world.RegisterChairResponse, error) {
	client, err := webapp.NewClient(c.webappClientConfig)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToCreateWebappClient, err)
	}

	response, err := client.ChairPostRegister(c.ctx, &api.ChairPostRegisterReq{
		Name:               data.Name,
		Model:              data.Model,
		ChairRegisterToken: provider.RegisteredData.ChairRegisterToken,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToRegisterChair, err)
	}

	return &world.RegisterChairResponse{
		ServerChairID: response.ID,
		ServerOwnerID: response.OwnerID,
		Client: &chairClient{
			ctx:    c.ctx,
			client: client,
		},
	}, nil
}

func (c *providerClient) GetProviderSales(ctx *world.Context, args *world.GetProviderSalesRequest) (*world.GetProviderSalesResponse, error) {
	params := api.OwnerGetSalesParams{}
	if !args.Since.IsZero() {
		params.Since.SetTo(args.Since.UnixMilli())
	}
	if !args.Until.IsZero() {
		params.Until.SetTo(args.Until.UnixMilli())
	}

	response, err := c.client.ProviderGetSales(c.ctx, &params)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToGetProviderSales, err)
	}

	return &world.GetProviderSalesResponse{
		Total: response.TotalSales,
		Chairs: lo.Map(response.Chairs, func(v api.OwnerGetSalesOKChairsItem, _ int) *world.ChairSales {
			return &world.ChairSales{
				ID:    v.ID,
				Name:  v.Name,
				Sales: v.Sales,
			}
		}),
		Models: lo.Map(response.Models, func(v api.OwnerGetSalesOKModelsItem, _ int) *world.ChairSalesPerModel {
			return &world.ChairSalesPerModel{
				Model: v.Model,
				Sales: v.Sales,
			}
		}),
	}, nil
}

func (c *providerClient) GetProviderChairs(ctx *world.Context, args *world.GetProviderChairsRequest) (*world.GetProviderChairsResponse, error) {
	response, err := c.client.ProviderGetChairs(c.ctx)
	if err != nil {
		return nil, err
	}

	return &world.GetProviderChairsResponse{Chairs: lo.Map(response.Chairs, func(v api.OwnerGetChairsOKChairsItem, _ int) *world.ProviderChair {
		return &world.ProviderChair{
			ID:                     v.ID,
			Name:                   v.Name,
			Model:                  v.Model,
			Active:                 v.Active,
			RegisteredAt:           time.UnixMilli(v.RegisteredAt),
			TotalDistance:          v.TotalDistance,
			TotalDistanceUpdatedAt: null.NewTime(time.UnixMilli(v.TotalDistanceUpdatedAt.Value), v.TotalDistanceUpdatedAt.Set),
		}
	})}, nil
}

func (c *chairClient) SendChairCoordinate(ctx *world.Context, chair *world.Chair) (*world.SendChairCoordinateResponse, error) {
	response, err := c.client.ChairPostCoordinate(c.ctx, &api.Coordinate{
		Latitude:  chair.Location.Current().X,
		Longitude: chair.Location.Current().Y,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToPostCoordinate, err)
	}

	return &world.SendChairCoordinateResponse{RecordedAt: time.UnixMilli(response.RecordedAt)}, nil
}

func (c *chairClient) SendAcceptRequest(ctx *world.Context, chair *world.Chair, req *world.Request) error {
	_, err := c.client.ChairPostRequestAccept(c.ctx, req.ServerID)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostAccept, err)
	}

	return nil
}

func (c *chairClient) SendDenyRequest(ctx *world.Context, chair *world.Chair, serverRequestID string) error {
	_, err := c.client.ChairPostRequestDeny(c.ctx, serverRequestID)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostDeny, err)
	}

	return nil
}

func (c *chairClient) SendDepart(ctx *world.Context, req *world.Request) error {
	_, err := c.client.ChairPostRequestDepart(c.ctx, req.ServerID)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostDepart, err)
	}

	return nil
}

func (c *chairClient) SendActivate(ctx *world.Context, chair *world.Chair) error {
	_, err := c.client.ChairPostActivate(c.ctx)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostActivate, err)
	}

	return nil
}

func (c *chairClient) SendDeactivate(ctx *world.Context, chair *world.Chair) error {
	_, err := c.client.ChairPostDeactivate(c.ctx)
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostDeactivate, err)
	}

	return nil
}

func (c *chairClient) GetRequestByChair(ctx *world.Context, chair *world.Chair, serverRequestID string) (*world.GetRequestByChairResponse, error) {
	_, err := c.client.ChairGetRequest(c.ctx, serverRequestID)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToGetChairRequest, err)
	}

	// TODO: GetRequestByChairResponse の中身入れる
	return &world.GetRequestByChairResponse{}, nil
}

func (c *chairClient) ConnectChairNotificationStream(ctx *world.Context, chair *world.Chair, receiver world.NotificationReceiverFunc) (world.NotificationStream, error) {
	sseContext, cancel := context.WithCancel(c.ctx)

	go func() {
		for r, err := range c.client.ChairGetNotification(sseContext) {
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					// TODO ロギング
					slog.Debug(err.Error())
				}
				continue
			}

			var event world.NotificationEvent
			switch r.Status {
			case api.RequestStatusMATCHING:
				event = &world.ChairNotificationEventMatched{
					ServerRequestID: r.RequestID,
				}
			case api.RequestStatusDISPATCHING:
				// event = &world.ChairNotificationEventDispatching{}
			case api.RequestStatusDISPATCHED:
				// event = &world.ChairNotificationEventDispatched{}
			case api.RequestStatusCARRYING:
				// event = &world.ChairNotificationEventCarrying{}
			case api.RequestStatusARRIVED:
				// event = &world.ChairNotificationEventArrived{}
			case api.RequestStatusCOMPLETED:
				event = &world.ChairNotificationEventCompleted{
					ServerRequestID: r.RequestID,
				}
			}
			if event == nil {
				// 意図しない通知の種類は無視する
				continue
			}
			receiver(event)
		}
	}()

	return &notificationConnectionImpl{
		close: cancel,
	}, nil
}

func (c *userClient) SendEvaluation(ctx *world.Context, req *world.Request, score int) (*world.SendEvaluationResponse, error) {
	res, err := c.client.AppPostRequestEvaluate(c.ctx, req.ServerID, &api.AppPostRequestEvaluateReq{
		Evaluation: score,
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToPostEvaluate, err)
	}

	return &world.SendEvaluationResponse{
		Fare:        res.Fare,
		CompletedAt: time.UnixMilli(res.CompletedAt),
	}, nil
}

func (c *userClient) SendCreateRequest(ctx *world.Context, req *world.Request) (*world.SendCreateRequestResponse, error) {
	pickup := req.PickupPoint
	destination := req.DestinationPoint
	response, err := c.client.AppPostRequest(c.ctx, &api.AppPostRequestReq{
		PickupCoordinate: api.Coordinate{
			Latitude:  pickup.X,
			Longitude: pickup.Y,
		},
		DestinationCoordinate: api.Coordinate{
			Latitude:  destination.X,
			Longitude: destination.Y,
		},
	})
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToPostRequest, err)
	}

	return &world.SendCreateRequestResponse{ServerRequestID: response.RequestID}, nil
}

func (c *userClient) RegisterPaymentMethods(ctx *world.Context, user *world.User) error {
	_, err := c.client.AppPostPaymentMethods(c.ctx, &api.AppPostPaymentMethodsReq{Token: user.PaymentToken})
	if err != nil {
		return WrapCodeError(ErrorCodeFailedToPostPaymentMethods, err)
	}
	return nil
}

func (c *userClient) GetRequests(ctx *world.Context) (*world.GetRequestsResponse, error) {
	res, err := c.client.AppGetRequests(c.ctx)
	if err != nil {
		return nil, WrapCodeError(ErrorCodeFailedToGetRequests, err)
	}

	requests := make([]*world.RequestHistory, len(res.Requests))
	for i, r := range res.Requests {
		requests[i] = &world.RequestHistory{
			ID: r.RequestID,
			PickupCoordinate: world.Coordinate{
				X: r.PickupCoordinate.Latitude,
				Y: r.PickupCoordinate.Longitude,
			},
			DestinationCoordinate: world.Coordinate{
				X: r.DestinationCoordinate.Latitude,
				Y: r.DestinationCoordinate.Longitude,
			},
			Chair: world.RequestHistoryChair{
				ID:    r.Chair.ID,
				Owner: r.Chair.Owner,
				Name:  r.Chair.Name,
				Model: r.Chair.Model,
			},
			Fare:        r.Fare,
			Evaluation:  r.Evaluation,
			RequestedAt: time.UnixMilli(r.RequestedAt),
			CompletedAt: time.UnixMilli(r.CompletedAt),
		}
	}

	return &world.GetRequestsResponse{
		Requests: requests,
	}, nil
}

func (c *userClient) ConnectUserNotificationStream(ctx *world.Context, user *world.User, receiver world.NotificationReceiverFunc) (world.NotificationStream, error) {
	sseContext, cancel := context.WithCancel(c.ctx)

	go func() {
		for r, err := range c.client.AppGetNotification(sseContext) {
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					// TODO ロギング
					slog.Debug(err.Error())
				}
				continue
			}

			var event world.NotificationEvent
			switch r.Status {
			case api.RequestStatusMATCHING:
				// event = &world.UserNotificationEventMatching{}
			case api.RequestStatusDISPATCHING:
				event = &world.UserNotificationEventDispatching{
					ServerRequestID: r.RequestID,
				}
			case api.RequestStatusDISPATCHED:
				event = &world.UserNotificationEventDispatched{
					ServerRequestID: r.RequestID,
				}
			case api.RequestStatusCARRYING:
				event = &world.UserNotificationEventCarrying{
					ServerRequestID: r.RequestID,
				}
			case api.RequestStatusARRIVED:
				event = &world.UserNotificationEventArrived{
					ServerRequestID: r.RequestID,
				}
			case api.RequestStatusCOMPLETED:
				event = &world.UserNotificationEventCompleted{
					ServerRequestID: r.RequestID,
				}
			}
			if event == nil {
				// 意図しない通知の種類は無視する
				continue
			}
			receiver(event)
		}
	}()

	return &notificationConnectionImpl{
		close: cancel,
	}, nil
}

type notificationConnectionImpl struct {
	close func()
}

func (c *notificationConnectionImpl) Close() {
	c.close()
}
