package common

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog/log"
)

type usernameKey struct{}

func Username(ctx context.Context) string {
	if usernamePtr, ok := ctx.Value(usernameKey{}).(*string); ok && usernamePtr != nil {
		return *usernamePtr
	}
	return ""
}

func UsernameMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		username := ""
		req = req.WithContext(context.WithValue(req.Context(), usernameKey{}, &username))
		next.ServeHTTP(w, req)
	})
}

func AddUsernameToRequest(req *http.Request, username string) *http.Request {
	if usernamePtr, ok := req.Context().Value(usernameKey{}).(*string); ok && usernamePtr != nil {
		*usernamePtr = username
	}
	return req
}

type HTTPResponse interface {
	StatusCode() int
}

type ErrorResponse struct {
	HTTPCode  int    `json:"-"`
	ErrorCode string `json:"error"`
	Message   string `json:"message"`
}

func (e ErrorResponse) StatusCode() int {
	return e.HTTPCode
}

func (e ErrorResponse) Error() string {
	return e.Message
}

func WriteResponse(ctx context.Context, w http.ResponseWriter, data interface{}) error {
	content, err := json.Marshal(data)
	if err != nil {
		writeMarshalError(ctx, w, err)
		return err
	}

	header := w.Header()
	header.Add("Content-Type", "application/json")
	header.Add("Content-Length", strconv.Itoa(len(content)))
	if response, ok := data.(HTTPResponse); ok {
		w.WriteHeader(response.StatusCode())
	}
	_, err = w.Write(content)
	return err
}

func writeMarshalError(ctx context.Context, w http.ResponseWriter, marshalErr error) error {
	log.Ctx(ctx).Warn().Err(marshalErr).Msg("failed to marshal json")
	response := ErrorResponse{
		HTTPCode:  500,
		ErrorCode: "ResponseSerializeError",
		Message:   marshalErr.Error(),
	}
	header := w.Header()
	body, err := json.Marshal(response)
	if err == nil {
		header.Add("Content-Type", "application/json")
	} else {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to marshal json")
		header.Add("Content-Type", "text/plain")
		body = []byte(response.Message)
	}
	header.Add("Content-Length", strconv.Itoa(len(body)))
	w.WriteHeader(response.HTTPCode)
	_, err = w.Write(body)
	return err
}

func WriteServerError(ctx context.Context, w http.ResponseWriter, err error) error {
	log.Ctx(ctx).Error().Err(err).Msg("500")
	return WriteResponse(ctx, w, ErrorResponse{
		HTTPCode:  500,
		ErrorCode: "InternalError",
		Message:   err.Error(),
	})
}

func WriteErrorFunc(knownErrors map[error]ErrorResponse) func(context.Context, http.ResponseWriter, error) error {
	return func(ctx context.Context, w http.ResponseWriter, err error) error {
		if response, ok := err.(HTTPResponse); ok {
			return WriteResponse(ctx, w, response)
		} else if response, ok := knownErrors[err]; ok {
			response.Message = err.Error()
			return WriteResponse(ctx, w, response)
		} else {
			return WriteServerError(ctx, w, err)
		}
	}
}

func WriteResponseOrErrorFunc(writeError func(context.Context, http.ResponseWriter, error) error) func(context.Context, http.ResponseWriter, interface{}, error) error {
	return func(ctx context.Context, w http.ResponseWriter, data interface{}, err error) error {
		if err != nil {
			return writeError(ctx, w, err)
		}
		return WriteResponse(ctx, w, data)
	}
}

func GetQueryInt(req *http.Request, key string) (int, error) {
	value, ok, err := GetQueryIntOpt(req, key)
	if err != nil {
		return 0, err
	} else if !ok {
		return 0, ErrorResponse{
			HTTPCode:  400,
			ErrorCode: "InvalidParam",
			Message:   "Parameter '" + key + "' is required",
		}
	}
	return value, nil
}

func GetQueryIntOpt(req *http.Request, key string) (int, bool, error) {
	values, ok := req.URL.Query()[key]
	if !ok || len(values) == 0 {
		return 0, false, nil
	}
	value, err := strconv.Atoi(values[0])
	if err != nil {
		return 0, true, ErrorResponse{
			HTTPCode:  400,
			ErrorCode: "InvalidParam",
			Message:   "Parameter '" + key + "' is not integer",
		}
	}
	return value, true, nil
}
