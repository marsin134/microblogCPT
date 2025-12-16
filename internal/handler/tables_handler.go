package handlers

import (
	"encoding/json"
	"net/http"
)

type TablesResponse struct {
	CountTables int `json:"countTables"`
}

func (h *Handlers) TablesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	count, err := h.TablesService.GetCountTablesBD(h.TablesRepo)
	if err != nil {
		WriteError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TablesResponse{count})
}
