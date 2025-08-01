// Code generated by ogen, DO NOT EDIT.

package api

import (
	"net/http"
	"net/url"

	"github.com/go-faster/errors"

	"github.com/ogen-go/ogen/conv"
	"github.com/ogen-go/ogen/middleware"
	"github.com/ogen-go/ogen/ogenerrors"
	"github.com/ogen-go/ogen/uri"
	"github.com/ogen-go/ogen/validate"
)

// AppGetNearbyChairsParams is parameters of app-get-nearby-chairs operation.
type AppGetNearbyChairsParams struct {
	// 緯度.
	Latitude int
	// 経度.
	Longitude int
	// 検索距離.
	Distance OptInt
}

func unpackAppGetNearbyChairsParams(packed middleware.Parameters) (params AppGetNearbyChairsParams) {
	{
		key := middleware.ParameterKey{
			Name: "latitude",
			In:   "query",
		}
		params.Latitude = packed[key].(int)
	}
	{
		key := middleware.ParameterKey{
			Name: "longitude",
			In:   "query",
		}
		params.Longitude = packed[key].(int)
	}
	{
		key := middleware.ParameterKey{
			Name: "distance",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Distance = v.(OptInt)
		}
	}
	return params
}

func decodeAppGetNearbyChairsParams(args [0]string, argsEscaped bool, r *http.Request) (params AppGetNearbyChairsParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode query: latitude.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "latitude",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt(val)
				if err != nil {
					return err
				}

				params.Latitude = c
				return nil
			}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "latitude",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: longitude.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "longitude",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToInt(val)
				if err != nil {
					return err
				}

				params.Longitude = c
				return nil
			}); err != nil {
				return err
			}
		} else {
			return err
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "longitude",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: distance.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "distance",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotDistanceVal int
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt(val)
					if err != nil {
						return err
					}

					paramsDotDistanceVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Distance.SetTo(paramsDotDistanceVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "distance",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}

// AppGetRideParams is parameters of app-get-ride operation.
type AppGetRideParams struct {
	// ライドID.
	RideID string
}

func unpackAppGetRideParams(packed middleware.Parameters) (params AppGetRideParams) {
	{
		key := middleware.ParameterKey{
			Name: "ride_id",
			In:   "path",
		}
		params.RideID = packed[key].(string)
	}
	return params
}

func decodeAppGetRideParams(args [1]string, argsEscaped bool, r *http.Request) (params AppGetRideParams, _ error) {
	// Decode path: ride_id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "ride_id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.RideID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "ride_id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// AppPostRideEvaluationParams is parameters of app-post-ride-evaluation operation.
type AppPostRideEvaluationParams struct {
	// ライドID.
	RideID string
}

func unpackAppPostRideEvaluationParams(packed middleware.Parameters) (params AppPostRideEvaluationParams) {
	{
		key := middleware.ParameterKey{
			Name: "ride_id",
			In:   "path",
		}
		params.RideID = packed[key].(string)
	}
	return params
}

func decodeAppPostRideEvaluationParams(args [1]string, argsEscaped bool, r *http.Request) (params AppPostRideEvaluationParams, _ error) {
	// Decode path: ride_id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "ride_id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.RideID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "ride_id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// ChairGetRideParams is parameters of chair-get-ride operation.
type ChairGetRideParams struct {
	// ライドID.
	RideID string
}

func unpackChairGetRideParams(packed middleware.Parameters) (params ChairGetRideParams) {
	{
		key := middleware.ParameterKey{
			Name: "ride_id",
			In:   "path",
		}
		params.RideID = packed[key].(string)
	}
	return params
}

func decodeChairGetRideParams(args [1]string, argsEscaped bool, r *http.Request) (params ChairGetRideParams, _ error) {
	// Decode path: ride_id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "ride_id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.RideID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "ride_id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// ChairPostRideStatusParams is parameters of chair-post-ride-status operation.
type ChairPostRideStatusParams struct {
	// ライドID.
	RideID string
}

func unpackChairPostRideStatusParams(packed middleware.Parameters) (params ChairPostRideStatusParams) {
	{
		key := middleware.ParameterKey{
			Name: "ride_id",
			In:   "path",
		}
		params.RideID = packed[key].(string)
	}
	return params
}

func decodeChairPostRideStatusParams(args [1]string, argsEscaped bool, r *http.Request) (params ChairPostRideStatusParams, _ error) {
	// Decode path: ride_id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "ride_id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.RideID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "ride_id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// OwnerGetChairParams is parameters of owner-get-chair operation.
type OwnerGetChairParams struct {
	// 椅子ID.
	ChairID string
}

func unpackOwnerGetChairParams(packed middleware.Parameters) (params OwnerGetChairParams) {
	{
		key := middleware.ParameterKey{
			Name: "chair_id",
			In:   "path",
		}
		params.ChairID = packed[key].(string)
	}
	return params
}

func decodeOwnerGetChairParams(args [1]string, argsEscaped bool, r *http.Request) (params OwnerGetChairParams, _ error) {
	// Decode path: chair_id.
	if err := func() error {
		param := args[0]
		if argsEscaped {
			unescaped, err := url.PathUnescape(args[0])
			if err != nil {
				return errors.Wrap(err, "unescape path")
			}
			param = unescaped
		}
		if len(param) > 0 {
			d := uri.NewPathDecoder(uri.PathDecoderConfig{
				Param:   "chair_id",
				Value:   param,
				Style:   uri.PathStyleSimple,
				Explode: false,
			})

			if err := func() error {
				val, err := d.DecodeValue()
				if err != nil {
					return err
				}

				c, err := conv.ToString(val)
				if err != nil {
					return err
				}

				params.ChairID = c
				return nil
			}(); err != nil {
				return err
			}
		} else {
			return validate.ErrFieldRequired
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "chair_id",
			In:   "path",
			Err:  err,
		}
	}
	return params, nil
}

// OwnerGetSalesParams is parameters of owner-get-sales operation.
type OwnerGetSalesParams struct {
	// 開始日時（含む）.
	Since OptInt64
	// 終了日時（含む）.
	Until OptInt64
}

func unpackOwnerGetSalesParams(packed middleware.Parameters) (params OwnerGetSalesParams) {
	{
		key := middleware.ParameterKey{
			Name: "since",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Since = v.(OptInt64)
		}
	}
	{
		key := middleware.ParameterKey{
			Name: "until",
			In:   "query",
		}
		if v, ok := packed[key]; ok {
			params.Until = v.(OptInt64)
		}
	}
	return params
}

func decodeOwnerGetSalesParams(args [0]string, argsEscaped bool, r *http.Request) (params OwnerGetSalesParams, _ error) {
	q := uri.NewQueryDecoder(r.URL.Query())
	// Decode query: since.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "since",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotSinceVal int64
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt64(val)
					if err != nil {
						return err
					}

					paramsDotSinceVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Since.SetTo(paramsDotSinceVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "since",
			In:   "query",
			Err:  err,
		}
	}
	// Decode query: until.
	if err := func() error {
		cfg := uri.QueryParameterDecodingConfig{
			Name:    "until",
			Style:   uri.QueryStyleForm,
			Explode: true,
		}

		if err := q.HasParam(cfg); err == nil {
			if err := q.DecodeParam(cfg, func(d uri.Decoder) error {
				var paramsDotUntilVal int64
				if err := func() error {
					val, err := d.DecodeValue()
					if err != nil {
						return err
					}

					c, err := conv.ToInt64(val)
					if err != nil {
						return err
					}

					paramsDotUntilVal = c
					return nil
				}(); err != nil {
					return err
				}
				params.Until.SetTo(paramsDotUntilVal)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}(); err != nil {
		return params, &ogenerrors.DecodeParamError{
			Name: "until",
			In:   "query",
			Err:  err,
		}
	}
	return params, nil
}
