package core_http_middleware

import (
	"net/http"
	"time"

	core_logger "github.com/Trykach34rus/Golang-todoapp/internal/core/logger"
	core_http_response "github.com/Trykach34rus/Golang-todoapp/internal/core/transport/http/response"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const(
	requestIDHeader = "X-Request-ID"	
)


func CORS() Middleware {
    allowedOrigins := map[string]struct{}{
        "http://localhost:3000": {},
        "http://localhost:5050": {},
				"http://127.0.0.1:5050": {},
				"null":{},
    }

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            origin := r.Header.Get("Origin")

            if _, ok := allowedOrigins[origin]; ok {
                w.Header().Set("Access-Control-Allow-Origin", origin)
                w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
                w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
                w.Header().Set("Vary", "Origin")
            }

            if r.Method == http.MethodOptions {
                w.WriteHeader(http.StatusNoContent)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(requestIDHeader)
			if requestID == "" {
				requestID = uuid.NewString() 
			}
			r.Header.Set(requestIDHeader,requestIDHeader)
			w.Header().Set(requestIDHeader,requestIDHeader)

			next.ServeHTTP(w,r)
 		})
	}
}


func Logger(log *core_logger.Logger) Middleware  {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
			requestID := r.Header.Get(requestIDHeader)

			l:=log.With(
				zap.String("request_id",requestID),
				zap.String("url",r.URL.String()),
			)

			ctx := core_logger.ToContext(r.Context(),l)

			next.ServeHTTP(w,r.WithContext(ctx))
		})

	}
}



func Trace() Middleware  {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			rw := core_http_response.NewResponseWriter(w)

			before := time.Now()


			log.Debug(
				">>> incoming HTTP request",
				zap.String("http_method",r.Method),
				zap.Time("time",before.UTC()),

			)
			
			next.ServeHTTP(rw,r)

			log.Debug(
				"<<< done HTTP request",
				zap.Int("status code",rw.GetStatusCode()),
				zap.Duration("latency",time.Now().Sub(before)),

			)
		})
	}
}

func Panic() Middleware  {
	 return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			log := core_logger.FromContext(ctx)
			responseHandler := core_http_response.NewHTTPResponseHandler(log,w)

			defer func ()  {
				if p := recover(); p != nil {
					responseHandler.PanicResponse(
						p,
						"during handler HTTP request got unexpected panic",
					)
				}
			}()

			next.ServeHTTP(w,r)
		})
	 }
}

