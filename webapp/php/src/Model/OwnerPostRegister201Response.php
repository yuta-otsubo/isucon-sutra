<?php

/**
 * ISURIDE API Specification
 * PHP version 7.4
 *
 * @package IsuRide
 * @author  OpenAPI Generator team
 * @link    https://github.com/openapitools/openapi-generator
 */

/**
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 * The version of the OpenAPI document: 1.0
 * Generated by: https://github.com/openapitools/openapi-generator.git
 */

/**
 * NOTE: This class is auto generated by the openapi generator program.
 * https://github.com/openapitools/openapi-generator
 */
namespace IsuRide\Model;

use IsuRide\BaseModel;

/**
 * OwnerPostRegister201Response
 *
 * @package IsuRide\Model
 * @author  OpenAPI Generator team
 * @link    https://github.com/openapitools/openapi-generator
 */
class OwnerPostRegister201Response extends BaseModel
{
    /**
     * @var string Models namespace.
     * Can be required for data deserialization when model contains referenced schemas.
     */
    protected const MODELS_NAMESPACE = '\IsuRide\Model';

    /**
     * @var string Constant with OAS schema of current class.
     * Should be overwritten by inherited class.
     */
    protected const MODEL_SCHEMA = <<<'SCHEMA'
{
  "required" : [ "chair_register_token", "id" ],
  "type" : "object",
  "properties" : {
    "id" : {
      "type" : "string",
      "description" : "オーナーID"
    },
    "chair_register_token" : {
      "type" : "string",
      "description" : "椅子をオーナーに紐づけるための椅子登録トークン"
    }
  }
}
SCHEMA;
}
