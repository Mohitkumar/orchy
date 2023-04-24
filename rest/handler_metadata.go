package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
)

func (s *Server) HandleCreateFlow(w http.ResponseWriter, r *http.Request) {
	var fl model.Workflow
	if err := json.NewDecoder(r.Body).Decode(&fl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	defer r.Body.Close()
	err := s.metadataService.ValidateFlow(fl)
	if err != nil {
		logger.Error("error validating workflow", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	err = s.metadataService.GetMetadataStorage().SaveWorkflowDefinition(fl)
	if err != nil {
		logger.Error("error creating workflow", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error creating workflow")
		return
	}
	respondOK(w, map[string]any{"created": true})
}

func (s *Server) HandleGetFlow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	flowName, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	wf, err := s.metadataService.GetMetadataStorage().GetWorkflowDefinition(flowName)
	if err != nil {
		logger.Info("wokflow does not exist", zap.String("name", flowName))
		respondWithError(w, http.StatusBadRequest, "wokflow does not exist")
		return
	}
	respondWithJSON(w, http.StatusOK, wf)
}
