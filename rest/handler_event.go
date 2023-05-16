package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
	"go.uber.org/zap"
)

func (s *Server) HandleEvent(w http.ResponseWriter, r *http.Request) {
	var req model.WorkflowEvent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	defer r.Body.Close()
	err := s.executorService.ConsumeEvent(req.Name, req.FlowId, req.Event)
	if err != nil {
		logger.Error("error consuming event", zap.String("name", req.Name), zap.String("id", req.FlowId), zap.String("event", req.Event), zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error consuming event workflow")
		return
	}
	respondOKWithoutBody(w)
}
