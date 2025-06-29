package main

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

type postProviderRegisterRequest struct {
	Name string `json:"name"`
}

type postProviderRegisterResponse struct {
	ID string `json:"id"`
}

func providerPostRegister(w http.ResponseWriter, r *http.Request) {
	req := &postProviderRegisterRequest{}
	if err := bindJSON(r, req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	providerID := ulid.Make().String()

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, errors.New("some of required fields(name) are empty"))
		return
	}

	accessToken := secureRandomStr(32)
	_, err := db.Exec(
		"INSERT INTO providers (id, name, access_token, created_at, updated_at) VALUES (?, ?, ?, isu_now(), isu_now())",
		providerID, req.Name, accessToken,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Path:     "/",
		Name:     "provider_session",
		Value:    accessToken,
		HttpOnly: true,
	})

	writeJSON(w, http.StatusCreated, &postProviderRegisterResponse{
		ID: providerID,
	})
}

type getProviderSalesResponse struct {
	TotalSales int          `json:"total_sales"`
	Chairs     []ChairSales `json:"chairs"`
	Models     []ModelSales `json:"models"`
}

type ChairSales struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Sales int    `json:"sales"`
}

type ModelSales struct {
	Model string `json:"model"`
	Sales int    `json:"sales"`
}

func providerGetSales(w http.ResponseWriter, r *http.Request) {
	provider := r.Context().Value("provider").(*Provider)

	since := time.Time{}
	until := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)
	if r.URL.Query().Get("since") != "" {
		parsed, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("since"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
		}
		since = parsed
	}
	if r.URL.Query().Get("until") != "" {
		parsed, err := time.Parse(time.RFC3339Nano, r.URL.Query().Get("until"))
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
		}
		until = parsed
	}

	chairs := []Chair{}
	if err := db.Select(&chairs, "SELECT * FROM chairs WHERE provider_id = ?", provider.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	res := getProviderSalesResponse{
		TotalSales: 0,
	}

	modelSalesByModel := map[string]int{}

	for _, chair := range chairs {
		reqs := []RideRequest{}
		if err := db.Select(&reqs, "SELECT * FROM ride_requests WHERE chair_id = ? AND status = 'COMPLETED' AND updated_at BETWEEN ? AND ?", chair.ID, since, until); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}

		chairSales := sumSales(reqs)
		res.TotalSales += chairSales

		res.Chairs = append(res.Chairs, ChairSales{
			ID:    chair.ID,
			Name:  chair.Name,
			Sales: chairSales,
		})

		modelSalesByModel[chair.Model] += chairSales
	}

	modelSales := []ModelSales{}
	for model, sales := range modelSalesByModel {
		modelSales = append(modelSales, ModelSales{
			Model: model,
			Sales: sales,
		})
	}

	res.Models = modelSales

	writeJSON(w, http.StatusOK, res)
}

const (
	initialFare     = 500
	farePerDistance = 100
)

func sumSales(requests []RideRequest) int {
	sale := 0
	for _, req := range requests {
		sale += calculateSale(req)
	}
	return sale
}

func calculateSale(req RideRequest) int {
	latDiff := max(req.DestinationLatitude-req.PickupLatitude, req.PickupLatitude-req.DestinationLatitude)
	lonDiff := max(req.DestinationLongitude-req.PickupLongitude, req.PickupLongitude-req.DestinationLongitude)
	return initialFare + farePerDistance*(latDiff+lonDiff)
}

type ChairWithDetail struct {
	ID                     string       `db:"id"`
	ProviderID             string       `db:"provider_id"`
	Name                   string       `db:"name"`
	AccessToken            string       `db:"access_token"`
	Model                  string       `db:"model"`
	IsActive               bool         `db:"is_active"`
	CreatedAt              time.Time    `db:"created_at"`
	UpdatedAt              time.Time    `db:"updated_at"`
	TotalDistance          int          `db:"total_distance"`
	TotalDistanceUpdatedAt sql.NullTime `db:"total_distance_updated_at"`
}

type getProviderChairResponse struct {
	Chairs []providerChair `json:"chairs"`
}

