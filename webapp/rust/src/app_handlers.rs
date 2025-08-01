use axum::extract::{Path, Query, State};
use axum::http::StatusCode;
use axum::response::IntoResponse as _;
use axum_extra::extract::CookieJar;
use ulid::Ulid;

use crate::models::{Chair, ChairLocation, Coupon, Owner, PaymentToken, Ride, RideStatus, User};
use crate::{AppState, Coordinate, Error};

pub fn app_routes(app_state: AppState) -> axum::Router<AppState> {
    let routes = axum::Router::new().route("/api/app/users", axum::routing::post(app_post_users));

    let authed_routes = axum::Router::new()
        .route(
            "/api/app/payment-methods",
            axum::routing::post(app_post_payment_methods),
        )
        .route(
            "/api/app/rides",
            axum::routing::get(app_get_rides).post(app_post_rides),
        )
        .route(
            "/api/app/rides/estimated-fare",
            axum::routing::post(app_post_rides_estimated_fare),
        )
        .route("/api/app/rides/:ride_id", axum::routing::get(app_get_ride))
        .route(
            "/api/app/rides/:ride_id/evaluation",
            axum::routing::post(app_post_ride_evaluation),
        )
        .route(
            "/api/app/notification",
            axum::routing::get(app_get_notification),
        )
        .route(
            "/api/app/nearby-chairs",
            axum::routing::get(app_get_nearby_chairs),
        )
        .route_layer(axum::middleware::from_fn_with_state(
            app_state.clone(),
            crate::middlewares::app_auth_middleware,
        ));

    routes.merge(authed_routes)
}

#[derive(Debug, serde::Deserialize)]
struct AppPostUsersRequest {
    username: String,
    firstname: String,
    lastname: String,
    date_of_birth: String,
    invitation_code: Option<String>,
}

#[derive(Debug, serde::Serialize)]
struct AppPostUsersResponse {
    id: String,
    invitation_code: String,
}

async fn app_post_users(
    State(AppState { pool, .. }): State<AppState>,
    jar: CookieJar,
    axum::Json(req): axum::Json<AppPostUsersRequest>,
) -> Result<(CookieJar, (StatusCode, axum::Json<AppPostUsersResponse>)), Error> {
    let user_id = Ulid::new().to_string();
    let access_token = crate::secure_random_str(32);
    let invitation_code = crate::secure_random_str(15);

    let mut tx = pool.begin().await?;

    sqlx::query("INSERT INTO users (id, username, firstname, lastname, date_of_birth, access_token, invitation_code) VALUES (?, ?, ?, ?, ?, ?, ?)")
        .bind(&user_id)
        .bind(req.username)
        .bind(req.firstname)
        .bind(req.lastname)
        .bind(req.date_of_birth)
        .bind(&access_token)
        .bind(&invitation_code)
        .execute(&mut *tx)
        .await?;

    // 初回登録キャンペーンのクーポンを付与
    sqlx::query("INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)")
        .bind(&user_id)
        .bind("CP_NEW2024")
        .bind(3000)
        .execute(&mut *tx)
        .await?;

    // 招待コードを使った登録
    if let Some(req_invitation_code) = req.invitation_code {
        if !req_invitation_code.is_empty() {
            // 招待する側の招待数をチェック
            let coupons: Vec<Coupon> =
                sqlx::query_as("SELECT * FROM coupons WHERE code = ? FOR UPDATE")
                    .bind(format!("INV_{req_invitation_code}"))
                    .fetch_all(&mut *tx)
                    .await?;
            if coupons.len() >= 3 {
                return Err(Error::BadRequest("この招待コードは使用できません。"));
            }

            // ユーザーチェック
            let Some(inviter): Option<User> =
                sqlx::query_as("SELECT * FROM users WHERE invitation_code = ?")
                    .bind(&req_invitation_code)
                    .fetch_optional(&mut *tx)
                    .await?
            else {
                return Err(Error::BadRequest("この招待コードは使用できません。"));
            };

            // 招待クーポン付与
            sqlx::query("INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)")
                .bind(&user_id)
                .bind(format!("INV_{req_invitation_code}"))
                .bind(1500)
                .execute(&mut *tx)
                .await?;
            // 招待した人にもRewardを付与
            sqlx::query("INSERT INTO coupons (user_id, code, discount) VALUES (?, ?, ?)")
                .bind(inviter.id)
                .bind(format!("RWD_{req_invitation_code}"))
                .bind(1000)
                .execute(&mut *tx)
                .await?;
        }
    }

    tx.commit().await?;

    let jar = jar
        .add(axum_extra::extract::cookie::Cookie::build(("app_session", access_token)).path("/"));

    Ok((
        jar,
        (
            StatusCode::CREATED,
            axum::Json(AppPostUsersResponse {
                id: user_id,
                invitation_code,
            }),
        ),
    ))
}

