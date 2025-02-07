package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gaba-bouliva/Chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits *atomic.Int32
	db             *database.Queries
	env            string
}

func main() {

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	dbURL := os.Getenv("DB_URL")
	envPlatform := os.Getenv("PLATFORM")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}

	dbQueries := database.New(db)
	fmt.Println("db connection established")

	apiCfg := apiConfig{
		fileserverHits: &atomic.Int32{},
		db:             dbQueries,
		env:            envPlatform,
	}

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.countHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirp)
	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("running server on localhost%s", server.Addr)
	err = server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email string `json:"email"`
	}

	var reqBodyParam reqBody

	w.Header().Set("Content-Type", "appliation/json")

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBodyParam)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error: ", err.Error())
		fmt.Fprintf(w, `{ "error": "invalid param(s) provided}`)
		return
	}

	createUserParam := database.CreateUserParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     reqBodyParam.Email,
	}

	createdUser, err := cfg.db.CreateUser(context.Background(), createUserParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("error: ", err.Error())
		fmt.Fprintf(w, `{ "error": "could not create user in db"}`)
		return
	}

	type resBody struct {
		ID        int32     `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	jsonData := resBody{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	jsonRes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalln(err)
	}
	w.WriteHeader(201)
	w.Write(jsonRes)
}

func validate_chirp(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Body string `json:"body"`
	}

	reqParams := ReqBody{}
	w.Header().Set("Content-Type", "application/json")

	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&reqParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"error": Something went wrong}`)
		return
	}

	if len(reqParams.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, `{"error": Chirp is too long}`)
		return
	}
	unWantedWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	reqBodyWords := strings.Split(reqParams.Body, " ")
	for i, word := range reqBodyWords {

		if hiddenWord, exists := unWantedWords[strings.ToLower(word)]; exists {
			reqBodyWords[i] = hiddenWord
		}
	}

	processedBodyStr := strings.Join(reqBodyWords, " ")

	w.WriteHeader(200)
	fmt.Fprintf(w, `{"cleaned_body": %q}`, processedBodyStr)

}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) countHits(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset-utf-8")
	w.WriteHeader(http.StatusOK)
	htmlRes := fmt.Sprintf(`
	<html>
	  <body>
	   <h1>Welcome, Chirpy Admin</h1>
	   <p>Chirpy has been visited %d times!</p>
	 </body>
	</html>`, cfg.fileserverHits.Load())
	w.Write([]byte(htmlRes))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset-utf-8")
	if cfg.env != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("you can't perfom this action in current environment"))
		return
	}
	cfg.fileserverHits.Swap(0)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}
