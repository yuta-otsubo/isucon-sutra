<?php
/**
 * RideStatus
 *
 * PHP version 7.4
 *
 * @category Class
 * @package  IsuRide
 * @author   OpenAPI Generator team
 * @link     https://openapi-generator.tech
 */

/**
 * ISURIDE API Specification
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: 1.0
 * Generated by: https://openapi-generator.tech
 * OpenAPI Generator version: 7.2.0
 */

/**
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

namespace IsuRide\Model;
use \IsuRide\ObjectSerializer;

/**
 * RideStatus Class Doc Comment
 *
 * @category Class
 * @description ライドのステータス  MATCHING: サービス上でマッチング処理を行なっていて椅子が確定していない ENROUTE: 椅子が確定し、乗車位置に向かっている PICKUP: 椅子が乗車位置に到着して、ユーザーの乗車を待機している CARRYING: ユーザーが乗車し、椅子が目的地に向かっている ARRIVED: 目的地に到着した COMPLETED: ユーザーの決済・椅子評価が完了した
 * @package  IsuRide
 * @author   OpenAPI Generator team
 * @link     https://openapi-generator.tech
 */
class RideStatus
{
    /**
     * Possible values of this enum
     */
    public const MATCHING = 'MATCHING';

    public const ENROUTE = 'ENROUTE';

    public const PICKUP = 'PICKUP';

    public const CARRYING = 'CARRYING';

    public const ARRIVED = 'ARRIVED';

    public const COMPLETED = 'COMPLETED';

    /**
     * Gets allowable values of the enum
     * @return string[]
     */
    public static function getAllowableEnumValues()
    {
        return [
            self::MATCHING,
            self::ENROUTE,
            self::PICKUP,
            self::CARRYING,
            self::ARRIVED,
            self::COMPLETED
        ];
    }
}