type providerChair struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Model                  string     `json:"model"`
	Active                 bool       `json:"active"`
	RegisteredAt           time.Time  `json:"registered_at"`
	TotalDistance          int        `json:"total_distance"`
	TotalDistanceUpdatedAt *time.Time `json:"total_distance_updated_at"`
}

func providerGetChairs(w http.ResponseWriter, r *http.Request) {
	provider := r.Context().Value("provider").(*Provider)

	chairs := []ChairWithDetail{}
	if err := db.Select(&chairs, `SELECT id,
       provider_id,
       name,
       access_token,
       model,
       is_active,
       created_at,
       updated_at,
       IFNULL(total_distance, 0) AS total_distance,
       total_distance_updated_at
FROM chairs
       LEFT JOIN (SELECT chair_id,
                          SUM(IFNULL(distance, 0)) AS total_distance,
                          MAX(created_at)          AS total_distance_updated_at
                   FROM (SELECT chair_id,
                                created_at,
                                ABS(latitude - LAG(latitude) OVER (PARTITION BY chair_id ORDER BY created_at)) +
                                ABS(longitude - LAG(longitude) OVER (PARTITION BY chair_id ORDER BY created_at)) AS distance
                         FROM chair_locations) tmp
                   GROUP BY chair_id) distance_table ON distance_table.chair_id = chairs.id
WHERE provider_id = ?
`, provider.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	res := getProviderChairResponse{}
	for _, chair := range chairs {
		c := providerChair{
			ID:            chair.ID,
			Name:          chair.Name,
			Model:         chair.Model,
			Active:        chair.IsActive,
			RegisteredAt:  chair.CreatedAt,
			TotalDistance: chair.TotalDistance,
		}
		if chair.TotalDistanceUpdatedAt.Valid {
			c.TotalDistanceUpdatedAt = &chair.TotalDistanceUpdatedAt.Time
		}
		res.Chairs = append(res.Chairs, c)
	}
	writeJSON(w, http.StatusOK, res)
}

type providerChairDetail struct {
	ID                     string     `json:"id"`
	Name                   string     `json:"name"`
	Model                  string     `json:"model"`
	Active                 bool       `json:"active"`
	RegisteredAt           time.Time  `json:"registered_at"`
	TotalDistance          int        `json:"total_distance"`
	TotalDistanceUpdatedAt *time.Time `json:"total_distance_updated_at"`
}

func providerGetChairDetail(w http.ResponseWriter, r *http.Request) {
	provider := r.Context().Value("provider").(*Provider)
	chairID := r.PathValue("chair_id")

	chair := ChairWithDetail{}
	if err := db.Get(&chair, `SELECT id,
       provider_id,
       name,
       access_token,
       model,
       is_active,
       created_at,
       updated_at,
       IFNULL(total_distance, 0) AS total_distance,
       total_distance_updated_at
FROM chairs
       LEFT JOIN (SELECT chair_id,
                          SUM(IFNULL(distance, 0)) AS total_distance,
                          MAX(created_at)          AS total_distance_updated_at
                   FROM (SELECT chair_id,
                                created_at,
                                ABS(latitude - LAG(latitude) OVER (PARTITION BY chair_id ORDER BY created_at)) +
                                ABS(longitude - LAG(longitude) OVER (PARTITION BY chair_id ORDER BY created_at)) AS distance
                         FROM chair_locations) tmp
                   GROUP BY chair_id) distance_table ON distance_table.chair_id = chairs.id
WHERE provider_id = ? AND id = ?`, provider.ID, chairID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, errors.New("chair not found"))
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	resp := providerChairDetail{
		ID:            chair.ID,
		Name:          chair.Name,
		Model:         chair.Model,
		Active:        chair.IsActive,
		RegisteredAt:  chair.CreatedAt,
		TotalDistance: chair.TotalDistance,
	}
	if chair.TotalDistanceUpdatedAt.Valid {
		resp.TotalDistanceUpdatedAt = &chair.TotalDistanceUpdatedAt.Time
	}
	writeJSON(w, http.StatusOK, resp)
}
