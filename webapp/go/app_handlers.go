package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/oklog/ulid/v2"
)

type appPostUsersRequest struct {
	Username       string  `json:"username"`
	FirstName      string  `json:"firstname"`
	LastName       string  `json:"lastname"`
	DateOfBirth    string  `json:"date_of_birth"`
	InvitationCode *string `json:"invitation_code"`
}

type appPostUsersResponse struct {
	ID             string `json:"id"`
	InvitationCode string `json:"invitation_code"`
}

func appPostUsers(w http.ResponseWriter, r *http.Request) {
	req := &appPostUsersRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Username == "" || req.FirstName == "" || req.LastName == "" || req.DateOfBirth == "" {
		writeError(w, http.StatusBadRequest, errors.New("required fields(username, firstname, lastname, date_of_birth) are empty"))
		return
	}

	userID := ulid.Make().String()
	accessToken := secureRandomStr(32)
	invitationCode := secureRandomStr(15)

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		"INSERT INTO users (id, username, firstname, lastname, date_of_birth, access_token, invitation_code) VALUES (?, ?, ?, ?, ?, ?, ?)",
		userID, req.Username, req.FirstName, req.LastName, req.DateOfBirth, accessToken, invitationCode,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// 初回登録キャンペーンのクーポンを付与
	_, err = tx.Exec(
		"INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)",
		userID, "CP_NEW2024", 3000,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// 招待コードを使った登録
	if req.InvitationCode != nil && *req.InvitationCode != "" {
		// 招待する側の招待数をチェック
		var coupons []Coupon
		err = tx.Select(&coupons, "SELECT * FROM coupons WHERE code = ? FOR UPDATE", "INV_"+*req.InvitationCode)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if len(coupons) >= 3 {
			writeError(w, http.StatusBadRequest, errors.New("この招待コードは使用できません。"))
			return
		}

		// ユーザーチェック
		var inviter User
		err = tx.Get(&inviter, "SELECT * FROM users WHERE invitation_code = ?", *req.InvitationCode)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusBadRequest, errors.New("この招待コードは使用できません。"))
				return
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		// 招待クーポン付与
		_, err = tx.Exec(
			"INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)",
			userID, "INV_"+*req.InvitationCode, 1500,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		// 招待した人にもRewardを付与
		_, err = tx.Exec(
			"INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)",
			inviter.ID, "RWD_"+*req.InvitationCode, 1000,
		)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:  "/",
		Name:  "app_session",
		Value: accessToken,
	})

	writeJSON(w, http.StatusCreated, &appPostUsersResponse{
		ID:             userID,
		InvitationCode: invitationCode,
	})
}

type appPostPaymentMethodsRequest struct {
	Token string `json:"token"`
}

