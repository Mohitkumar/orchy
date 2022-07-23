package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohitkumar/orchy/server/container"
	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/service"
	"go.uber.org/zap"
)

type Server struct {
	http.Server
	Port            int
	container       *container.DIContiner
	executorService *service.WorkflowExecutionService
}

func NewServer(httpPort int, container *container.DIContiner, executorService *service.WorkflowExecutionService) (*Server, error) {

	s := &Server{
		Server: http.Server{
			Addr: fmt.Sprintf(":%d", httpPort),
		},
		container:       container,
		executorService: executorService,
		Port:            httpPort,
	}

	router := mux.NewRouter()
	router.HandleFunc("/workflow", s.HandleCreateFlow).Methods(http.MethodPost)
	router.HandleFunc("/workflow/{name}", s.HandleGetFlow).Methods(http.MethodGet)
	router.HandleFunc("/flow/execute", s.HandleRunFlow).Methods(http.MethodPost)
	router.Use(loggingMiddleware)
	s.Handler = router
	return s, nil
}

func (s *Server) Start() error {
	logger.Info("startting http server on", zap.Int("port", s.Port))
	if err := s.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) Stop() error {
	logger.Info("stopping http server")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.Shutdown(ctx)
	if err != nil {
		logger.Error("error shutting down http server")
	}
	return nil
}
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondOK(w http.ResponseWriter, message string) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	res, _ := json.Marshal(map[string]string{"message": message})
	w.Write(res)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}
