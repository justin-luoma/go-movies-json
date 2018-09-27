package main // import "go-movies-json"

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/pat"
)

type config struct {
	Referers []string `json:"referers"`
}

type movies struct {
	Movies []movie `json:"movies" `
}

type movie struct {
	Title  string `json:"title"`
	Rating int    `json:"rating"`
	ID     int    `json:"id"`
}

type srv struct {
	c *config
	m *movies
	n int
}

func homeHandler(wr http.ResponseWriter, req *http.Request) {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatal(err)
	}
	wr.WriteHeader(http.StatusOK)
	fmt.Fprintf(wr, "%q", dump)
}

func write404Status(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("{\"status\":\"not found\"}"))
}

func write500Status(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("{\"status\":\"internal server error\"}"))
}

func write403Status(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte("{\"status\":\"forbidden\"}"))
}

func write200Status(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\"status\":\"ok\"}"))
}

func write200StatusWithBody(w http.ResponseWriter, body interface{}) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(body)
}

func (s *srv) moviesContainsID(id int) int {
	for i := range s.m.Movies {
		if s.m.Movies[i].ID == id {
			return i
		}
	}

	return -1
}

func (s *srv) refererHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(s.c.Referers) != 0 {
			if !containsString(s.c.Referers, r.Referer()) {
				write403Status(w)
				return
			}
		} else {
			log.Println("no referers defined or debug enabled")
		}
		h.ServeHTTP(w, r)
	})
}

func (s *srv) getMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("%q", err)
		write500Status(w)
		return
	}

	if i := s.moviesContainsID(id); i >= 0 {
		write200StatusWithBody(w, s.m.Movies[i])
		return
	}

	write404Status(w)
}

func (s *srv) getMovies(w http.ResponseWriter, r *http.Request) {
	write200StatusWithBody(w, s.m.Movies)
}

func (s *srv) addMovie(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%q\n", dump)
	defer r.Body.Close()
	var m movie
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("failed to decode message body with error : %v\n", err)
		write500Status(w)
		return
	}

	if m.ID == 0 {
		m.ID = s.n
		s.n++
	}

	s.m.Movies = append(s.m.Movies, m)

	fmt.Printf("%+v\n", s.m.Movies)
	write200Status(w)
}

func (s *srv) updateMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("%q", err)
		write500Status(w)
		return
	}

	fmt.Println(id)

	if i := s.moviesContainsID(id); i >= 0 {
		defer r.Body.Close()
		var m movie
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			log.Printf("failed to decode message body with error : %v\n", err)
			write500Status(w)
			return
		}
		s.m.Movies[i] = movie{
			Title:  m.Title,
			Rating: m.Rating,
			ID:     id,
		}
		write200Status(w)
		return
	}

	write404Status(w)
}

func (s *srv) deleteMovie(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get(":id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("%q", err)
		write500Status(w)
		return
	}

	if i := s.moviesContainsID(id); i >= 0 {
		s.m.Movies = append(s.m.Movies[:i], s.m.Movies[i+1:]...)
		write200Status(w)
		return
	}

	write404Status(w)
}

// func (s *srv) mailHandler(w http.ResponseWriter, r *http.Request) {
// 	if s.c == nil {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
// 		return
// 	}

// if len(s.c.Referers) != 0 {
// 	if !containsString(s.c.Referers, r.Referer()) {
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.WriteHeader(http.StatusForbidden)
// 		fmt.Fprint(w, http.StatusText(http.StatusForbidden))
// 		return
// 	}
// } else {
// 	log.Println("no referers defined")
// }

// 	defer r.Body.Close()
// 	var m message
// if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
// 	log.Printf("failed to decode message body with error : %v\n", err)
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.WriteHeader(http.StatusInternalServerError)
// 	fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
// 	return
// }

// 	if m.Honeypot != "" {
// 		goto StatusOK
// 	}

// 	if err := s.sendMessage(m.Email, m.Name, m.Message, s.c.ToAddress); err != nil {
// 		log.Printf("failed to send email with error: %v\n", err)
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.WriteHeader(http.StatusInternalServerError)
// 		fmt.Fprint(w, http.StatusText(http.StatusInternalServerError))
// 		return
// 	}

// StatusOK:
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Access-Control-Allow-Origin", "*")
// 	w.Header().Set("X-Content-Type-Options", "nosniff")
// 	w.WriteHeader(http.StatusOK)
// 	w.Write([]byte("{\"status\":\"ok\"}"))
// }

func (s *srv) loadConfigFromFile(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		s.c = nil
		return err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		s.c = nil
		return err
	}

	var c config
	if err := json.Unmarshal(bytes, &c); err != nil {
		s.c = nil
		return err
	}
	s.c = &c

	return nil
}

func (s *srv) loadConfigFromEnv() error {
	c := config{
		Referers: strings.Split(os.Getenv("REFERERS"), ","),
	}

	s.c = &c

	return nil
}

func (s *srv) loadDBFromFile(path string) error {
	jsonFile, err := os.Open(path)
	if err != nil {
		s.m = nil
		return err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		s.m = nil
		return err
	}

	var m movies
	if err := json.Unmarshal(bytes, &m); err != nil {
		s.m = nil
		return err
	}

	max := -1
	for _, v := range m.Movies {
		if v.ID > max {
			max = v.ID
		}
	}

	max++

	s.n = max

	s.m = &m

	return nil
}

func (s *srv) saveDBToFile(path string) error {
	bytes, err := json.MarshalIndent(s.m, "", "    ")
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(path, bytes, 0644); err != nil {
		return err
	}

	return nil
}

func containsString(sl []string, v string) bool {
	for _, vv := range sl {
		if vv == v {
			return true
		}
	}
	return false
}

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	var s srv
	if err := s.loadConfigFromFile(".config"); err != nil {
		log.Println(err)
		err = nil
		if err = s.loadConfigFromEnv(); err != nil {
			log.Println(err)
		}
	}

	if err := s.loadDBFromFile("./db.json"); err != nil {
		panic(err)
	}

	router := pat.New()
	// router.Options("/movies/{id}", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, POST, GET")
	// 	write200Status(w)
	// 	return
	// })
	router.Get("/movies/{id}", s.getMovie)
	router.Get("/movies", s.getMovies)
	router.Post("/movies", s.addMovie)
	router.Put("/movies/{id}", s.updateMovie)
	router.Delete("/movies/{id}", s.deleteMovie)
	router.Options("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE, POST, GET")
		write200Status(w)
		return
	})

	// router.Get("/", homeHandler)

	httpSrv := &http.Server{
		Addr:         ":" + port,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CombinedLoggingHandler(os.Stdout, s.refererHandler(router)),
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			log.Printf("web server didn't start with error: %v", err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	if err := s.saveDBToFile("./db.json"); err != nil {
		log.Printf("failed to save db to file with error %q", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	httpSrv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