func appPostPaymentMethods(w http.ResponseWriter, r *http.Request) {
	req := &appPostPaymentMethodsRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Token == "" {
		writeError(w, http.StatusBadRequest, errors.New("token is required but was empty"))
		return
	}

	user := r.Context().Value("user").(*User)

	_, err := db.Exec(
		`INSERT INTO payment_tokens (user_id, token) VALUES (?, ?)`,
		user.ID,
		req.Token,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type getAppRidesResponse struct {
	Rides []getAppRidesResponseItem `json:"rides"`
}

// Ride
type getAppRidesResponseItem struct {
	ID                    string                       `json:"id"`
	PickupCoordinate      Coordinate                   `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate                   `json:"destination_coordinate"`
	Chair                 getAppRidesResponseItemChair `json:"chair"`
	Fare                  int                          `json:"fare"`
	Evaluation            int                          `json:"evaluation"`
	RequestedAt           int64                        `json:"requested_at"`
	CompletedAt           int64                        `json:"completed_at"`
}

type getAppRidesResponseItemChair struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

func appGetRides(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)

	rides := []Ride{}
	if err := db.Select(
		&rides,
		`SELECT * FROM rides WHERE user_id = ? ORDER BY created_at DESC`,
		user.ID,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	items := []getAppRidesResponseItem{}
	for _, ride := range rides {
		status, err := getLatestRideStatus(db, ride.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if status != "COMPLETED" {
			continue
		}

		item := getAppRidesResponseItem{
			ID:                    ride.ID,
			PickupCoordinate:      Coordinate{Latitude: ride.PickupLatitude, Longitude: ride.PickupLongitude},
			DestinationCoordinate: Coordinate{Latitude: ride.DestinationLatitude, Longitude: ride.DestinationLongitude},
			Fare:                  calculateSale(ride),
			Evaluation:            *ride.Evaluation,
			RequestedAt:           ride.CreatedAt.UnixMilli(),
			CompletedAt:           ride.UpdatedAt.UnixMilli(),
		}

		item.Chair = getAppRidesResponseItemChair{}

		chair := &Chair{}
		if err := db.Get(chair, `SELECT * FROM chairs WHERE id = ?`, ride.ChairID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		item.Chair.ID = chair.ID
		item.Chair.Name = chair.Name
		item.Chair.Model = chair.Model

		owner := &Owner{}
		if err := db.Get(owner, `SELECT * FROM owners WHERE id = ?`, chair.OwnerID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		item.Chair.Owner = owner.Name

		items = append(items, item)
	}

	writeJSON(w, http.StatusOK, &getAppRidesResponse{
		Rides: items,
	})
}

type appPostRidesRequest struct {
	PickupCoordinate      *Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate *Coordinate `json:"destination_coordinate"`
}

type appPostRidesResponse struct {
	RideID string `json:"ride_id"`
	Fare   int    `json:"fare"`
}

type executableGet interface {
	Get(dest interface{}, query string, args ...interface{}) error
}

func getLatestRideStatus(tx executableGet, rideID string) (string, error) {
	status := ""
	if err := tx.Get(&status, `SELECT status FROM ride_statuses WHERE ride_id = ? ORDER BY created_at DESC LIMIT 1`, rideID); err != nil {
		return "", err
	}
	return status, nil
}

func appPostRides(w http.ResponseWriter, r *http.Request) {
	req := &appPostRidesRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.PickupCoordinate == nil || req.DestinationCoordinate == nil {
		writeError(w, http.StatusBadRequest, errors.New("required fields(pickup_coordinate, destination_coordinate) are empty"))
		return
	}

	user := r.Context().Value("user").(*User)
	rideID := ulid.Make().String()

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	rides := []Ride{}
	if err := tx.Select(&rides, `SELECT * FROM rides WHERE user_id = ?`, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	continuingRideCount := 0
	for _, ride := range rides {
		status, err := getLatestRideStatus(tx, ride.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if status != "COMPLETED" && status != "CANCELED" {
			continuingRideCount++
		}
	}

	if continuingRideCount > 0 {
		writeError(w, http.StatusConflict, errors.New("ride already exists"))
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO rides (id, user_id, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude)
				  VALUES (?, ?, ?, ?, ?, ?)`,
		rideID, user.ID, req.PickupCoordinate.Latitude, req.PickupCoordinate.Longitude, req.DestinationCoordinate.Latitude, req.DestinationCoordinate.Longitude,
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if _, err := tx.Exec(
		`INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)`,
		ulid.Make().String(), rideID, "MATCHING",
	); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var rideCount int
	if err := tx.Get(&rideCount, `SELECT COUNT(*) FROM rides WHERE user_id = ? `, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var coupon Coupon
	if rideCount == 1 {
		// 初回利用で、初回利用クーポンがあれば必ず使う
		if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE user_id = ? AND code = 'CP_NEW2024' AND used_by IS NULL FOR UPDATE", user.ID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			// 無ければ他のクーポンを付与された順番に使う
			if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1 FOR UPDATE", user.ID); err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			} else {
				if _, err := tx.Exec(
					"UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = ?",
					rideID, user.ID, coupon.Code,
				); err != nil {
					writeError(w, http.StatusInternalServerError, err)
					return
				}
			}
		} else {
			if _, err := tx.Exec(
				"UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = 'CP_NEW2024'",
				rideID, user.ID,
			); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		}
	} else {
		// 他のクーポンを付与された順番に使う
		if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1 FOR UPDATE", user.ID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		} else {
			if _, err := tx.Exec(
				"UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = ?",
				rideID, user.ID, coupon.Code,
			); err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		}
	}

	ride := Ride{}
	if err := tx.Get(&ride, "SELECT * FROM rides WHERE id = ?", rideID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	fare, err := calculateDiscountedFare(tx, user.ID, &ride, req.PickupCoordinate.Latitude, req.PickupCoordinate.Longitude, req.DestinationCoordinate.Latitude, req.DestinationCoordinate.Longitude)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusAccepted, &appPostRidesResponse{
		RideID: rideID,
		Fare:   fare,
	})
}

