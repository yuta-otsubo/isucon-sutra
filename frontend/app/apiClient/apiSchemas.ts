/**
 * Generated by @openapi-codegen
 *
 * @version 1.0
 */
/**
 * 座標情報
 */
export type Coordinate = {
  /**
   * 経度
   */
  latitude: number;
  /**
   * 緯度
   */
  longitude: number;
};

/**
 * ライドのステータス
 *
 * MATCHING: サービス上でマッチング処理を行なっていて椅子が確定していない
 * ENROUTE: 椅子が確定し、乗車位置に向かっている
 * PICKUP: 椅子が乗車位置に到着して、ユーザーの乗車を待機している
 * CARRYING: ユーザーが乗車し、椅子が目的地に向かっている
 * ARRIVED: 目的地に到着した
 * COMPLETED: ユーザーの決済・椅子評価が完了した
 */
export type RideStatus =
  | "MATCHING"
  | "ENROUTE"
  | "PICKUP"
  | "CARRYING"
  | "ARRIVED"
  | "COMPLETED";

/**
 * App向けの椅子情報
 */
export type AppChair = {
  /**
   * 椅子ID
   */
  id: string;
  /**
   * 椅子の名前
   */
  name: string;
  /**
   * 椅子のモデル
   */
  model: string;
  /**
   * 椅子の統計情報
   */
  stats: {
    /**
     * 最近の乗車情報
     */
    recent_rides: {
      /**
       * ライドID
       */
      id: string;
      pickup_coordinate: Coordinate;
      destination_coordinate: Coordinate;
      /**
       * 移動距離
       */
      distance: number;
      /**
       * 移動時間 (ミリ秒)
       *
       * @format int64
       */
      duration: number;
      /**
       * 評価
       */
      evaluation: number;
    }[];
    /**
     * 総乗車回数
     */
    total_rides_count: number;
    /**
     * 総評価平均
     */
    total_evaluation_avg: number;
  };
};

/**
 * 簡易ユーザー情報
 */
export type User = {
  /**
   * ユーザーID
   */
  id: string;
  /**
   * ユーザー名
   */
  name: string;
};

/**
 * App向けライド情報
 */
export type AppRide = {
  /**
   * ライドID
   */
  id: string;
  pickup_coordinate: Coordinate;
  destination_coordinate: Coordinate;
  status: RideStatus;
  chair?: AppChair;
  /**
   * 配車要求日時
   *
   * @format int64
   */
  created_at: number;
  /**
   * 配車要求更新日時
   *
   * @format int64
   */
  updated_at: number;
};

/**
 * Chair向けライド情報
 */
export type ChairRide = {
  /**
   * ライドID
   */
  id: string;
  user: User;
  pickup_coordinate?: Coordinate;
  destination_coordinate: Coordinate;
  status?: RideStatus;
};

export type Error = {
  message: string;
};
