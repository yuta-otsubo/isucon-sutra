package webapp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"strings"
	"time"

	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp/api"
)

func (c *Client) ChairPostRegister(ctx context.Context, reqBody *api.ChairPostChairsReq) (*api.ChairPostChairsCreated, error) {
	reqBodyBuf, err := reqBody.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/chair/chairs", bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("POST /api/chair/chairsのリクエストが失敗しました: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("POST /api/chair/chairsへのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusCreated, resp.StatusCode)
	}

	resBody := &api.ChairPostChairsCreated{}
	if err := json.NewDecoder(resp.Body).Decode(resBody); err != nil {
		return nil, fmt.Errorf("registerのJSONのdecodeに失敗しました: %w", err)
	}

	return resBody, nil
}

func (c *Client) ChairPostActivity(ctx context.Context, reqBody *api.ChairPostActivityReq) (*api.ChairPostActivityNoContent, error) {
	reqBodyBuf, err := reqBody.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/chair/activity", bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("POST /api/chair/activityのリクエストが失敗しました: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("POST /api/chair/activityへのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusNoContent, resp.StatusCode)
	}

	resBody := &api.ChairPostActivityNoContent{}
	return resBody, nil
}

func (c *Client) ChairPostCoordinate(ctx context.Context, reqBody *api.Coordinate) (*api.ChairPostCoordinateOK, error) {
	reqBodyBuf, err := reqBody.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/chair/coordinate", bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("POST /api/chair/coordinateのリクエストが失敗しました: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST /api/chair/coordinateへのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusOK, resp.StatusCode)
	}

	resBody := &api.ChairPostCoordinateOK{}
	if err := json.NewDecoder(resp.Body).Decode(resBody); err != nil {
		return nil, fmt.Errorf("POST /api/chair/coordinateのJSONのdecodeに失敗しました: %w", err)
	}

	return resBody, nil
}

func (c *Client) ChairGetRequest(ctx context.Context, rideID string) (*api.ChairRide, error) {
	req, err := c.agent.NewRequest(http.MethodGet, fmt.Sprintf("/api/chair/rides/%s", rideID), nil)
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("GET /api/chair/rides/{rideID}のリクエストが失敗しました: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET /api/chair/rides/{rideID}へのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusOK, resp.StatusCode)
	}

	resBody := &api.ChairRide{}
	if err := json.NewDecoder(resp.Body).Decode(resBody); err != nil {
		return nil, fmt.Errorf("GET /api/chair/rides/{rideID}のJSONのdecodeに失敗しました: %w", err)
	}

	return resBody, nil
}

func (c *Client) ChairPostRideStatus(ctx context.Context, rideID string, reqBody *api.ChairPostRideStatusReq) (*api.ChairPostRideStatusNoContent, error) {
	reqBodyBuf, err := reqBody.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := c.agent.NewRequest(http.MethodPost, fmt.Sprintf("/api/chair/rides/%s/status", rideID), bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("POST /api/chair/rides/{rideID}/statusのリクエストが失敗しました: %w", err)
	}
	defer closeBody(resp)

	if resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("POST /api/chair/rides/{rideID}/statusへのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusNoContent, resp.StatusCode)
	}

	resBody := &api.ChairPostRideStatusNoContent{}
	return resBody, nil
}

func (c *Client) ChairGetNotification(ctx context.Context) iter.Seq2[*api.ChairGetNotificationOK, error] {
	return func(yield func(*api.ChairGetNotificationOK, error) bool) {
		for notification, err := range c.chairGetNotification(ctx, false) {
			if notification == nil {
				if !yield(nil, err) {
					return
				}
			} else {
				if !yield(&api.ChairGetNotificationOK{
					RideID:                notification.RideID,
					User:                  notification.User,
					PickupCoordinate:      notification.PickupCoordinate,
					DestinationCoordinate: notification.DestinationCoordinate,
					Status:                notification.Status,
				}, err) {
					return
				}
			}
		}
	}
}

func (c *Client) chairGetNotification(ctx context.Context, nested bool) iter.Seq2[*api.ChairGetNotificationOK, error] {
	req, err := c.agent.NewRequest(http.MethodGet, "/api/chair/notification", nil)
	if err != nil {
		return func(yield func(*api.ChairGetNotificationOK, error) bool) { yield(nil, err) }
	}

	for _, modifier := range c.requestModifiers {
		modifier(req)
	}

	httpClient := &http.Client{
		Transport:     c.agent.HttpClient.Transport,
		CheckRedirect: c.agent.HttpClient.CheckRedirect,
		Jar:           c.agent.HttpClient.Jar,
		Timeout:       60 * time.Second,
	}

	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return func(yield func(*api.ChairGetNotificationOK, error) bool) {
			yield(nil, fmt.Errorf("GET /api/chair/notificationのリクエストが失敗しました: %w", err))
		}
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return func(yield func(*api.ChairGetNotificationOK, error) bool) {
			defer closeBody(resp)

			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				request := &api.ChairGetNotificationOK{}
				line := scanner.Text()
				if strings.HasPrefix(line, "data:") {
					err := json.Unmarshal([]byte(line[5:]), request)
					if !yield(request, err) || err != nil {
						return
					}
				}
			}
		}
	}

	defer closeBody(resp)
	request := &api.ChairGetNotificationOK{}
	if resp.StatusCode == http.StatusOK {
		if err = json.NewDecoder(resp.Body).Decode(request); err != nil {
			err = fmt.Errorf("GET /api/chair/notificationのJSONのdecodeに失敗しました: %w", err)
		}
	} else if resp.StatusCode != http.StatusNoContent {
		err = fmt.Errorf("GET /api/chair/notificationへのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d or %d, actual:%d)", http.StatusOK, http.StatusNoContent, resp.StatusCode)
	}

	if nested {
		return func(yield func(*api.ChairGetNotificationOK, error) bool) { yield(request, err) }
	} else {
		return func(yield func(*api.ChairGetNotificationOK, error) bool) {
			if !yield(request, err) {
				return
			}

			const defaultWaitTime = 30 * time.Millisecond
			waitTime := defaultWaitTime
			for {
				select {
				// こちらから切断
				case <-ctx.Done():
					return

				default:
					time.Sleep(waitTime)
					for notification, err := range c.chairGetNotification(ctx, true) {
						if !yield(notification, err) {
							return
						}
						if notification != nil && notification.RetryAfterMs.IsSet() {
							waitTime = time.Duration(notification.RetryAfterMs.Value) * time.Millisecond
						} else {
							waitTime = defaultWaitTime
						}
					}
				}
			}
		}
	}
}