type appPostRidesEstimatedFareRequest struct {
	PickupCoordinate      *Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate *Coordinate `json:"destination_coordinate"`
}

type appPostRidesEstimatedFareResponse struct {
	Fare     int `json:"fare"`
	Discount int `json:"discount"`
}

func appPostRidesEstimatedFare(w http.ResponseWriter, r *http.Request) {
	req := &appPostRidesEstimatedFareRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.PickupCoordinate == nil || req.DestinationCoordinate == nil {
		writeError(w, http.StatusBadRequest, errors.New("required fields(pickup_coordinate, destination_coordinate) are empty"))
		return
	}

	user := r.Context().Value("user").(*User)

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	discounted, err := calculateDiscountedFare(tx, user.ID, nil, req.PickupCoordinate.Latitude, req.PickupCoordinate.Longitude, req.DestinationCoordinate.Latitude, req.DestinationCoordinate.Longitude)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusAccepted, &appPostRidesEstimatedFareResponse{
		Fare:     discounted,
		Discount: calculateFare(req.PickupCoordinate.Latitude, req.PickupCoordinate.Longitude, req.DestinationCoordinate.Latitude, req.DestinationCoordinate.Longitude) - discounted,
	})
}

type recentRide struct {
	ID                    string     `json:"id"`
	PickupCoordinate      Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Distance              int        `json:"distance"`
	Duration              int64      `json:"duration"`
	Evaluation            int        `json:"evaluation"`
}

type appChairStats struct {
	// 最近の乗車履歴
	RecentRides []recentRide `json:"recent_rides"`

	// 累計の情報
	TotalRidesCount    int     `json:"total_rides_count"`
	TotalEvaluationAvg float64 `json:"total_evaluation_avg"`
}

type appChair struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Model string        `json:"model"`
	Stats appChairStats `json:"stats"`
}

type appGetRideResponse struct {
	ID                    string     `json:"id"`
	PickupCoordinate      Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Status                string     `json:"status"`
	Chair                 *appChair  `json:"chair,omitempty"`
	CreatedAt             int64      `json:"created_at"`
	UpdateAt              int64      `json:"updated_at"`
}