#[derive(Debug, serde::Deserialize)]
struct AppPostPaymentMethodsRequest {
    token: String,
}

async fn app_post_payment_methods(
    State(AppState { pool, .. }): State<AppState>,
    axum::Extension(user): axum::Extension<User>,
    axum::Json(req): axum::Json<AppPostPaymentMethodsRequest>,
) -> Result<StatusCode, Error> {
    sqlx::query("INSERT INTO payment_tokens (user_id, token) VALUES (?, ?)")
        .bind(user.id)
        .bind(req.token)
        .execute(&pool)
        .await?;

    Ok(StatusCode::NO_CONTENT)
}

#[derive(Debug, serde::Serialize)]
struct GetAppRidesResponse {
    rides: Vec<GetAppRidesResponseItem>,
}

// Ride
#[derive(Debug, serde::Serialize)]
struct GetAppRidesResponseItem {
    id: String,
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
    chair: GetAppRidesResponseItemChair,
    fare: i32,
    evaluation: i32,
    requested_at: i64,
    completed_at: i64,
}

#[derive(Debug, serde::Serialize)]
struct GetAppRidesResponseItemChair {
    id: String,
    owner: String,
    name: String,
    model: String,
}

async fn app_get_rides(
    State(AppState { pool, .. }): State<AppState>,
    axum::Extension(user): axum::Extension<User>,
) -> Result<axum::Json<GetAppRidesResponse>, Error> {
    let rides: Vec<Ride> =
        sqlx::query_as("SELECT * FROM rides WHERE user_id = ? ORDER BY created_at DESC")
            .bind(user.id)
            .fetch_all(&pool)
            .await?;

    let mut items = Vec::with_capacity(rides.len());
    for ride in rides {
        let status = crate::get_latest_ride_status(&pool, &ride.id).await?;
        if status != "COMPLETED" {
            continue;
        }

        let chair: Chair = sqlx::query_as("SELECT * FROM chairs WHERE id = ?")
            .bind(&ride.chair_id)
            .fetch_one(&pool)
            .await?;

        let owner: Owner = sqlx::query_as("SELECT * FROM owners WHERE id = ?")
            .bind(chair.owner_id)
            .fetch_one(&pool)
            .await?;

        let fare = crate::calculate_sale(&ride);
        items.push(GetAppRidesResponseItem {
            id: ride.id,
            pickup_coordinate: Coordinate {
                latitude: ride.pickup_latitude,
                longitude: ride.pickup_longitude,
            },
            destination_coordinate: Coordinate {
                latitude: ride.destination_latitude,
                longitude: ride.destination_longitude,
            },
            chair: GetAppRidesResponseItemChair {
                id: chair.id,
                owner: owner.name,
                name: chair.name,
                model: chair.model,
            },
            fare,
            evaluation: ride.evaluation.unwrap(),
            requested_at: ride.created_at.timestamp_millis(),
            completed_at: ride.updated_at.timestamp_millis(),
        });
    }

    Ok(axum::Json(GetAppRidesResponse { rides: items }))
}

#[derive(Debug, serde::Deserialize)]
struct AppPostRidesRequest {
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
}

#[derive(Debug, serde::Serialize)]
struct AppPostRidesResponse {
    ride_id: String,
    fare: i32,
}

