package main

import (
	"fmt"
	"net/http"
	"text/template"
)

// func (apiCfg *apiConfig) handlerView(w http.ResponseWriter, r *http.Request, user database.User) {
func (apiCfg *apiConfig) handlerView(w http.ResponseWriter, r *http.Request) {
	//placeholder := []byte("signature list here")
	html, err := template.ParseFiles("html/view.html")
	//_, err := w.Write(placeholder)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error generate view: %v", err))
		return
	}
	err = html.Execute(w, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Error generate view: %v", err))
	}
}
