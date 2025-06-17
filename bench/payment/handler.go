package payment

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type PostPaymentRequest struct {
	Token  string `json:"token"`
	Amount int    `json:"amount"`
}

func (r *PostPaymentRequest) IsSamePayload(p *Payment) bool {
	return r.Token == p.Token && r.Amount == p.Amount
}

func (s *Server) PaymentHandler(w http.ResponseWriter, r *http.Request) {
	var (
		p          *Payment
		newPayment bool
	)

	idk := r.Header.Get(IdempotencyKeyHeader)
	if len(idk) > 0 {
		p, newPayment = s.knownKeys.GetOrSetDefault(idk, func() *Payment { return NewPayment(idk) })
		if !newPayment && p.locked.Load() {
			writeJSON(w, http.StatusConflict, map[string]string{"message": "既に処理中です"})
			return
		}
	} else {
		p = NewPayment("")
		newPayment = true
	}

	var req PostPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "不正なリクエスト形式です"})
		return
	}
	if !newPayment {
		if !req.IsSamePayload(p) {
			writeJSON(w, http.StatusUnprocessableEntity, map[string]string{"message": "リクエストペイロードがサーバーに記録されているものと異なります"})
			return
		}
		writeResponse(w, p.Status)
		return
	} else {
		p.Token = req.Token
		p.Amount = req.Amount
	}

	// 決済処理
	// キューに入れて完了を待つ(ブロッキング)
	s.queue <- p
	<-p.processChan
	p.locked.Store(false)

	select {
	case <-r.Context().Done():
		// クライアントが既に切断している
		return
	default:
		writeResponse(w, p.Status)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error(err.Error())
	}
}

func writeResponse(w http.ResponseWriter, paymentStatus Status) {
	switch paymentStatus {
	case StatusSuccess:
		w.WriteHeader(http.StatusNoContent)
	case StatusInvalidAmount:
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "決済額が不正です"})
	case StatusInvalidToken:
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "決済トークンが無効です"})
	}
}
