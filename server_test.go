package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type StubPlayerStore struct {
	scores   map[string]int
	winCalls []string
}

func (s *StubPlayerStore) GetPlayerScore(name string) int {
	return s.scores[name]
}

func (s *StubPlayerStore) RecordWin(name string) {
	s.winCalls = append(s.winCalls, name)
}

func TestGETPlayers(t *testing.T) {
	playerStore := &StubPlayerStore{map[string]int{"Pepper": 20, "Floyd": 10}, []string{}}
	server := &PlayerServer{playerStore}
	t.Run("returns peppers score", func(t *testing.T) {
		request := newGetScoreRequest("Pepper")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(response.Code, http.StatusOK, t)
		assertResponseBody(response.Body.String(), "20", t)
	})

	t.Run("return floyds score", func(t *testing.T) {
		request := newGetScoreRequest("Floyd")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)
		assertStatus(response.Code, http.StatusOK, t)
		assertResponseBody(response.Body.String(), "10", t)
	})

	t.Run("return 404 on missing players", func(t *testing.T) {
		request := newGetScoreRequest("Apollo")
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		got := response.Code
		want := http.StatusNotFound

		assertStatus(got, want, t)
	})
}

func TestStoreWins(t *testing.T) {
	store := StubPlayerStore{
		map[string]int{},
		[]string{},
	}

	server := &PlayerServer{&store}

	t.Run("it returns accepted on POST", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodPost, "/players/Pepper", nil)
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		assertStatus(response.Code, http.StatusAccepted, t)

		want := []string{"Pepper"}
		if !reflect.DeepEqual(store.winCalls, want) {
			t.Errorf("Expected call to record win but not found. got %+v, want %+v", store.winCalls, want)
		}
	})
}

func assertStatus(got int, want int, t testing.TB) {
	t.Helper()
	if got != want {
		t.Errorf("got status %d want %d", got, want)
	}
}

func assertResponseBody(got string, want string, t testing.TB) {
	t.Helper()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func newGetScoreRequest(name string) *http.Request {
	request, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/players/%s", name), nil)
	return request
}