async fn app_post_rides(
    State(AppState { pool, .. }): State<AppState>,
    axum::Extension(user): axum::Extension<User>,
    axum::Json(req): axum::Json<AppPostRidesRequest>,
) -> Result<(StatusCode, axum::Json<AppPostRidesResponse>), Error> {
    let ride_id = Ulid::new().to_string();

    let mut tx = pool.begin().await?;

    let rides: Vec<Ride> = sqlx::query_as("SELECT * FROM rides WHERE user_id = ?")
        .bind(&user.id)
        .fetch_all(&mut *tx)
        .await?;

    let mut continuing_ride_count = 0;
    for ride in rides {
        let status = crate::get_latest_ride_status(&mut *tx, &ride.id).await?;
        if status != "COMPLETED" && status != "CANCELED" {
            continuing_ride_count += 1;
        }
    }

    if continuing_ride_count > 0 {
        return Err(Error::Conflict("ride already exists"));
    }

    sqlx::query("INSERT INTO rides (id, user_id, pickup_latitude, pickup_longitude, destination_latitude, destination_longitude) VALUES (?, ?, ?, ?, ?, ?)")
        .bind(&ride_id)
        .bind(&user.id)
        .bind(req.pickup_coordinate.latitude)
        .bind(req.pickup_coordinate.longitude)
        .bind(req.destination_coordinate.latitude)
        .bind(req.destination_coordinate.longitude)
        .execute(&mut *tx)
        .await?;

    sqlx::query("INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)")
        .bind(Ulid::new().to_string())
        .bind(&ride_id)
        .bind("MATCHING")
        .execute(&mut *tx)
        .await?;

    let ride_count: i64 = sqlx::query_scalar("SELECT COUNT(*) FROM rides WHERE user_id = ?")
        .bind(&user.id)
        .fetch_one(&mut *tx)
        .await?;

    if ride_count == 1 {
        // 初回利用で、初回利用クーポンがあれば必ず使う
        let coupon: Option<Coupon> = sqlx::query_as("SELECT * FROM coupons WHERE user_id = ? AND code = 'CP_NEW2024' AND used_by IS NULL FOR UPDATE")
            .bind(&user.id)
            .fetch_optional(&mut *tx)
            .await?;
        if coupon.is_some() {
            sqlx::query("UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = 'CP_NEW2024'")
                .bind(&ride_id)
                .bind(&user.id)
                .execute(&mut *tx)
                .await?;
        } else {
            // 無ければ他のクーポンを付与された順番に使う
            let coupon: Option<Coupon> = sqlx::query_as("SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1 FOR UPDATE")
                .bind(&user.id)
                .fetch_optional(&mut *tx)
                .await?;
            if let Some(coupon) = coupon {
                sqlx::query("UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = ?")
                    .bind(&ride_id)
                    .bind(&user.id)
                    .bind(coupon.code)
                    .execute(&mut *tx)
                    .await?;
            }
        }
    } else {
        // 他のクーポンを付与された順番に使う
        let coupon: Option<Coupon> = sqlx::query_as("SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1 FOR UPDATE")
                .bind(&user.id)
                .fetch_optional(&mut *tx)
                .await?;
        if let Some(coupon) = coupon {
            sqlx::query("UPDATE coupons SET used_by = ? WHERE user_id = ? AND code = ?")
                .bind(&ride_id)
                .bind(&user.id)
                .bind(coupon.code)
                .execute(&mut *tx)
                .await?;
        }
    }

    let ride: Ride = sqlx::query_as("SELECT * FROM rides WHERE id = ?")
        .bind(&ride_id)
        .fetch_one(&mut *tx)
        .await?;

    let fare = calculate_discounted_fare(
        &mut tx,
        &user.id,
        Some(&ride),
        req.pickup_coordinate.latitude,
        req.pickup_coordinate.longitude,
        req.destination_coordinate.latitude,
        req.destination_coordinate.longitude,
    )
    .await?;

    tx.commit().await?;

    Ok((
        StatusCode::ACCEPTED,
        axum::Json(AppPostRidesResponse { ride_id, fare }),
    ))
}

#[derive(Debug, serde::Deserialize)]
struct AppPostRidesEstimatedFareRequest {
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
}

#[derive(Debug, serde::Serialize)]
struct AppPostRidesEstimatedFareResponse {
    fare: i32,
    discount: i32,
}