func appGetRide(w http.ResponseWriter, r *http.Request) {
	rideID := r.PathValue("ride_id")

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	ride := &Ride{}
	err = tx.Get(
		ride,
		`SELECT * FROM rides WHERE id = ?`,
		rideID,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("ride not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := getLatestRideStatus(tx, ride.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := &appGetRideResponse{
		ID:                    ride.ID,
		PickupCoordinate:      Coordinate{Latitude: ride.PickupLatitude, Longitude: ride.PickupLongitude},
		DestinationCoordinate: Coordinate{Latitude: ride.DestinationLatitude, Longitude: ride.DestinationLongitude},
		Status:                status,
		CreatedAt:             ride.CreatedAt.UnixMilli(),
		UpdateAt:              ride.UpdatedAt.UnixMilli(),
	}

	if ride.ChairID.Valid {
		chair := &Chair{}
		err := tx.Get(
			chair,
			`SELECT * FROM chairs WHERE id = ?`,
			ride.ChairID,
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

		response.Chair = &appChair{
			ID:    chair.ID,
			Name:  chair.Name,
			Model: chair.Model,
			Stats: stats,
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func getChairStats(tx *sqlx.Tx, chairID string) (appChairStats, error) {
	stats := appChairStats{RecentRides: make([]recentRide, 0)}

	// 最近の乗車履歴
	rides := []Ride{}
	err := tx.Select(
		&rides,
		`SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC`,
		chairID,
	)
	if err != nil {
		return stats, err
	}

	totalRideCount := len(rides)
	totalEvaluation := 0.0
	for _, ride := range rides {
		chairLocations := []ChairLocation{}
		err := tx.Select(
			&chairLocations,
			`SELECT * FROM chair_locations WHERE chair_id = ? AND created_at BETWEEN ? AND ? ORDER BY created_at`,
			chairID, ride.CreatedAt, ride.UpdatedAt,
		)
		if err != nil {
			return stats, err
		}

		rideStatuses := []RideStatus{}
		err = tx.Select(
			&rideStatuses,
			`SELECT * FROM ride_statuses WHERE ride_id = ? ORDER BY created_at`,
			ride.ID,
		)
		if err != nil {
			return stats, err
		}

		var arrivedAt, rodeAt *time.Time
		var isCompleted bool
		for _, status := range rideStatuses {
			if status.Status == "ARRIVED" {
				arrivedAt = &status.CreatedAt
			} else if status.Status == "CARRYING" {
				rodeAt = &status.CreatedAt
			}
			if status.Status == "COMPLETED" {
				isCompleted = true
			}
		}
		if arrivedAt == nil || rodeAt == nil {
			continue
		}
		if !isCompleted {
			continue
		}

		distance := 0
		lastLocation := ChairLocation{
			Latitude:  ride.PickupLatitude,
			Longitude: ride.PickupLongitude,
		}
		for _, location := range chairLocations {
			distance += calculateDistance(lastLocation.Latitude, lastLocation.Longitude, location.Latitude, location.Longitude)
			lastLocation = location
		}
		distance += calculateDistance(lastLocation.Latitude, lastLocation.Longitude, ride.DestinationLatitude, ride.DestinationLongitude)

		stats.RecentRides = append(stats.RecentRides, recentRide{
			ID:                    ride.ID,
			PickupCoordinate:      Coordinate{Latitude: ride.PickupLatitude, Longitude: ride.PickupLongitude},
			DestinationCoordinate: Coordinate{Latitude: ride.DestinationLatitude, Longitude: ride.DestinationLongitude},
			Distance:              distance,
			Duration:              arrivedAt.Sub(*rodeAt).Milliseconds(),
			Evaluation:            *ride.Evaluation,
		})

		totalEvaluation += float64(*ride.Evaluation)
	}

	// 5件以上の履歴がある場合は5件までにする
	if totalRideCount > 5 {
		stats.RecentRides = stats.RecentRides[:5]
	}

	stats.TotalRidesCount = totalRideCount
	if totalRideCount > 0 {
		stats.TotalEvaluationAvg = totalEvaluation / float64(totalRideCount)
	}

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

type appPostRideEvaluationRequest struct {
	Evaluation int `json:"evaluation"`
}

type appPostRideEvaluationResponse struct {
	Fare        int   `json:"fare"`
	CompletedAt int64 `json:"completed_at"`
}

func appPostRideEvaluatation(w http.ResponseWriter, r *http.Request) {
	rideID := r.PathValue("ride_id")

	req := &appPostRideEvaluationRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if req.Evaluation < 1 || req.Evaluation > 5 {
		writeError(w, http.StatusBadRequest, errors.New("evaluation must be between 1 and 5"))
		return
	}

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	ride := &Ride{}
	if err := tx.Get(ride, `SELECT * FROM rides WHERE id = ?`, rideID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("ride not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	status, err := getLatestRideStatus(tx, ride.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if status != "ARRIVED" {
		writeError(w, http.StatusBadRequest, errors.New("not arrived yet"))
		return
	}

	result, err := tx.Exec(
		`UPDATE rides SET evaluation = ? WHERE id = ?`,
		req.Evaluation, rideID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if count, err := result.RowsAffected(); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	} else if count == 0 {
		writeError(w, http.StatusNotFound, errors.New("ride not found"))
		return
	}

	_, err = tx.Exec(
		`INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)`,
		ulid.Make().String(), rideID, "COMPLETED")
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := tx.Get(ride, `SELECT * FROM rides WHERE id = ?`, rideID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("ride not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	paymentToken := &PaymentToken{}
	if err := tx.Get(paymentToken, `SELECT * FROM payment_tokens WHERE user_id = ?`, ride.UserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, errors.New("payment token not registered"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	fare, err := calculateDiscountedFare(tx, ride.UserID, ride, ride.PickupLatitude, ride.PickupLongitude, ride.DestinationLatitude, ride.DestinationLongitude)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	paymentGatewayRequest := &paymentGatewayPostPaymentRequest{
		Amount: fare,
	}

	var paymentGatewayURL string
	if err := tx.Get(&paymentGatewayURL, "SELECT value FROM settings WHERE name = 'payment_gateway_url'"); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if err := requestPaymentGatewayPostPayment(paymentGatewayURL, paymentToken.Token, paymentGatewayRequest, func() ([]Ride, error) {
		rides := []Ride{}
		if err := tx.Select(&rides, `SELECT * FROM rides WHERE user_id = ? ORDER BY created_at ASC`, ride.UserID); err != nil {
			return nil, err
		}
		return rides, nil
	}); err != nil {
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

	writeJSON(w, http.StatusOK, &appPostRideEvaluationResponse{
		Fare:        fare,
		CompletedAt: ride.UpdatedAt.UnixMilli(),
	})
}

type appGetNotificationResponse struct {
	RideID                string     `json:"ride_id"`
	PickupCoordinate      Coordinate `json:"pickup_coordinate"`
	DestinationCoordinate Coordinate `json:"destination_coordinate"`
	Status                string     `json:"status"`
	Chair                 *appChair  `json:"chair,omitempty"`
	CreatedAt             int64      `json:"created_at"`
	UpdateAt              int64      `json:"updated_at"`
}

func appGetNotification(w http.ResponseWriter, r *http.Request) {
	user := r.Context().Value("user").(*User)

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	ride := &Ride{}
	if err := tx.Get(ride, `SELECT * FROM rides WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	status, err := getLatestRideStatus(tx, ride.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	response := &appGetNotificationResponse{
		RideID: ride.ID,
		PickupCoordinate: Coordinate{
			Latitude:  ride.PickupLatitude,
			Longitude: ride.PickupLongitude,
		},
		DestinationCoordinate: Coordinate{
			Latitude:  ride.DestinationLatitude,
			Longitude: ride.DestinationLongitude,
		},
		Status:    status,
		CreatedAt: ride.CreatedAt.UnixMilli(),
		UpdateAt:  ride.UpdatedAt.UnixMilli(),
	}

	if ride.ChairID.Valid {
		chair := &Chair{}
		if err := tx.Get(chair, `SELECT * FROM chairs WHERE id = ?`, ride.ChairID); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		stats, err := getChairStats(tx, chair.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		response.Chair = &appChair{
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

	var lastRide *Ride
	var lastRideStatus string
	for {
		select {
		case <-r.Context().Done():
			w.WriteHeader(http.StatusOK)
			return

		default:
			err := func() error {
				tx, err := db.Beginx()
				if err != nil {
					return err
				}
				defer tx.Rollback()

				ride := &Ride{}
				err = tx.Get(ride, `SELECT * FROM rides WHERE user_id = ? ORDER BY created_at DESC LIMIT 1`, user.ID)
				if err != nil {
					if errors.Is(err, sql.ErrNoRows) {
						time.Sleep(100 * time.Millisecond)
						return nil
					}
					return err
				}
				status, err := getLatestRideStatus(tx, ride.ID)
				if err != nil {
					return err
				}
				if lastRide != nil && ride.ID == lastRide.ID && status == lastRideStatus {
					time.Sleep(100 * time.Millisecond)
					return nil
				}

				chair := &Chair{}
				stats := appChairStats{}
				if ride.ChairID.Valid {
					if err := tx.Get(chair, `SELECT * FROM chairs WHERE id = ?`, ride.ChairID); err != nil {
						return err
					}
					stats, err = getChairStats(tx, chair.ID)
					if err != nil {
						return err
					}
				}

				if err := writeSSE(w, "matched", &appGetNotificationResponse{
					RideID: ride.ID,
					PickupCoordinate: Coordinate{
						Latitude:  ride.PickupLatitude,
						Longitude: ride.PickupLongitude,
					},
					DestinationCoordinate: Coordinate{
						Latitude:  ride.DestinationLatitude,
						Longitude: ride.DestinationLongitude,
					},
					Status: status,
					Chair: &appChair{
						ID:    chair.ID,
						Name:  chair.Name,
						Model: chair.Model,
						Stats: stats,
					},
					CreatedAt: ride.CreatedAt.UnixMilli(),
					UpdateAt:  ride.UpdatedAt.UnixMilli(),
				}); err != nil {
					return err
				}
				lastRide = ride
				lastRideStatus = status

				return nil
			}()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}

		}
	}
}

type appGetNearbyChairsResponse struct {
	Chairs      []appChair `json:"chairs"`
	RetrievedAt int64      `json:"retrieved_at"`
}

func appGetNearbyChairs(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("latitude")
	lonStr := r.URL.Query().Get("longitude")
	distanceStr := r.URL.Query().Get("distance")
	if latStr == "" || lonStr == "" {
		writeError(w, http.StatusBadRequest, errors.New("latitude or longitude is empty"))
		return
	}

	lat, err := strconv.Atoi(latStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("latitude is invalid"))
		return
	}

	lon, err := strconv.Atoi(lonStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, errors.New("longitude is invalid"))
		return
	}

	distance := 50
	if distanceStr != "" {
		distance, err = strconv.Atoi(distanceStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, errors.New("distance is invalid"))
			return
		}
	}

	coordinate := Coordinate{Latitude: lat, Longitude: lon}

	tx, err := db.Beginx()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	defer tx.Rollback()

	chairs := []Chair{}
	err = tx.Select(
		&chairs,
		`SELECT * FROM chairs`,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	nearbyChairs := []appChair{}
	for _, chair := range chairs {
		// 現在進行中のリクエストがある場合はスキップ
		ride := &Ride{}
		err := tx.Get(
			ride,
			`SELECT * FROM rides WHERE chair_id = ? ORDER BY created_at DESC LIMIT 1`,
			chair.ID,
		)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
		}
		status, err := getLatestRideStatus(tx, ride.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		if status != "COMPLETED" {
			continue
		}

		// 5分以内に更新されている最新の位置情報を取得
		chairLocation := &ChairLocation{}
		err = tx.Get(
			chairLocation,
			`SELECT * FROM chair_locations WHERE chair_id = ? AND created_at > DATE_SUB(CURRENT_TIMESTAMP(6), INTERVAL 5 MINUTE) ORDER BY created_at DESC LIMIT 1`,
			chair.ID,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		if calculateDistance(coordinate.Latitude, coordinate.Longitude, chairLocation.Latitude, chairLocation.Longitude) <= distance {
			stats, err := getChairStats(tx, chair.ID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}

			nearbyChairs = append(nearbyChairs, appChair{
				ID:    chair.ID,
				Name:  chair.Name,
				Model: chair.Model,
				Stats: stats,
			})
		}
	}

	retrievedAt := &time.Time{}
	err = tx.Get(
		retrievedAt,
		`SELECT CURRENT_TIMESTAMP(6)`,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, &appGetNearbyChairsResponse{
		Chairs:      nearbyChairs,
		RetrievedAt: retrievedAt.UnixMilli(),
	})
}

func calculateFare(pickupLatitude, pickupLongitude, destLatitude, destLongitude int) int {
	meteredFare := farePerDistance * calculateDistance(pickupLatitude, pickupLongitude, destLatitude, destLongitude)
	return initialFare + meteredFare
}

func calculateDiscountedFare(tx *sqlx.Tx, userID string, ride *Ride, pickupLatitude, pickupLongitude, destLatitude, destLongitude int) (int, error) {
	var coupon Coupon
	discount := 0
	if ride != nil {
		destLatitude = ride.DestinationLatitude
		destLongitude = ride.DestinationLongitude
		pickupLatitude = ride.PickupLatitude
		pickupLongitude = ride.PickupLongitude

		// すでにクーポンが紐づいているならそれの割引額を参照
		if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE used_by = ?", ride.ID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}
		} else {
			discount = coupon.Discount
		}
	} else {
		// 初回利用クーポンを最優先で使う
		if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE user_id = ? AND code = 'CP_NEW2024' AND used_by IS NULL", userID); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return 0, err
			}

			// 無いなら他のクーポンを付与された順番に使う
			if err := tx.Get(&coupon, "SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1", userID); err != nil {
				if !errors.Is(err, sql.ErrNoRows) {
					return 0, err
				}
			} else {
				discount = coupon.Discount
			}
		} else {
			discount = coupon.Discount
		}
	}

	meteredFare := farePerDistance * calculateDistance(pickupLatitude, pickupLongitude, destLatitude, destLongitude)
	discountedMeteredFare := max(meteredFare-discount, 0)

	return initialFare + discountedMeteredFare, nil
}
