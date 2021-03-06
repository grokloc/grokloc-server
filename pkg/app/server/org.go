package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-server/pkg/app/admin/org/events"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"go.uber.org/zap"
)

func (srv *Instance) CreateOrg(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer zap.L().Sync() // nolint
	sugar := zap.L().Sugar()

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}
	// only root can create an org
	if authLevel != AuthRoot {
		http.Error(w, "auth inadequate", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		sugar.Debugw("read body",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	var event events.Create
	err = json.Unmarshal(body, &event)
	if err != nil {
		http.Error(w, "malformed org create event", http.StatusBadRequest)
		return
	}

	// password assumed cleartext, derive
	ownerPassword, err := security.DerivePassword(
		event.OwnerPassword,
		srv.ST.Argon2Cfg,
	)
	if err != nil {
		sugar.Debugw("derive org owner password",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	event.OwnerPassword = ownerPassword

	o, err := srv.OrgController.Create(ctx, event)
	if err != nil {
		if err == models.ErrConflict {
			http.Error(w, "duplicate org args", http.StatusConflict)
			return
		}
		sugar.Debugw("insert org",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	bs, err := json.Marshal(o)
	if err != nil {
		sugar.Debugw("marshal org json",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("location", OrgRoute+"/"+o.ID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(bs)
	if err != nil {
		sugar.Debugw("write http output",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
}
