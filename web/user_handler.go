package web

import (
	"html/template"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/blrobin2/goreddit"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	store    goreddit.Store
	sessions *scs.SessionManager
}

func (h *UserHandler) New() http.HandlerFunc {
	type data struct {
		SessionData
		CSRF template.HTML
	}
	templ := template.Must(template.ParseFiles("templates/layout.html", "templates/user_register.html"))
	return func(w http.ResponseWriter, r *http.Request) {
		templ.Execute(w, data{
			SessionData: GetSessionData(h.sessions, r.Context()),
			CSRF:        csrf.TemplateField(r),
		})
	}
}

func (h *UserHandler) Register() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		form := RegisterUserForm{
			Username:        r.FormValue("username"),
			Password:        r.FormValue("password"),
			PasswordConfirm: r.FormValue("password-confirm"),
			UsernameTaken:   false,
		}
		if _, err := h.store.UserByUsername(form.Username); err == nil {
			form.UsernameTaken = true
		}
		if !form.Validate() {
			h.sessions.Put(r.Context(), "form", form)
			http.Redirect(w, r, r.Referer(), http.StatusFound)
			return
		}

		password, err := bcrypt.GenerateFromPassword([]byte(form.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.store.CreateUser(&goreddit.User{
			ID:       uuid.New(),
			Username: form.Username,
			Password: string(password),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		h.sessions.Put(r.Context(), "flash", "You registration was successful. Please log in.")
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
