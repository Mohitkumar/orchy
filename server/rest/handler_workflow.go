package rest

import (
	"encoding/json"
	"net/http"

	"github.com/mohitkumar/orchy/server/logger"
	"github.com/mohitkumar/orchy/server/model"
	"go.uber.org/zap"
)

func (s *Server) HandleRunFlow(w http.ResponseWriter, r *http.Request) {
	var runReq model.WorkflowRunRequest
	if err := json.NewDecoder(r.Body).Decode(&runReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	defer r.Body.Close()
	flowId, err := s.executorService.StartFlow(runReq.Name, runReq.Input)
	if err != nil {
		logger.Error("error running workflow", zap.String("name", runReq.Name), zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error running workflow")
		return
	}
	respondOK(w, map[string]any{"flowId": flowId})
}

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
