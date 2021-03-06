package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

type Scholarship struct {
	Name   string
	Amount int
}

type scholarshipHandlers struct {
	store map[string]Scholarship
}

func (h *scholarshipHandlers) get(w http.ResponseWriter, r *http.Request) {
	scholarship := h.store["test"]
	fmt.Fprint(w, "name: ", scholarship.Name, ", amount: ", scholarship.Amount)
}

func newScholarshipHandler() *scholarshipHandlers {
	return &scholarshipHandlers{
		store: map[string]Scholarship{
			"test": Scholarship{
				Name:   "test",
				Amount: 1000,
			},
		},
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Homepage Endpoint Hit")
}

func timePage(w http.ResponseWriter, r *http.Request) {
	t := time.Now()
	year, month, day := t.Date()
	fmt.Fprint(w, "Today is: ", day, year, month)
}

func handleRequests() {
	http.HandleFunc("/", homePage)
	http.HandleFunc("/time", timePage)
	h := newScholarshipHandler()
	http.HandleFunc("/scholarships", h.get)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	handleRequests()
}
