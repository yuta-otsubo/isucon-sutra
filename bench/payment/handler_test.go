package payment

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestPostPaymentRequest_IsSamePayload(t *testing.T) {
	tests := []struct {
		token  string
		req    PostPaymentRequest
		p      *Payment
		expect bool
	}{
		{
			token:  "t1",
			req:    PostPaymentRequest{Amount: 1000},
			p:      &Payment{Token: "t1", Amount: 1000},
			expect: true,
		},
		{
			token:  "t2",
			req:    PostPaymentRequest{Amount: 1000},
			p:      &Payment{Token: "t1", Amount: 1000},
			expect: false,
		},
		{
			token:  "t1",
			req:    PostPaymentRequest{Amount: 10000},
			p:      &Payment{Token: "t1", Amount: 1000},
			expect: false,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.expect, tt.req.IsSamePayload(tt.token, tt.p))
		})
	}
}

func TestServer_PaymentHandler(t *testing.T) {
	prepare := func(t *testing.T) (*Server, *MockVerifier, *httpexpect.Expect) {
		mockCtrl := gomock.NewController(t)
		verifier := NewMockVerifier(mockCtrl)
		server := NewServer(verifier, 1*time.Millisecond, 1)
		httpServer := httptest.NewServer(server)
		t.Cleanup(httpServer.Close)
		e := httpexpect.Default(t, httpServer.URL)

		return server, verifier, e
	}

	t.Run("冪等性ヘッダーあり", func(t *testing.T) {
		t.Run("キーがサーバーにない", func(t *testing.T) {
			t.Run("Status = StatusSuccess", func(t *testing.T) {
				_, verifier, e := prepare(t)

				token := "token1"
				amount := 1000

				verifier.EXPECT().
					Verify(gomock.Cond(func(x *Payment) bool {
						return x.Token == token && x.Amount == amount
					})).
					Return(StatusSuccess)

				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, "idk1").
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusNoContent)
				e.GET("/payments").
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					Expect().
					Status(http.StatusOK).
					JSON().
					Array().
					IsEqual([]ResponsePayment{{
						Amount: amount,
						Status: StatusSuccess.String(),
					}})
			})
			t.Run("Status = StatusInvalidAmount", func(t *testing.T) {
				_, verifier, e := prepare(t)

				token := "token1"
				amount := 0

				verifier.EXPECT().
					Verify(gomock.Cond(func(x *Payment) bool {
						return x.Token == token && x.Amount == amount
					})).
					Return(StatusInvalidAmount)

				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, "idk1").
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().IsEqual(map[string]string{"message": "決済額が不正です"})
			})
			t.Run("Status = StatusInvalidToken", func(t *testing.T) {
				_, verifier, e := prepare(t)

				token := "token1"
				amount := 1000

				verifier.EXPECT().
					Verify(gomock.Cond(func(x *Payment) bool {
						return x.Token == token && x.Amount == amount
					})).
					Return(StatusInvalidToken)

				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, "idk1").
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().IsEqual(map[string]string{"message": "決済トークンが無効です"})
			})
		})
		t.Run("キーがサーバーにあって、処理済み", func(t *testing.T) {
			t.Run("Status = StatusSuccess", func(t *testing.T) {
				server, _, e := prepare(t)

				idk := "idk1"
				token := "token1"
				amount := 1000

				server.knownKeys.Set(idk, &Payment{
					IdempotencyKey: idk,
					Token:          token,
					Amount:         amount,
					Status:         StatusSuccess,
				})

				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, idk).
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusNoContent)
			})
			t.Run("Status = StatusInvalidAmount", func(t *testing.T) {
				server, _, e := prepare(t)

				idk := "idk1"
				token := "token1"
				amount := 0

				server.knownKeys.Set(idk, &Payment{
					IdempotencyKey: idk,
					Token:          token,
					Amount:         amount,
					Status:         StatusInvalidAmount,
				})

				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, idk).
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().IsEqual(map[string]string{"message": "決済額が不正です"})
			})
			t.Run("Status = StatusInvalidToken", func(t *testing.T) {
				server, _, e := prepare(t)

				idk := "idk1"
				token := "token1"
				amount := 1000

				server.knownKeys.Set(idk, &Payment{
					IdempotencyKey: idk,
					Token:          token,
					Amount:         amount,
					Status:         StatusInvalidToken,
				})
				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, idk).
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusBadRequest).
					JSON().Object().IsEqual(map[string]string{"message": "決済トークンが無効です"})
			})
			t.Run("ペイロード不一致", func(t *testing.T) {
				server, _, e := prepare(t)

				idk := "idk1"
				token := "token1"
				amount := 1000

				server.knownKeys.Set(idk, &Payment{
					IdempotencyKey: idk,
					Token:          token,
					Amount:         10001,
					Status:         StatusSuccess,
				})
				e.POST("/payments").
					WithHeader(IdempotencyKeyHeader, idk).
					WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
					WithJSON(map[string]any{
						"amount": amount,
					}).
					Expect().
					Status(http.StatusUnprocessableEntity).
					JSON().Object().IsEqual(map[string]string{"message": "リクエストペイロードがサーバーに記録されているものと異なります"})
			})
		})
		t.Run("キーがサーバーにあって、処理中", func(t *testing.T) {
			server, _, e := prepare(t)

			idk := "idk1"
			token := "token1"
			amount := 1000

			p := &Payment{
				IdempotencyKey: idk,
				Token:          token,
				Amount:         1000,
				Status:         StatusInitial,
			}
			p.locked.Store(true)

			server.knownKeys.Set(idk, p)
			e.POST("/payments").
				WithHeader(IdempotencyKeyHeader, idk).
				WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
				WithJSON(map[string]any{
					"amount": amount,
				}).
				Expect().
				Status(http.StatusConflict).
				JSON().Object().IsEqual(map[string]string{"message": "既に処理中です"})
		})
	})
	t.Run("冪等性ヘッダーなし", func(t *testing.T) {
		t.Run("Status = StatusSuccess", func(t *testing.T) {
			_, verifier, e := prepare(t)

			token := "token1"
			amount := 1000

			verifier.EXPECT().
				Verify(gomock.Cond(func(x *Payment) bool {
					return x.Token == token && x.Amount == amount
				})).
				Return(StatusSuccess)

			e.POST("/payments").
				WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
				WithJSON(map[string]any{
					"amount": amount,
				}).
				Expect().
				Status(http.StatusNoContent)
		})
		t.Run("Status = StatusInvalidAmount", func(t *testing.T) {
			_, verifier, e := prepare(t)

			token := "token1"
			amount := 0

			verifier.EXPECT().
				Verify(gomock.Cond(func(x *Payment) bool {
					return x.Token == token && x.Amount == amount
				})).
				Return(StatusInvalidAmount)

			e.POST("/payments").
				WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
				WithJSON(map[string]any{
					"amount": amount,
				}).
				Expect().
				Status(http.StatusBadRequest).
				JSON().Object().IsEqual(map[string]string{"message": "決済額が不正です"})
		})
		t.Run("Status = StatusInvalidToken", func(t *testing.T) {
			_, verifier, e := prepare(t)

			token := "token1"
			amount := 1000

			verifier.EXPECT().
				Verify(gomock.Cond(func(x *Payment) bool {
					return x.Token == token && x.Amount == amount
				})).
				Return(StatusInvalidToken)

			e.POST("/payments").
				WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
				WithJSON(map[string]any{
					"amount": amount,
				}).
				Expect().
				Status(http.StatusBadRequest).
				JSON().Object().IsEqual(map[string]string{"message": "決済トークンが無効です"})
		})
	})
}

