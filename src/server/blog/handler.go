package blog

import (
	"net/http"
	"time"

	"server/tools"
	"server/users"

	u "github.com/quollix/common/utils"
)

type Handler struct {
	UserRepo     users.UserRepository
	GlobalConfig *tools.GlobalConfig
}

func (h *Handler) MarkBlogCheckedAndRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	user, err := users.GetAuthFromContext(r)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	user, err = h.UserRepo.GetUserById(user.Id)
	if err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}
	user.LastBlogCheck = time.Now().UTC()
	if err := h.UserRepo.UpdateUser(user); err != nil {
		u.WriteResponseError(w, nil, err)
		return
	}

	http.Redirect(w, r, h.GlobalConfig.BlogURL, http.StatusSeeOther)
}

func (h *Handler) MockBlogPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}
	_, _ = w.Write([]byte("Mock Quollix blog page for component tests"))
}
