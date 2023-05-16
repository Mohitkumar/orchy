package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
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

func (s *Server) HandlePauseFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowName, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	flowId, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	err := s.executorService.PauseFlow(flowName, flowId)
	if err != nil {
		logger.Error("error pausing workflow", zap.String("name", flowName), zap.String("id", flowId), zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error consuming event workflow")
		return
	}
	respondOKWithoutBody(w)
}

func (s *Server) HandleResumeFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowName, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	flowId, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	err := s.executorService.ResumeFlow(flowName, flowId)
	if err != nil {
		logger.Error("error resuming workflow", zap.String("name", flowName), zap.String("id", flowId), zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error consuming event workflow")
		return
	}
	respondOKWithoutBody(w)
}

func (s *Server) HandleGetFlowExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowName, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	flowId, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	flowContext, err := s.executorService.GetFlow(flowName, flowId)
	if err != nil {
		logger.Error("error getting flow execution", zap.String("name", flowName), zap.String("id", flowId), zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "flow execution not found")
		return
	}
	state := "UNDEFINED"
	switch flowContext.State {
	case model.RUNNING:
		state = "RUNNING"
	case model.COMPLETED:
		state = "COMPLETED"
	case model.FAILED:
		state = "FAILED"
	case model.WAITING_DELAY:
		state = "DELAY WAITING"
	case model.WAITING_EVENT:
		state = "EVENT WAITING"
	case model.PAUSED:
		state = "PAUSED"
	}
	runningActions := make([]int, 0, len(flowContext.CurrentActionIds))
	for k := range flowContext.CurrentActionIds {
		runningActions = append(runningActions, k)
	}

	executedActions := make([]int, 0, len(flowContext.ExecutedActions))
	for k := range flowContext.ExecutedActions {
		executedActions = append(executedActions, k)
	}

	flowExecution := model.FlowExecution{
		Id:              flowContext.Id,
		State:           state,
		RunningActions:  runningActions,
		ExecutedActions: executedActions,
		Event:           flowContext.Event,
		Data:            flowContext.Data,
	}
	respondWithJSON(w, http.StatusOK, flowExecution)
}
