package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
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

func (s *StubPlayerStore) GetLeagueTable() []Player {
	players := []Player{}
	for k, v := range s.scores {
		players = append(players, Player{k, v})
	}
	return players
}

func TestGETPlayers(t *testing.T) {
	playerStore := &StubPlayerStore{map[string]int{"Pepper": 20, "Floyd": 10}, []string{}}
	server := NewPlayerServer(playerStore)
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

	server := NewPlayerServer(&store)

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

func TestLeague(t *testing.T) {
	store := StubPlayerStore{map[string]int{"Vineeth": 10, "Ajay": 20, "Zara": 30}, []string{}}
	server := NewPlayerServer(&store)

	t.Run("it returns 200 on /league and gets players info", func(t *testing.T) {
		request := newGetLeagueRequest()
		response := httptest.NewRecorder()

		server.ServeHTTP(response, request)

		var got []Player
		want := []Player{
			{
				"Vineeth",
				10,
			}, {
				"Ajay",
				20,
			}, {
				"Zara",
				30,
			},
		}
		err := json.NewDecoder(response.Body).Decode(&got)
		assertStatus(response.Code, http.StatusOK, t)
		assertJsonContentType(response, t)
		assertNoErrorWhileDecodingJson(err, t, response)
		assertPlayers(got, want, t)
	})
}

func assertJsonContentType(response *httptest.ResponseRecorder, t testing.TB) {
	t.Helper()
	if response.Result().Header.Get("content-type") != "application/json" {
		t.Errorf("response did not have content-type of application/json, got %v", response.Result().Header)
	}
}

func assertNoErrorWhileDecodingJson(err error, t testing.TB, response *httptest.ResponseRecorder) {
	if err != nil {
		t.Fatalf("Unable to parse response from server %q into slice of Player, '%v'", response.Body, err)
	}
}

func assertPlayers(got []Player, want []Player, t testing.TB) {
	sort.SliceStable(got, func(i, j int) bool {
		return got[i].Name < got[j].Name
	})
	sort.SliceStable(want, func(i, j int) bool {
		return want[i].Name < want[j].Name
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v want %v", got, want)
	}
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

func newPostWinRequest(player string) *http.Request {
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/players/%s", player), nil)
	return request
}

func newGetLeagueRequest() *http.Request {
	request, _ := http.NewRequest(http.MethodGet, "/league", nil)
	return request
}
