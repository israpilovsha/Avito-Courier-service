package handler

import "github.com/gorilla/mux"

func RegisterCourierRoutes(r *mux.Router, h *Handler) {
	r.HandleFunc("/courier/{id}", h.GetByID).Methods("GET")
	r.HandleFunc("/couriers", h.GetAll).Methods("GET")
	r.HandleFunc("/courier", h.Create).Methods("POST")
	r.HandleFunc("/courier", h.Update).Methods("PUT")
	r.HandleFunc("/ping", h.Ping).Methods("GET")
	r.HandleFunc("/healthcheck", h.HealthCheck).Methods("HEAD")
}
