// 管理者向けの問い合わせ関連のHTTPリクエストを処理するハンドラー関数を定義
package main

import (
	"net/http"
	"strconv"
)

type getAdminInquiriesResponse struct {
	Inquiries []struct {
		ID        string `json:"id"`
		Subject   string `json:"subject"`
		CreatedAt int64  `json:"created_at"`
	} `json:"inquiries"`
}

// 管理者向けの問い合わせ一覧を取得
func adminGetInquiries(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	if limitStr == "" {
		limitStr = "20"
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}
	cursor, err := strconv.Atoi(cursorStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var inquiries []Inquiry
	err = db.Select(&inquiries, "SELECT * FROM inquiries ORDER BY id DESC LIMIT ? OFFSET ?", limit, cursor)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := getAdminInquiriesResponse{}
	response.Inquiries = make([]struct {
		ID        string `json:"id"`
		Subject   string `json:"subject"`
		CreatedAt int64  `json:"created_at"`
	}, len(inquiries))

	for i, inquiry := range inquiries {
		response.Inquiries[i].ID = inquiry.ID
		response.Inquiries[i].Subject = inquiry.Subject
		response.Inquiries[i].CreatedAt = inquiry.CreatedAt.Unix()
	}

	// resoponse ではなく inquiries (生データ) を返してしまっている
	respondJSON(w, http.StatusOK, inquiries)
}

type getAdminInquiryResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	CreatedAt int64  `json:"created_at"`
}

// 特定の1件の問い合わせ詳細情報を取得する
func adminGetInquiry(w http.ResponseWriter, r *http.Request) {
	inquiryID := r.URL.Query().Get("inquiry_id")

	inquiry := Inquiry{}
	err := db.Get(&inquiry, "SELECT * FROM inquiries WHERE id = ?", inquiryID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := getAdminInquiryResponse{
		ID:        inquiry.ID,
		UserID:    inquiry.UserID,
		Subject:   inquiry.Subject,
		Body:      inquiry.Body,
		CreatedAt: inquiry.CreatedAt.Unix(),
	}

	respondJSON(w, http.StatusOK, response)
}
