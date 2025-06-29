package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

type postAppRegisterRequest struct {
	Username    string `json:"username"`
	FirstName   string `json:"firstname"`
	LastName    string `json:"lastname"`
	DateOfBirth string `json:"date_of_birth"`
}

type postAppRegisterResponse struct {
	ID string `json:"id"`
}

func appPostRegister(w http.ResponseWriter, r *http.Request) {
	req := &postAppRegisterRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	userID := ulid.Make().String()

	if req.Username == "" || req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		writeError(w, http.StatusBadRequest, errors.New("required fields(username, firstname, lastname, date_of_birth) are empty"))
		return
	}
	accessToken := secureRandomStr(32)
	_, err := db.Exec(
		"INSERT INTO users (id, username, firstname, lastname, date_of_birth, access_token, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, isu_now(), isu_now())",
		userID, req.Username, req.FirstName, req.LastName, req.DateOfBirth, accessToken,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "app_session",
		Value:    accessToken,
		HttpOnly: true,
	})

	writeJSON(w, http.StatusCreated, &postAppRegisterResponse{
		ID: userID,
	})
}

type postAppPaymentMethodsRequest struct {
	Token string `json:"token"`
}

func appPostPaymentMethods(w http.ResponseWriter, r *http.Request) {
	req := &postAppPaymentMethodsRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	user := r.Context().Value("user").(*User)

	_, err := db.Exec(
		`INSERT INTO payment_tokens (user_id, token, created_at) VALUES (?, ?, isu_now())`,
		user.ID,
		req.Token,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type postAppRequestsRequest struct {
	PickupCoordinate      *Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate *Coordinate `json:"destination_coordinate"`
}

type postAppRequestsResponse struct {
	RequestID string `json:"request_id"`
}

func appPostRequests(w http.ResponseWriter, r *http.Request) {
	req := &postAppRequestsRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	user := r.Context().Value("user").(*User)

	if req.PickupCoordinate == nil || req.DestinationCoordinate == nil {
		writeError(w, http.StatusBadRequest, errors.New("required fields(pickup_coordinate, destination_coordinate) are empty"))
		return
	}
	requestID := ulid.Make().String()

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	requestCount := 0
	if err := tx.Get(&requestCount, `SELECT COUNT(*) FROM ride_requests WHERE user_id = ? AND status NOT IN ('COMPLETED', 'CANCELED')`, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if requestCount > 0 {
		writeError(w, http.StatusConflict, errors.New("request already exists"))
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO ride_requests (id, user_id, status, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude, requested_at, updated_at)
				  VALUES (?, ?, ?, ?, ?, ?, ?, isu_now(), isu_now())`,
		requestID, user.ID, "MATCHING", req.PickupCoordinate.Latitude, req.PickupCoordinate.Longitude, req.DestinationCoordinate.Latitude, req.DestinationCoordinate.Longitude,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusAccepted, &postAppRequestsResponse{
		RequestID: requestID,
	})
}

type appChair struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Model string `json:"model"`
	Stats appChairStats
}

type appChairStats struct {
	// 最近の乗車履歴
	RecentRides []recentRide `json:"recent_rides"`

	// 累計の情報
	TotalRidesCount    int     `json:"total_rides"`
	TotalEvaluationAvg float64 `json:"total_evaluation_avg"`
}

type recentRide struct {
	ID                    string        `json:"id"`
	PickupCoordinate      Coordinate    `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate    `json:"destination_coordinate"`
	Distance              int           `json:"distance"`
	Duration              time.Duration `json:"duration"`
	Evaluation            int           `json:"evaluation"`
}

type getAppRequestResponse struct {
	RequestID             string     `json:"request_id"`
	PickupCoordinate      Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Status                string     `json:"status"`
	Chair                 appChair   `json:"chair"`
	CreatedAt             int64      `json:"created_at"`
	UpdateAt              int64      `json:"updated_at"`
}

func appGetRequest(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("request_id")

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	rideRequest := &RideRequest{}
	err = tx.Get(
		rideRequest,
		`SELECT * FROM ride_requests WHERE id = ?`,
		requestID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("request not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := &getAppRequestResponse{
		RequestID:             rideRequest.ID,
		PickupCoordinate:      Coordinate{Latitude: rideRequest.PickupLatitude, Longitude: rideRequest.PickupLongitude},
		DestinationCoordinate: Coordinate{Latitude: rideRequest.DestinationLatitude, Longitude: rideRequest.DestinationLongitude},
		Status:                rideRequest.Status,
		CreatedAt:             rideRequest.RequestedAt.Unix(),
		UpdateAt:              rideRequest.UpdatedAt.Unix(),
	}

	if rideRequest.ChairID.Valid {
		chair := &Chair{}
		err := db.Get(
			chair,
			`SELECT * FROM chairs WHERE id = ?`,
			rideRequest.ChairID,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		stats, err := getChairStats(tx, chair.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		response.Chair = appChair{
			ID:    chair.ID,
			Name:  chair.Name,
			Model: chair.Model,
			Stats: stats,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func getChairStats(tx *sqlx.Tx, chairID string) (appChairStats, error) {
	stats := appChairStats{}

	// 最近の乗車履歴
	rideRequests := []RideRequest{}
	err := tx.Select(
		&rideRequests,
		`SELECT *
		 FROM ride_requests
		 WHERE ride_requests.chair_id = ? AND ride_requests.status = 'COMPLETED'
		 ORDER BY ride_requests.updated_at DESC`,
		chairID,
	)
	if err != nil {
		return stats, err
	}

	totalRideCount := len(rideRequests)
	totalEvaluation := 0.0
	for _, rideRequest := range rideRequests {
		chairLocations := []ChairLocation{}
		err := tx.Select(
			&chairLocations,
			`SELECT * FROM chair_locations WHERE chair_id = ? AND created_at BETWEEN ? AND ? ORDER BY created_at`,
			chairID, rideRequest.RequestedAt, rideRequest.UpdatedAt,
		)
		if err != nil {
			return stats, err
		}

		distance := 0
		lastLocation := ChairLocation{
			Latitude:  rideRequest.PickupLatitude,
			Longitude: rideRequest.PickupLongitude,
		}
		for _, location := range chairLocations {
			distance += calculateDistance(lastLocation.Latitude, lastLocation.Longitude, location.Latitude, location.Longitude)
			lastLocation = location
		}
		distance += calculateDistance(lastLocation.Latitude, lastLocation.Longitude, rideRequest.DestinationLatitude, rideRequest.DestinationLongitude)

		stats.RecentRides = append(stats.RecentRides, recentRide{
			ID:                    rideRequest.ID,
			PickupCoordinate:      Coordinate{Latitude: rideRequest.PickupLatitude, Longitude: rideRequest.PickupLongitude},
			DestinationCoordinate: Coordinate{Latitude: rideRequest.DestinationLatitude, Longitude: rideRequest.DestinationLongitude},
			Distance:              distance,
			Duration:              rideRequest.ArrivedAt.Sub(*rideRequest.RodeAt),
			Evaluation:            *rideRequest.Evaluation,
		})

		totalEvaluation += float64(*rideRequest.Evaluation)
	}

	// 5件以上の履歴がある場合は5件までにする
	if totalRideCount > 5 {
		stats.RecentRides = stats.RecentRides[:5]
	}

	stats.TotalRidesCount = totalRideCount
	stats.TotalEvaluationAvg = totalEvaluation / float64(totalRideCount)

	return stats, nil
}

// マンハッタン距離を求める
func calculateDistance(aLatitude, aLongitude, bLatitude, bLongitude int) int {
	return abs(aLatitude-bLatitude) + abs(aLongitude-bLongitude)
}
func abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

type postAppEvaluateRequest struct {
	Evaluation int `json:"evaluation"`
}

type postAppEvaluateResponse struct {
	Fare        int       `json:"fare"`
	CompletedAt time.Time `json:"completed_at"`
}

func appPostRequestEvaluate(w http.ResponseWriter, r *http.Request) {
	requestID := r.PathValue("request_id")
	postAppEvaluateRequest := &postAppEvaluateRequest{}
	if err := bindJSON(r, postAppEvaluateRequest); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	rideRequest := &RideRequest{}
	if err := tx.Get(rideRequest, `SELECT * FROM ride_requests WHERE id = ?`, requestID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("request not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if rideRequest.Status != "ARRIVED" {
		writeError(w, http.StatusBadRequest, errors.New("not arrived yet"))
		return
	}

	result, err := tx.Exec(
		`UPDATE ride_requests SET evaluation = ?, status = ?, updated_at = isu_now() WHERE id = ?`,
		postAppEvaluateRequest.Evaluation, "COMPLETED", requestID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if count, err := result.RowsAffected(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	} else if count == 0 {
		writeError(w, http.StatusNotFound, errors.New("request not found"))
		return
	}

	if err := tx.Get(rideRequest, `SELECT * FROM ride_requests WHERE id = ?`, requestID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("request not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	paymentToken := &PaymentToken{}
	if err := tx.Get(paymentToken, `SELECT * FROM payment_tokens WHERE user_id = ?`, rideRequest.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, errors.New("payment token not registered"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	paymentGatewayRequest := &paymentGatewayPostPaymentRequest{
		Token:  paymentToken.Token,
		Amount: calculateSale(*rideRequest),
	}
	if err := requestPaymentGatewayPostPayment(paymentGatewayRequest); err != nil {
		if errors.Is(err, erroredUpstream) {
			writeError(w, http.StatusBadGateway, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, &postAppEvaluateResponse{
		Fare:        calculateSale(*rideRequest),
		CompletedAt: rideRequest.UpdatedAt,
	})
}

func appGetNotification(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)

	rideRequest := &RideRequest{}
	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()
	if err := tx.Get(rideRequest, `SELECT * FROM ride_requests WHERE user_id = ? ORDER BY requested_at DESC LIMIT 1`, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := &getAppRequestResponse{
		RequestID: rideRequest.ID,
		PickupCoordinate: Coordinate{
			Latitude:  rideRequest.PickupLatitude,
			Longitude: rideRequest.PickupLongitude,
		},
		DestinationCoordinate: Coordinate{
			Latitude:  rideRequest.DestinationLatitude,
			Longitude: rideRequest.DestinationLongitude,
		},
		Status:    rideRequest.Status,
		CreatedAt: rideRequest.RequestedAt.Unix(),
		UpdateAt:  rideRequest.UpdatedAt.Unix(),
	}

	if rideRequest.ChairID.Valid {
		chair := &Chair{}
		if err := tx.Get(chair, `SELECT * FROM chairs WHERE id = ?`, rideRequest.ChairID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		stats, err := getChairStats(tx, chair.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		response.Chair = appChair{
			ID:    chair.ID,
			Name:  chair.Name,
			Model: chair.Model,
			Stats: stats,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func appGetNotificationSSE(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)

	// Server Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")

	var lastRideRequest *RideRequest
	for {
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusOK)
			return

		default:
			rideRequest := &RideRequest{}
			err := db.Get(rideRequest, `SELECT * FROM ride_requests WHERE user_id = ? ORDER BY requested_at DESC LIMIT 1`, user.ID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			if lastRideRequest != nil && rideRequest.ID == lastRideRequest.ID && rideRequest.Status == lastRideRequest.Status {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			chair := &Chair{}
			if rideRequest.ChairID.Valid {
				if err := db.Get(chair, `SELECT * FROM chairs WHERE id = ?`, rideRequest.ChairID); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			}

			if err := writeSSE(w, "matched", &getAppRequestResponse{
				RequestID: rideRequest.ID,
				PickupCoordinate: Coordinate{
					Latitude:  rideRequest.PickupLatitude,
					Longitude: rideRequest.PickupLongitude,
				},
				DestinationCoordinate: Coordinate{
					Latitude:  rideRequest.DestinationLatitude,
					Longitude: rideRequest.DestinationLongitude,
				},
				Status: rideRequest.Status,
				Chair: appChair{
					ID:    chair.ID,
					Name:  chair.Name,
					Model: chair.Model,
				},
				CreatedAt: rideRequest.RequestedAt.Unix(),
				UpdateAt:  rideRequest.UpdatedAt.Unix(),
			}); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			lastRideRequest = rideRequest
		}
	}
}