func TestServer_GetPaymentsHandler(t *testing.T) {
	prepare := func(t *testing.T) (*Server, *MockVerifier, *httpexpect.Expect) {
		mockCtrl := gomock.NewController(t)
		verifier := NewMockVerifier(mockCtrl)
		server := NewServer(verifier, 1*time.Millisecond, 1)
		httpServer := httptest.NewServer(server)
		t.Cleanup(httpServer.Close)
		e := httpexpect.Default(t, httpServer.URL)

		return server, verifier, e
	}

	t.Run("そのトークンによる決済が存在する", func(t *testing.T) {
		_, verifier, e := prepare(t)

		token := "token1"
		token2 := "token2"
		validTokens := []string{token, token2}
		invalidAmount := 0
		initialStatusAmount := 500

		payments := []*Payment{{
			Token:  token,
			Amount: 1000,
			Status: StatusSuccess,
		}, {
			Token:  token,
			Amount: 0,
			Status: StatusInvalidAmount,
		}, {
			Token:  token,
			Amount: 500,
			Status: StatusInitial,
		}, {
			Token:  token2,
			Amount: 1000,
			Status: StatusSuccess,
		}}
		expectedResponse := []ResponsePayment{}
		for _, p := range payments {
			if p.Token != token {
				continue
			}
			expectedResponse = append(expectedResponse, NewResponsePayment(p))
		}

		verifier.EXPECT().
			Verify(gomock.Any()).
			Times(len(payments)).
			DoAndReturn(func(p *Payment) Status {
				isValidToken := false
				for _, t := range validTokens {
					if p.Token == t {
						isValidToken = true
						break
					}
				}
				if !isValidToken {
					return StatusInvalidToken
				}
				switch p.Amount {
				case invalidAmount:
					return StatusInvalidAmount
				case initialStatusAmount:
					return StatusInitial
				}
				return StatusSuccess
			})

		for _, p := range payments {
			status := http.StatusNoContent
			if p.Status != StatusSuccess && p.Status != StatusInitial {
				status = http.StatusBadRequest
			}
			e.POST("/payments").
				WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+p.Token).
				WithJSON(map[string]any{
					"amount": p.Amount,
				}).
				Expect().
				Status(status)
		}

		e.GET("/payments").
			WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			IsEqual(expectedResponse)
	})
	t.Run("そのトークンによる決済が存在しない", func(t *testing.T) {
		_, verifier, e := prepare(t)

		token := "token1"

		p := Payment{
			Token:  "token2",
			Amount: 1000,
			Status: StatusSuccess,
		}

		verifier.EXPECT().
			Verify(gomock.Cond(func(x *Payment) bool {
				return p.Token == x.Token && p.Amount == x.Amount
			})).
			Return(StatusSuccess)

		e.POST("/payments").
			WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+p.Token).
			WithJSON(map[string]any{
				"amount": p.Amount,
			}).
			Expect().
			Status(http.StatusNoContent)

		e.GET("/payments").
			WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			IsEmpty()
	})
	t.Run("決済が存在しない", func(t *testing.T) {
		_, _, e := prepare(t)

		token := "token1"

		e.GET("/payments").
			WithHeader(AuthorizationHeader, AuthorizationHeaderPrefix+token).
			Expect().
			Status(http.StatusOK).
			JSON().
			Array().
			IsEmpty()
	})
}
