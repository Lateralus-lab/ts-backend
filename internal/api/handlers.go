package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/Lateralus-lab/ts-backend/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Welcome to the Rave Ticket API",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllEvents(w http.ResponseWriter, r *http.Request) {
	events, err := app.DB.AllEvents()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, events)
}

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, errors.New("invalid credentials"), http.StatusBadRequest)
		return
	}

	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) refreshToken(w http.ResponseWriter, r *http.Request) {
	for _, cookie := range r.Cookies() {
		if cookie.Name == app.auth.CookieName {
			claims := &Claims{}
			refreshToken := cookie.Value

			_, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(app.JWTSecret), nil
			})
			if err != nil {
				app.errorJSON(w, errors.New("unauthorized"), http.StatusUnauthorized)
				return
			}

			userID, err := strconv.Atoi(claims.Subject)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			user, err := app.DB.GetUserByID(userID)
			if err != nil {
				app.errorJSON(w, errors.New("unknown user"), http.StatusUnauthorized)
				return
			}

			u := jwtUser{
				ID:        user.ID,
				FirstName: user.FirstName,
				LastName:  user.LastName,
			}

			tokenPairs, err := app.auth.GenerateTokenPair(&u)
			if err != nil {
				app.errorJSON(w, errors.New("error generating tokens"), http.StatusUnauthorized)
				return
			}

			http.SetCookie(w, app.auth.GetRefreshCookie(tokenPairs.RefreshToken))

			app.writeJSON(w, http.StatusOK, tokenPairs)
		}
	}
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
}

func (app *application) ManageEvents(w http.ResponseWriter, r *http.Request) {
	events, err := app.DB.AllEvents()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, events)
}

func (app *application) GetEvent(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	event, err := app.DB.OneEvent(eventID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, event)
}

func (app *application) EventForEdit(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	eventID, err := strconv.Atoi(id)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	event, genres, err := app.DB.OneEventForEdit(eventID)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	var payload = struct {
		Event  *models.Event   `json:"event"`
		Genres []*models.Genre `json:"genres"`
	}{
		event,
		genres,
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) AllGenres(w http.ResponseWriter, r *http.Request) {
	genres, err := app.DB.AllGenres()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, genres)
}

func (app *application) InsertEvent(w http.ResponseWriter, r *http.Request) {
	var event models.Event

	err := app.readJSON(w, r, &event)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	event = app.getPoster(event)

	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	newID, err := app.DB.InsertEvent(event)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	err = app.DB.UpdateEventGenres(newID, event.GenresArray)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	resp := JSONResponse{
		Error:   false,
		Message: "Event successfully updated",
	}

	app.writeJSON(w, http.StatusAccepted, resp)
}

func (app *application) getPoster(event models.Event) models.Event {
	type TheEventDB struct {
		Page    int `json:"page"`
		Results []struct {
			PosterPath string `json:"poster_path"`
		} `json:"results"`
		TotalPages int `json:"total_pages"`
	}

	client := &http.Client{}
	theUrl := fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s", app.APIKey)

	req, err := http.NewRequest("GET", theUrl+"$query="+url.QueryEscape(event.Title), nil)
	if err != nil {
		log.Panicln(err)
		return event
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		log.Panicln(err)
		return event
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panicln(err)
		return event
	}

	var responseObject TheEventDB

	json.Unmarshal(bodyBytes, &responseObject)
	if len(responseObject.Results) > 0 {
		event.Image = responseObject.Results[0].PosterPath
	}

	return event
}
