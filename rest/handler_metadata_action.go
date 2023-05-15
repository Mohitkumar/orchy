package rest

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/mohitkumar/orchy/logger"
	"github.com/mohitkumar/orchy/model"
	"go.uber.org/zap"
)

func (s *Server) HandleCreateActionDefinition(w http.ResponseWriter, r *http.Request) {
	var actionDef model.ActionDefinition
	if err := json.NewDecoder(r.Body).Decode(&actionDef); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
	defer r.Body.Close()
	err := s.metadataService.GetMetadataStorage().SaveActionDefinition(actionDef)
	if err != nil {
		logger.Error("error creating action definition", zap.Error(err))
		respondWithError(w, http.StatusBadRequest, "error creating action definition")
		return
	}
	respondOK(w, map[string]any{"created": true})
}

func (s *Server) HandleGetActionDefinition(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	actionName, ok := vars["name"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
	}
	wf, err := s.metadataService.GetMetadataStorage().GetActionDefinition(actionName)
	if err != nil {
		logger.Info("action definition not found", zap.String("name", actionName))
		respondWithError(w, http.StatusBadRequest, "action definition not found")
		return
	}
	respondWithJSON(w, http.StatusOK, wf)
}
