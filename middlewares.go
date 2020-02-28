package http

import (
	"encoding/json"
	"net/http"

	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	l "git.dar.kz/dareco-go/logger"
	"git.dar.kz/dareco-go/utils/jwt"
)

var errSys = &ErrorSystem{"ACL", 20}

type Handler func(w http.ResponseWriter, r *http.Request)

type Endpoint func(w http.ResponseWriter, r *http.Request) Response

//func Caching(gc cache.Cache, exp time.Duration, fn Endpoint) Endpoint {
//	return func(w http.ResponseWriter, r *http.Request) Response {
//		if r.Method != "GET" {
//			return fn(w, r)
//		}
//
//		d, er := gc.Get(r.URL.String())
//		if er == nil {
//			data, ok := d.(Response)
//			if ok {
//				return data
//			}
//		}
//
//		resp := fn(w, r)
//		if resp.StatusCode() == http.StatusOK {
//			_ = gc.SetWithExpire(r.URL.String(), resp, exp)
//		}
//
//		return resp
//	}
//}

type Form int

const (
	JsonForm Form = iota
	ImagePngForm
	FileXlsForm
	RedirectForm
)

type OktaJwt struct {
	Ver int      `json:"ver"`
	Jti string   `json:"jti"`
	Iss string   `json:"iss"`
	Aud string   `json:"aud"`
	Sub string   `json:"sub"`
	Iat int      `json:"iat"`
	Exp int      `json:"exp"`
	Cid string   `json:"cid"`
	UID string   `json:"uid"`
	Scp []string `json:"scp"`
}

func (idToken *OktaJwt) Valid() error {
	return nil
}

func OktaJWT(fn Endpoint, vldtr *jwt.OktaJWTValidator) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		authHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authHeader) < 2 {
			return errSys.BadRequest(11, "Authorization must be Bearer token")
		}
		idToken := authHeader[1]
		if idToken == "" {
			return errSys.BadRequest(10, "Authorization not provided")
		}
		tkn, err := vldtr.Validate(idToken)
		if err != nil {
			return errSys.BadRequest(10, err.Error())
		}

		ctx := context.WithValue(r.Context(), "token", tkn)
		r = r.WithContext(ctx)
		return fn(w, r)
	}
}

func ACL(log l.Logger, fn Endpoint) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		authHeader := strings.Split(r.Header.Get("Authorization"), " ")
		if len(authHeader) < 2 {
			return errSys.BadRequest(11, "Authorization must be Bearer token")
		}
		idToken := authHeader[1]
		if idToken == "" {
			return errSys.BadRequest(10, "Authorization not provided")
		}
		claims, err := jwt.Parse(idToken)
		if err != nil {
			return errSys.BadRequest(10, err.Error())
		}
		parts := strings.Split(claims.Audience, "/")
		if len(parts) < 3 {
			return errSys.BadRequest(11, "Audience is invalid")
		}
		bucket := parts[0]
		brand := parts[1]
		clientId := parts[2]

		ctx := context.WithValue(r.Context(), "bucket", bucket)
		ctx = context.WithValue(ctx, "brand", brand)
		ctx = context.WithValue(ctx, "client_id", clientId)
		ctx = context.WithValue(ctx, "user_id", claims.Subject)
		ctx = context.WithValue(ctx, "acl", claims.Acl)
		ctx = context.WithValue(ctx, "email", claims.Email)
		ctx = context.WithValue(ctx, "phone_number", claims.PhoneNumber)
		log.Info("ACL for the user with ID: ", claims.Subject, " is: ", claims.Acl)

		r = r.WithContext(ctx)
		return fn(w, r)
	}
}

func Json(fn Endpoint) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		d := fn(w, r)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		for k, v := range d.Headers() {
			w.Header().Set(k, v)
		}

		statusCode := d.StatusCode()
		if statusCode == 302 || statusCode == 301 {
			http.Redirect(w, r, d.Response().(string), statusCode)
			return
		}

		w.WriteHeader(d.StatusCode())
		err := json.NewEncoder(w).Encode(d.Response())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func ImagePNG(fn Endpoint) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		d := fn(w, r)
		w.Header().Set("Content-Type", "image/png")
		for k, v := range d.Headers() {
			w.Header().Set(k, v)
		}
		w.WriteHeader(d.StatusCode())
		resp := d.Response()
		data, ok := resp.([]byte)
		if !ok {
			json.NewEncoder(w).Encode(d.Response())
		} else {
			w.Write(data)
		}
	}
}

func FileXls(fn Endpoint) Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		d := fn(w, r)
		for k, v := range d.Headers() {
			w.Header().Set(k, v)
		}
		w.WriteHeader(d.StatusCode())
	}
}

func Logging(log l.Logger, fn Endpoint) Endpoint {
	return func(w http.ResponseWriter, r *http.Request) Response {
		start := time.Now()

		// Read the content
		var bodyBytes []byte
		if r.Body != nil {
			bodyBytes, _ = ioutil.ReadAll(r.Body)
		}
		// Restore the io.ReadCloser to its original state
		r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))
		// Use the content
		bodyString := string(bodyBytes)

		d := fn(w, r)
		dBytes, _ := json.Marshal(d.Response())

		if d.StatusCode()%200 < 100 || d.StatusCode()%300 < 100 {
			log.Warn(LogRequest(r), " body= ", bodyString, time.Since(start), " ", d.StatusCode(), " ", string(dBytes))
		} else {
			log.Debug(LogRequest(r), " body= ", bodyString, time.Since(start), " ", d.StatusCode(), " ", string(dBytes))
		}

		return d
	}
}

func LogRequest(r *http.Request) string {
	// Create return string
	var request []string
	// Add the request string
	urlPath := fmt.Sprintf("%v %v", r.Method, r.URL)
	request = append(request, urlPath)
	// Add the host
	request = append(request, fmt.Sprintf("Host: %v", r.Host))
	// Loop through headers
	for name, headers := range r.Header {
		name = strings.ToLower(name)
		for _, h := range headers {
			request = append(request, fmt.Sprintf("%v: %v", name, h))
		}
	}

	// If this is a POST, add post data
	if r.Method == "POST" {
		_ = r.ParseForm()
		request = append(request, " ")
		request = append(request, r.Form.Encode())
	}
	// Return the request as a string
	return strings.Join(request, " ") + " "
}
