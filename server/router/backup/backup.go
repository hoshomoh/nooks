// Package backup serves the Instance's database as a file to download.
package backup

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/hoshomoh/nooks/server/auth"
	"github.com/hoshomoh/nooks/store"
)

// Handler streams a copy of the database to an Admin.
type Handler struct {
	store    store.Store
	resolver *auth.Resolver
	now      func() time.Time
}

// NewHandler builds it. now may be nil, in which case time.Now is used.
func NewHandler(s store.Store, resolver *auth.Resolver, now func() time.Time) *Handler {
	if now == nil {
		now = time.Now
	}
	return &Handler{store: s, resolver: resolver, now: now}
}

/*
ServeHTTP hands over the whole Instance as one file.

A download rather than an RPC, because that is what it is: a browser should be able to
save it, and a script should be able to `curl` it into a backup directory. Nothing about
it is shaped for a client to parse.

Admins only. It contains every Member's password hash and every List in the household,
which is exactly what a backup has to contain and exactly why it is not everybody's.
*/
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	grant, ok := h.resolver.Grant(r.Context(), r.Header)
	if !ok || !grant.Member.IsAdmin() {
		// The same answer either way: whether an Instance has a backup route worth
		// finding is not something an anonymous request gets to learn.
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	dir, err := os.MkdirTemp("", "nooks-backup")
	if err != nil {
		http.Error(w, "could not start a backup", http.StatusInternalServerError)
		return
	}
	// The copy exists only for as long as it takes to send it.
	defer func() { _ = os.RemoveAll(dir) }()

	path := filepath.Join(dir, "nooks.db")
	if err := h.store.BackupTo(r.Context(), path); err != nil {
		if errors.Is(err, store.ErrNoBackup) {
			http.Error(w, "this instance's database exports with its own tools", http.StatusNotImplemented)
			return
		}
		http.Error(w, "could not copy the database", http.StatusInternalServerError)
		return
	}

	file, err := os.Open(path)
	if err != nil {
		http.Error(w, "could not read the backup", http.StatusInternalServerError)
		return
	}
	defer func() { _ = file.Close() }()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", nameFor(h.now())))
	// Never a cached copy of somebody's whole household.
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "", h.now(), file)
}

// nameFor is what the file is called once it is on somebody's machine. The date is in
// it because a backup nobody can date is a backup nobody trusts.
func nameFor(at time.Time) string {
	return fmt.Sprintf("nooks-%s.db", at.Format("2006-01-02"))
}
