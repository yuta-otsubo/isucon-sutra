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
 * 配車要求ステータス
 *
 * matching: サービス上でマッチング処理を行なっていてドライバーが確定していない
 * dispatching: ドライバーが確定し、乗車位置に向かっている
 * dispatched: ドライバーが乗車位置に到着して、ユーザーの乗車を待機している
 * carrying: ユーザーが乗車し、ドライバーが目的地に向かっている
 * arrived: 目的地に到着した
 * completed: ユーザーの決済・ドライバー評価が完了した
 * canceled: 何らかの理由により途中でキャンセルされた(一定時間待ったがドライバーを割り当てられなかった場合などを想定)
 */
export type RequestStatus =
  | "matching"
  | "dispatching"
  | "carrying"
  | "arrived"
  | "completed"
  | "canceled"
  | "dispatched";

/**
 * 簡易ドライバー情報
 */
export type Driver = {
  /**
   * ドライバーID
   */
  id: string;
  /**
   * ドライバー名
   */
  name: string;
  /**
   * 車種
   */
  car_model: string;
  /**
   * カーナンバー
   */
  car_no: string;
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
 * 問い合わせ内容
 */
export type InquiryContent = {
  /**
   * 問い合わせID
   */
  id: string;
  /**
   * 件名
   */
  subject: string;
  /**
   * 問い合わせ内容
   */
  body: string;
  /**
   * 問い合わせ日時
   */
  created_at: number;
};
