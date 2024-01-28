package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecordingWinsAndRetrievingThem(t *testing.T) {
	store := InMemoryPlayerStore{map[string]int{}}
	server := NewPlayerServer(&store)
	player := "Pepper"

	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))
	server.ServeHTTP(httptest.NewRecorder(), newPostWinRequest(player))

	response := httptest.NewRecorder()
	server.ServeHTTP(response, newGetScoreRequest(player))
	assertStatus(response.Code, http.StatusOK, t)

	assertResponseBody(response.Body.String(), "3", t)

	response = httptest.NewRecorder()
	server.ServeHTTP(response, newGetLeagueRequest())
	assertStatus(response.Code, http.StatusOK, t)
	var got []Player
	err := json.NewDecoder(response.Body).Decode(&got)
	assertNoErrorWhileDecodingJson(err, t, response)
	assertPlayers(got, []Player{{"Pepper", 3}}, t)

}
