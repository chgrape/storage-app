package handler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, fmt.Sprintf("error parsing form: %v", err), http.StatusBadRequest)
	}
	var c http.Client

	values := url.Values{
		"client_id":  {"media-cli"},
		"username":   {r.FormValue("username")},
		"password":   {r.FormValue("password")},
		"grant_type": {"password"},
	}

	res, err := c.PostForm("http://"+os.Getenv("KEYCLOAK_ADDR")+"/realms/media/protocol/openid-connect/token", values)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	defer res.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	io.Copy(w, res.Body)
}
