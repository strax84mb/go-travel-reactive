package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/strax84mb/go-travel-reactive/internal/web"
)

func RegisterUserHandlers(r *mux.Router, service authService) {
	r.Methods(http.MethodPost).Path("/user/login").HandlerFunc(login(service))
	r.Methods(http.MethodPost).Path("/user/signup").HandlerFunc(signup(service))
}

func login(service authService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			web.BadRequest(w, "payload needed", nil)
			return
		}

		defer r.Body.Close()

		var payload loginInput

		if err = json.Unmarshal(bytes, &payload); err != nil {
			web.BadRequest(w, "incorrect payload", nil)
			return
		}

		token, err := service.Login(r.Context(), payload.Username, payload.Password)
		if err != nil {
			web.InternalServerError(w, "could not login", map[string][]string{
				"error": {err.Error()},
			})

			return
		}

		web.Ok(w, loginOutput{Token: token})
	}
}

type loginInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginOutput struct {
	Token string `json:"token"`
}

func signup(service authService) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadAll(r.Body)
		if err != nil {
			web.BadRequest(w, "payload needed", nil)
			return
		}

		defer r.Body.Close()

		var payload loginInput

		if err = json.Unmarshal(bytes, &payload); err != nil {
			web.BadRequest(w, "incorrect payload", nil)
			return
		}

		id, err := service.SaveUser(r.Context(), payload.Username, payload.Password)
		if err != nil {
			web.InternalServerError(w, "could not sign up", map[string][]string{
				"error": {err.Error()},
			})

			return
		}

		web.Created(w, struct {
			ID int `json:"id"`
		}{
			ID: id,
		})
	}
}