async fn app_post_rides_estimated_fare(
    State(AppState { pool, .. }): State<AppState>,
    axum::Extension(user): axum::Extension<User>,
    axum::Json(req): axum::Json<AppPostRidesEstimatedFareRequest>,
) -> Result<(StatusCode, axum::Json<AppPostRidesEstimatedFareResponse>), Error> {
    let mut tx = pool.begin().await?;

    let discounted = calculate_discounted_fare(
        &mut tx,
        &user.id,
        None,
        req.pickup_coordinate.latitude,
        req.pickup_coordinate.longitude,
        req.destination_coordinate.latitude,
        req.destination_coordinate.longitude,
    )
    .await?;

    tx.commit().await?;

    Ok((
        StatusCode::ACCEPTED,
        axum::Json(AppPostRidesEstimatedFareResponse {
            fare: discounted,
            discount: crate::calculate_fare(
                req.pickup_coordinate.latitude,
                req.pickup_coordinate.longitude,
                req.destination_coordinate.latitude,
                req.destination_coordinate.longitude,
            ),
        }),
    ))
}

#[derive(Debug, serde::Serialize)]
struct RecentRide {
    id: String,
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
    distance: i32,
    duration: i64,
    evaluation: i32,
}

#[derive(Debug, serde::Serialize)]
struct AppChairStats {
    // 最近の乗車履歴
    recent_rides: Vec<RecentRide>,

    // 累計の情報
    total_rides_count: i32,
    total_evaluation_avg: f64,
}

#[derive(Debug, serde::Serialize)]
struct AppChair {
    id: String,
    name: String,
    model: String,
    stats: AppChairStats,
}

#[derive(Debug, serde::Serialize)]
struct AppGetRideResponse {
    id: String,
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
    status: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    chair: Option<AppChair>,
    created_at: i64,
    updated_at: i64,
}

async fn app_get_ride(
    State(AppState { pool, .. }): State<AppState>,
    Path((ride_id,)): Path<(String,)>,
) -> Result<axum::Json<AppGetRideResponse>, Error> {
    let mut tx = pool.begin().await?;

    let Some(ride): Option<Ride> = sqlx::query_as("SELECT * FROM rides WHERE id = ?")
        .bind(ride_id)
        .fetch_optional(&mut *tx)
        .await?
    else {
        return Err(Error::NotFound("ride not found"));
    };

    let status = crate::get_latest_ride_status(&mut *tx, &ride.id).await?;

    let mut response = AppGetRideResponse {
        id: ride.id,
        pickup_coordinate: Coordinate {
            latitude: ride.pickup_latitude,
            longitude: ride.pickup_longitude,
        },
        destination_coordinate: Coordinate {
            latitude: ride.destination_latitude,
            longitude: ride.destination_longitude,
        },
        status,
        chair: None,
        created_at: ride.created_at.timestamp_millis(),
        updated_at: ride.updated_at.timestamp_millis(),
    };

    if let Some(chair_id) = ride.chair_id {
        let chair: Chair = sqlx::query_as("SELECT * FROM chairs WHERE id = ?")
            .bind(chair_id)
            .fetch_one(&mut *tx)
            .await?;
        let stats = get_chair_stats(&mut tx, &chair.id).await?;

        response.chair = Some(AppChair {
            id: chair.id,
            name: chair.name,
            model: chair.model,
            stats,
        });
    }

    Ok(axum::Json(response))
}

