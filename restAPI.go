package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Scholarship struct {
	Name   string `json:"Name"`
	Amount int    `json:"Amount"`
}

type Scholarships map[string]Scholarship

type scholarshipHandler struct {
	sync.Mutex
	scholarships Scholarships
}

// handle requests
func (sh *scholarshipHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		sh.get(w, r)
	case "POST":
		sh.post(w, r)
	case "PUT", "PATCH":
	 	sh.put(w, r)
	case "DELETE":
		sh.delete(w, r)
	default:
		respondWithError(w, http.StatusMethodNotAllowed, "invalid method")
	}
}

// write JSON responses
func respondWithJSON(w http.ResponseWriter, code int, data interface{}) {
	response, _ := json.Marshal(data)
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	respondWithJSON(w, code, map[string]string{"error": msg})
}

// get all scholarships or a particular scholarship, specified in Url
func (sh *scholarshipHandler) get(w http.ResponseWriter, r *http.Request) {
	defer sh.Unlock()
	sh.Lock()

	id, err := nameFromUrl(r)
	scholarship, exists := sh.scholarships[id]

	// if error not nil, means they want all scholarships
	if err != nil {
		respondWithJSON(w, http.StatusOK, sh.scholarships)
		return
	}
	if !exists {
		respondWithJSON(w, http.StatusOK, sh.scholarships)
		return
	}
	respondWithJSON(w, http.StatusOK, scholarship)
}

func nameFromUrl(r *http.Request) (string, error) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		return "", errors.New("not found")
	}
	id := parts[len(parts)-1]
	return id, nil
}

// create new scholarship
func (sh *scholarshipHandler) post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		respondWithError(w, http.StatusUnsupportedMediaType, "content type `application/json` required")
		return
	}
	var scholarship Scholarship
	err = json.Unmarshal(body, &scholarship)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer sh.Unlock()
	sh.Lock()
	sh.scholarships[scholarship.Name] = scholarship
	respondWithJSON(w, http.StatusCreated, scholarship)
}

// modify existing scholarship
func (sh *scholarshipHandler) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	name, err := nameFromUrl(r)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != "application/json" {
		respondWithError(w, http.StatusUnsupportedMediaType, "content type 'application/json' required")
		return
	}
	var scholarship Scholarship
	err = json.Unmarshal(body, &scholarship)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer sh.Unlock()
	sh.Lock()
	_, exists := sh.scholarships[name]
	if !exists {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	if scholarship.Amount != 0 {
		sh.scholarships[name] = scholarship
	}
	respondWithJSON(w, http.StatusOK, sh.scholarships[name])
}

// delete existing scholarship
func (sh *scholarshipHandler) delete(w http.ResponseWriter, r *http.Request) {
	name, err := nameFromUrl(r)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}

	defer sh.Unlock()
	sh.Lock()
	_, exists := sh.scholarships[name]
	if !exists {
		respondWithError(w, http.StatusNotFound, "not found")
		return
	}
	delete(sh.scholarships, name)
	respondWithJSON(w, http.StatusNoContent, "")
}

// scholarship handler map
func newScholarshipHandler() *scholarshipHandler {
	return &scholarshipHandler{
		scholarships: Scholarships{
			"test": Scholarship{
				Name:   "test",
				Amount: 1000,
			},
			"test2": Scholarship{
				Name:   "test2",
				Amount: 2000,
			},
		},
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Homepage Endpoint")
}

func timePage(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	year, month, day := t.Date()
	fmt.Fprint(w, "Today is: ", day, year, month)
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/time", timePage)
	sh := newScholarshipHandler()
	http.Handle("/scholarships", sh)
	http.Handle("/scholarships/", sh)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
