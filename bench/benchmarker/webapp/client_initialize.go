package webapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yuta-otsubo/isucon-sutra/bench/benchmarker/webapp/api"
	"go.uber.org/zap"
)

type PostInitializeResponse struct {
	Language string `json:"language"`
}

func (c *Client) PostInitialize(ctx context.Context, reqBody *api.PostInitializeReq) (*PostInitializeResponse, error) {
	reqBodyBuf, err := reqBody.MarshalJSON()
	if err != nil {
		return nil, err
	}

	req, err := c.agent.NewRequest(http.MethodPost, "/api/initialize", bytes.NewReader(reqBodyBuf))
	if err != nil {
		return nil, err
	}

	resp, err := c.agent.Do(ctx, req)
	if err != nil {
		c.contestantLogger.Warn("POST /api/initialize のリクエストが失敗しました", zap.Error(err))
		return nil, fmt.Errorf("POST /api/initialize のリクエストが失敗しました: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("POST /api/initialize へのリクエストに対して、期待されたHTTPステータスコードが確認できませんでした (expected:%d, actual:%d)", http.StatusOK, resp.StatusCode)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	var response PostInitializeResponse
	if json.NewDecoder(resp.Body).Decode(&response) != nil {
		return nil, fmt.Errorf("initializeのJSONのdecodeに失敗しました: %w", err)
	}

	return &response, nil
}