async fn get_chair_stats(
    tx: &mut sqlx::MySqlConnection,
    chair_id: &str,
) -> Result<AppChairStats, Error> {
    // 最近の乗車履歴
    let rides: Vec<Ride> =
        sqlx::query_as("SELECT * FROM rides WHERE chair_id = ? ORDER BY updated_at DESC")
            .bind(chair_id)
            .fetch_all(&mut *tx)
            .await?;

    let total_ride_count = rides.len() as i32;
    let mut total_evaluation = 0.0;
    let mut recent_rides = Vec::new();
    for ride in rides {
        let chair_locations: Vec<ChairLocation> = sqlx::query_as("SELECT * FROM chair_locations WHERE chair_id = ? AND created_at BETWEEN ? AND ? ORDER BY created_at")
            .bind(chair_id)
            .bind(ride.created_at)
            .bind(ride.updated_at)
            .fetch_all(&mut *tx)
            .await?;

        let ride_statuses: Vec<RideStatus> =
            sqlx::query_as("SELECT * FROM ride_statuses WHERE ride_id = ? ORDER BY created_at")
                .bind(&ride.id)
                .fetch_all(&mut *tx)
                .await?;

        let Some(arrived_at) = ride_statuses
            .iter()
            .find_map(|status| (status.status == "ARRIVED").then_some(status.created_at))
        else {
            continue;
        };
        let Some(rode_at) = ride_statuses
            .iter()
            .find_map(|status| (status.status == "CARRYING").then_some(status.created_at))
        else {
            continue;
        };
        let is_completed = ride_statuses
            .iter()
            .any(|status| status.status == "COMPLETED");
        if !is_completed {
            continue;
        }

        let mut distance = 0;
        let mut last_latitude = ride.pickup_latitude;
        let mut last_longitude = ride.pickup_longitude;
        for location in chair_locations {
            distance += crate::calculate_distance(
                last_latitude,
                last_longitude,
                location.latitude,
                location.longitude,
            );
            last_latitude = location.latitude;
            last_longitude = location.longitude;
        }
        distance += crate::calculate_distance(
            last_latitude,
            last_longitude,
            ride.destination_latitude,
            ride.destination_longitude,
        );

        recent_rides.push(RecentRide {
            id: ride.id,
            pickup_coordinate: Coordinate {
                latitude: ride.pickup_latitude,
                longitude: ride.pickup_longitude,
            },
            destination_coordinate: Coordinate {
                latitude: ride.destination_latitude,
                longitude: ride.destination_longitude,
            },
            distance,
            duration: (arrived_at - rode_at).num_milliseconds(),
            evaluation: ride.evaluation.unwrap(),
        });

        total_evaluation += ride.evaluation.unwrap() as f64;
    }

    // 5件以上の履歴がある場合は5件までにする
    if total_ride_count > 5 {
        recent_rides.truncate(5);
    }

    let total_evaluation_avg = if total_ride_count > 0 {
        total_evaluation / total_ride_count as f64
    } else {
        0.0
    };

    Ok(AppChairStats {
        recent_rides,
        total_rides_count: total_ride_count,
        total_evaluation_avg,
    })
}

#[derive(Debug, serde::Deserialize)]
struct AppPostRideEvaluationRequest {
    evaluation: i32,
}

#[derive(Debug, serde::Serialize)]
struct AppPostRideEvaluationResponse {
    fare: i32,
    completed_at: i64,
}

async fn app_post_ride_evaluation(
    State(AppState {
        pool, payment_url, ..
    }): State<AppState>,
    Path((ride_id,)): Path<(String,)>,
    axum::Json(req): axum::Json<AppPostRideEvaluationRequest>,
) -> Result<axum::Json<AppPostRideEvaluationResponse>, Error> {
    if req.evaluation < 1 || req.evaluation > 5 {
        return Err(Error::BadRequest("evaluation must be between 1 and 5"));
    }

    let mut tx = pool.begin().await?;

    let Some(ride): Option<Ride> = sqlx::query_as("SELECT * FROM rides WHERE id = ?")
        .bind(&ride_id)
        .fetch_optional(&mut *tx)
        .await?
    else {
        return Err(Error::NotFound("ride not found"));
    };
    let status = crate::get_latest_ride_status(&mut *tx, &ride.id).await?;

    if status != "ARRIVED" {
        return Err(Error::BadRequest("not arrived yet"));
    }

    let result = sqlx::query("UPDATE rides SET evaluation = ? WHERE id = ?")
        .bind(req.evaluation)
        .bind(&ride_id)
        .execute(&mut *tx)
        .await?;
    let count = result.rows_affected();
    if count == 0 {
        return Err(Error::NotFound("ride not found"));
    }

    sqlx::query("INSERT INTO ride_statuses (id, ride_id, status) VALUES (?, ?, ?)")
        .bind(Ulid::new().to_string())
        .bind(&ride_id)
        .bind("COMPLETED")
        .execute(&mut *tx)
        .await?;

    let Some(ride): Option<Ride> = sqlx::query_as("SELECT * FROM rides WHERE id = ?")
        .bind(&ride_id)
        .fetch_optional(&mut *tx)
        .await?
    else {
        return Err(Error::NotFound("ride not found"));
    };

    let Some(payment_token): Option<PaymentToken> =
        sqlx::query_as("SELECT * FROM payment_tokens WHERE user_id = ?")
            .bind(&ride.user_id)
            .fetch_optional(&mut *tx)
            .await?
    else {
        return Err(Error::BadRequest("payment token not registered"));
    };

    let fare = calculate_discounted_fare(
        &mut tx,
        &ride.user_id,
        Some(&ride),
        ride.pickup_latitude,
        ride.pickup_longitude,
        ride.destination_latitude,
        ride.destination_longitude,
    )
    .await?;

    let payment_url = payment_url
        .read()
        .expect("payment_url rwlock is poisoned")
        .clone();

    async fn retrieve_rides_order_by_created_at_asc(
        tx: &mut sqlx::MySqlConnection,
        user_id: &str,
    ) -> Result<Vec<Ride>, Error> {
        sqlx::query_as("SELECT * FROM rides WHERE user_id = ? ORDER BY created_at ASC")
            .bind(user_id)
            .fetch_all(tx)
            .await
            .map_err(Error::Sqlx)
    }

    crate::payment_gateway::request_payment_gateway_post_payment(
        &payment_url,
        &payment_token.token,
        &crate::payment_gateway::PaymentGatewayPostPaymentRequest { amount: fare },
        &mut tx,
        &ride.user_id,
        retrieve_rides_order_by_created_at_asc,
    )
    .await?;

    tx.commit().await?;

    Ok(axum::Json(AppPostRideEvaluationResponse {
        fare,
        completed_at: ride.updated_at.timestamp_millis(),
    }))
}

