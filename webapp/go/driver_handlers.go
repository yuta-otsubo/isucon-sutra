package main

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

type postDriverRegisterRequest struct {
	Username    string `json:"username"`
	Firstname   string `json:"firstname"`
	Lastname    string `json:"lastname"`
	DateOfBirth string `json:"date_of_birth"`
	CarModel    string `json:"car_model"`
	CarNo       string `json:"car_no"`
}

type postDriverRegisterResponse struct {
	AccessToken string `json:"access_token"`
}

// 配車サービスのドライバー登録処理
func postDriverRegister(w http.ResponseWriter, r *http.Request) {
	req := &postDriverRegisterRequest{}
	if err := bindJSON(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	driverID := ulid.Make().String()

	if req.Username == "" || req.Firstname == "" || req.Lastname == "" || req.DateOfBirth == "" || req.CarModel == "" || req.CarNo == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	accessToken := secureRandomStr(32)
	_, err := db.Exec(
		// FIX:
		"INSERT INTO drivers (id, username, firstname, lastname, date_of_birth, car_model, car_no, is_active, accesstoken) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		driverID, req.Username, req.Firstname, req.Lastname, req.DateOfBirth, req.CarModel, req.CarNo, false, accessToken,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusCreated, &postDriverRegisterResponse{
		AccessToken: accessToken,
	})
}

// API認証を処理するミドルウェア
func driverAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessToken := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "))
		if accessToken == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		driver := &Driver{}
		// FIX:
		err := db.Get(driver, "SELECT * FROM drivers WHERE access_token", accessToken)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "driver", driver)
		next(w, r.WithContext(ctx))
	}
}

// ドライバーが自分の状態をアクティブに設定する
func postDriverActivate(w http.ResponseWriter, r *http.Request) {
	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err := db.Exec("UPDATE drivers SET is_active = ? WHERE id = ?", true, driver.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ドライバーが自分の状態を非アクティブに設定する
func postDriverDeactivate(w http.ResponseWriter, r *http.Request) {
	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err := db.Exec("UPDATE drivers SET is_active = ? WHERE id = ?", false, driver.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ドライバーが自分の現在位置を更新する
func postDriverCoordinate(w http.ResponseWriter, r *http.Request) {
	req := &Coordinate{}
	if err := bindJSON(r, req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err := db.Exec(
		`INSERT INTO driver_locations (driver_id, latitude, longitude) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE latitude = ?, longitude = ?`,
		driver.ID, req.Latitude, req.Longitude, req.Latitude, req.Longitude,
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ドライバー向けのリアルタイム通知機能
func getDriverNotification(w http.ResponseWriter, r *http.Request) {
	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Server Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	// 接続が切れるまで無限ループ
	for {
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusOK)
			return

		default:
			rideRequest := &RideRequest{}
			// FIX: SELECT *
			err := db.Get(rideRequest, "SELECT * FROM ride_requests WHERE driver_id = ? AND status = ?", driver.ID, "DISPATCHING")
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					time.Sleep(1 * time.Second)
					continue
				}
				w.WriteHeader(http.StatusInternalServerError)
			}

			if err := writeSSE(w, "matched", rideRequest); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

type simpleUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getDriverRequestResponse struct {
	RequestID             string     `json:"request_id"`
	User                  simpleUser `json:"user"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Status                string     `json:"status"`
}

// ドライバーが特定の配車リクエストの詳細情報を取得
func getDriverRequest(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")

	rideRequest := &RideRequest{}
	// FIX:
	err := db.Get(rideRequest, "SELECT * FROM ride_requests WHERE id = ?", requestID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	user := &User{}
	// FIX:
	err = db.Get(user, "SELECT * FROM users WHERE id = ?", rideRequest.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondJSON(w, http.StatusOK, &getDriverRequestResponse{
		RequestID: rideRequest.ID,
		User: simpleUser{
			ID:   user.ID,
			Name: user.Firstname + " " + user.Lastname,
		},
		DestinationCoordinate: Coordinate{
			Latitude:  rideRequest.DestinationLatitude,
			Longitude: rideRequest.DestinationLongitude,
		},
		Status: rideRequest.Status,
	})
}

// ドライバーが配車リクエストを受け入れる
func postDriverAccept(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")

	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// TODO: トランザクションを使って排他制御を行う
	rideRequest := &RideRequest{}
	// FIX:
	err := db.Get(rideRequest, "SELECT * FROM ride_requests WHERE id = ?", requestID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if rideRequest.DriverID != driver.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE ride_requests SET status = ? WHERE id = ?", "DISPATCHED", requestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ドライバーが配車リクエストを拒否する
func postDriverDeny(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")

	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rideRequest := &RideRequest{}
	// FIX:
	err := db.Get(rideRequest, "SELECT * FROM ride_requests WHERE id = ?", requestID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if rideRequest.DriverID != driver.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE ride_requests SET driver_id = NULL, matched_at = NULL WHERE id = ?", requestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ドライバーが乗客を乗せて出発したことを報告する
func postDriverDepart(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")

	driver, ok := r.Context().Value("driver").(*Driver)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	rideRequest := &RideRequest{}
	// FIX:
	err := db.Get(rideRequest, "SELECT * FROM ride_requests WHERE id = ?", requestID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if rideRequest.DriverID != driver.ID {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, err = db.Exec("UPDATE ride_requests SET status = ? WHERE id = ?", "CARRYING", requestID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
