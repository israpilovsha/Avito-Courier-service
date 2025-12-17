package handler

import "github.com/gorilla/mux"

func RegisterDeliveryRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/delivery/assign", h.Assign).Methods("POST")
	r.HandleFunc("/delivery/unassign", h.Unassign).Methods("POST")
}