#[derive(Debug, serde::Serialize)]
struct AppGetNotificationResponse {
    ride_id: String,
    pickup_coordinate: Coordinate,
    destination_coordinate: Coordinate,
    status: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    chair: Option<AppChair>,
    created_at: i64,
    updated_at: i64,
}

async fn app_get_notification(
    State(AppState { pool, .. }): State<AppState>,
    axum::Extension(user): axum::Extension<User>,
) -> Result<axum::response::Response, Error> {
    let mut tx = pool.begin().await?;

    let Some(ride): Option<Ride> =
        sqlx::query_as("SELECT * FROM rides WHERE user_id = ? ORDER BY created_at DESC LIMIT 1")
            .bind(user.id)
            .fetch_optional(&mut *tx)
            .await?
    else {
        return Ok(StatusCode::NO_CONTENT.into_response());
    };

    let status = crate::get_latest_ride_status(&mut *tx, &ride.id).await?;

    let mut response = AppGetNotificationResponse {
        ride_id: ride.id,
        pickup_coordinate: Coordinate {
            latitude: ride.pickup_latitude,
            longitude: ride.pickup_longitude,
        },
        destination_coordinate: Coordinate {
            latitude: ride.destination_latitude,
            longitude: ride.destination_longitude,
        },
        status,
        chair: None,
        created_at: ride.created_at.timestamp_millis(),
        updated_at: ride.updated_at.timestamp_millis(),
    };

    if let Some(chair_id) = ride.chair_id {
        let chair: Chair = sqlx::query_as("SELECT * FROM chairs WHERE id = ?")
            .bind(chair_id)
            .fetch_one(&mut *tx)
            .await?;

        let stats = get_chair_stats(&mut tx, &chair.id).await?;

        response.chair = Some(AppChair {
            id: chair.id,
            name: chair.name,
            model: chair.model,
            stats,
        });
    }

    Ok(axum::Json(response).into_response())
}

#[derive(Debug, serde::Deserialize)]
struct AppGetNearbyChairsQuery {
    latitude: i32,
    longitude: i32,
    distance: Option<i32>,
}

#[derive(Debug, serde::Serialize)]
struct AppGetNearbyChairsResponse {
    chairs: Vec<AppChair>,
    retrieved_at: i64,
}

