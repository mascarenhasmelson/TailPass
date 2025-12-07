package api

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Router struct {
	ctx  context.Context
	pool *pgxpool.Pool
	mux  *http.ServeMux
}

func NewRouter(ctx context.Context, pool *pgxpool.Pool) *http.ServeMux {
	r := &Router{
		ctx:  ctx,
		pool: pool,
		mux:  http.NewServeMux(),
	}

	r.routes()
	return r.mux
}

func (r *Router) routes() {
	r.mux.HandleFunc("/services", r.servicesHandler)
	r.mux.HandleFunc("/services/", r.serviceHandler)
	// r.mux.HandleFunc("/services/isp", r.ispHandler)
}

func (r *Router) servicesHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}
	if req.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	switch req.Method {
	case http.MethodGet:
		HandleFetchServices(r.ctx, w, r.pool)
	case http.MethodPost:
		HandleAddService(r.ctx, w, r.pool, req)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (r *Router) serviceHandler(w http.ResponseWriter, req *http.Request) {
	if EnableCORS(&w, req) {
		return
	}

	idStr := strings.TrimPrefix(req.URL.Path, "/services/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid service ID", http.StatusBadRequest)
		return
	}

	if req.Method == http.MethodDelete {
		HandleDeleteService(r.ctx, w, r.pool, id)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// func (r *Router) ispHandler(w http.ResponseWriter, req *http.Request) {
// 	if EnableCORS(&w, req) {
// 		return
// 	}
// 	if req.Method == http.MethodOptions {
// 		w.WriteHeader(http.StatusNoContent)
// 		return
// 	}
// 	switch req.Method {
// 	case http.MethodGet:
// 		HandleGetISPInfo(r.ctx, w)
// 	default:
// 		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
// 	}
// }
