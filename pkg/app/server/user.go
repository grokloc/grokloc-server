package server

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/grokloc/grokloc-server/pkg/app/admin/user/events"
	"github.com/grokloc/grokloc-server/pkg/models"
	"github.com/grokloc/grokloc-server/pkg/security"
	"go.uber.org/zap"
)

func (srv *Instance) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	defer zap.L().Sync() // nolint
	sugar := zap.L().Sugar()

	authLevel, ok := ctx.Value(authLevelCtxKey).(int)
	if !ok {
		panic("auth missing")
	}

	session, ok := ctx.Value(sessionCtxKey).(Session)
	if !ok {
		panic("session missing")
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
		http.Error(w, "malformed user create event", http.StatusBadRequest)
		return
	}

	if (authLevel == AuthOrg && session.Org.ID != event.Org) ||
		(authLevel == AuthUser) {
		// caller was either org owner (but not for org of prospective user),
		// or was just a regular user (and can't create other users)
		http.Error(w, "auth inadequate", http.StatusForbidden)
		return
	}

	event.Password, err = security.DerivePassword(event.Password, srv.ST.Argon2Cfg)
	if err != nil {
		sugar.Debugw("derive password",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	u, err := srv.UserController.Create(ctx, event)
	if err != nil {
		if err == models.ErrConflict {
			http.Error(w, "duplicate user args", http.StatusConflict)
			return
		}
		sugar.Debugw("insert uer",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	bs, err := json.Marshal(u)
	if err != nil {
		sugar.Debugw("marshal user json",
			"reqid", middleware.GetReqID(ctx),
			"err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("location", UserRoute+"/"+u.ID)
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