async fn app_get_nearby_chairs(
    State(AppState { pool, .. }): State<AppState>,
    Query(query): Query<AppGetNearbyChairsQuery>,
) -> Result<axum::Json<AppGetNearbyChairsResponse>, Error> {
    let distance = query.distance.unwrap_or(50);
    let coordinate = Coordinate {
        latitude: query.latitude,
        longitude: query.longitude,
    };

    let mut tx = pool.begin().await?;

    let chairs: Vec<Chair> = sqlx::query_as("SELECT * FROM chairs")
        .fetch_all(&mut *tx)
        .await?;

    let mut nearby_chairs = Vec::new();
    for chair in chairs {
        // 現在進行中のリクエストがある場合はスキップ
        let ride: Option<Ride> = sqlx::query_as(
            "SELECT * FROM rides WHERE chair_id = ? ORDER BY created_at DESC LIMIT 1",
        )
        .bind(&chair.id)
        .fetch_optional(&mut *tx)
        .await?;
        if let Some(ride) = ride {
            let status = crate::get_latest_ride_status(&mut *tx, &ride.id).await?;
            if status != "COMPLETED" {
                continue;
            }
        };

        // 5分以内に更新されている最新の位置情報を取得
        let Some(chair_location): Option<ChairLocation> = sqlx::query_as("SELECT * FROM chair_locations WHERE chair_id = ? AND created_at > DATE_SUB(CURRENT_TIMESTAMP(6), INTERVAL 5 MINUTE) ORDER BY created_at DESC LIMIT 1")
            .bind(&chair.id)
            .fetch_optional(&mut *tx)
            .await? else {
                continue;
            };
        if crate::calculate_distance(
            coordinate.latitude,
            coordinate.longitude,
            chair_location.latitude,
            chair_location.longitude,
        ) <= distance
        {
            let stats = get_chair_stats(&mut tx, &chair.id).await?;

            nearby_chairs.push(AppChair {
                id: chair.id,
                name: chair.name,
                model: chair.model,
                stats,
            });
        }
    }

    let retrieved_at: chrono::DateTime<chrono::Utc> =
        sqlx::query_scalar("SELECT CURRENT_TIMESTAMP(6)")
            .fetch_one(&mut *tx)
            .await?;

    Ok(axum::Json(AppGetNearbyChairsResponse {
        chairs: nearby_chairs,
        retrieved_at: retrieved_at.timestamp(),
    }))
}

async fn calculate_discounted_fare(
    tx: &mut sqlx::MySqlConnection,
    user_id: &str,
    ride: Option<&Ride>,
    mut pickup_latitude: i32,
    mut pickup_longitude: i32,
    mut dest_latitude: i32,
    mut dest_longitude: i32,
) -> sqlx::Result<i32> {
    let discount = if let Some(ride) = ride {
        dest_latitude = ride.destination_latitude;
        dest_longitude = ride.destination_longitude;
        pickup_latitude = ride.pickup_latitude;
        pickup_longitude = ride.pickup_longitude;

        // すでにクーポンが紐づいているならそれの割引額を参照
        let coupon: Option<Coupon> = sqlx::query_as("SELECT * FROM coupons WHERE used_by = ?")
            .bind(&ride.id)
            .fetch_optional(&mut *tx)
            .await?;
        coupon.map(|c| c.discount).unwrap_or(0)
    } else {
        // 初回利用クーポンを最優先で使う
        let coupon: Option<Coupon> = sqlx::query_as(
            "SELECT * FROM coupons WHERE user_id = ? AND code = 'CP_NEW2024' AND used_by IS NULL",
        )
        .bind(user_id)
        .fetch_optional(&mut *tx)
        .await?;
        if let Some(coupon) = coupon {
            coupon.discount
        } else {
            // 無いなら他のクーポンを付与された順番に使う
            let coupon: Option<Coupon> = sqlx::query_as("SELECT * FROM coupons WHERE user_id = ? AND used_by IS NULL ORDER BY created_at LIMIT 1")
                .bind(user_id)
                .fetch_optional(&mut *tx)
                .await?;
            coupon.map(|c| c.discount).unwrap_or(0)
        }
    };

    let metered_fare = crate::FARE_PER_DISTANCE
        * crate::calculate_distance(
            pickup_latitude,
            pickup_longitude,
            dest_latitude,
            dest_longitude,
        );
    let discounted_metered_fare = std::cmp::max(metered_fare - discount, 0);

    Ok(crate::INITIAL_FARE + discounted_metered_fare)
}
