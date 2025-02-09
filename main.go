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

	"github.com/gaba-bouliva/Chirpy/internal/auth"
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

type chirpsResponseBody struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    string    `json:"user_id"`
}

type usersResponseBody struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
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

	mux.HandleFunc("POST /api/chirps", apiCfg.handleChirp)
	mux.HandleFunc("GET /api/chirps", apiCfg.handleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{id}", apiCfg.handleGetChirpByID)

	mux.HandleFunc("POST /api/users", apiCfg.handleCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handleLogin)

	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("GET /admin/metrics", apiCfg.countHits)
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

func (cfg *apiConfig) handleGetChirpByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := r.PathValue("id")
	if len(id) < 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid id provided"))
		return
	}
	chirp, err := cfg.db.GetChirpById(context.Background(), id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	chirpData := chirpsResponseBody{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.Body,
	}

	jsonRes, err := json.Marshal(chirpData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(jsonRes)
}

func (cfg *apiConfig) handleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chirpList, err := cfg.db.GetAllChirps(context.Background())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	chirpListData := []chirpsResponseBody{}

	for _, chirp := range chirpList {
		newChirp := chirpsResponseBody{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		}

		chirpListData = append(chirpListData, newChirp)
	}

	jsonRes, err := json.Marshal(chirpListData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	w.Write(jsonRes)

}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var reqBodyParams reqBody

	w.Header().Set("Content-Type", "appliation/json")

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBodyParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("error: ", err.Error())
		fmt.Fprintf(w, "error encountered decoding request body")
		return
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), reqBodyParams.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "invalid email or password provided")
		}
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("error encountered please try again later"))
		return
	}

	err = auth.CheckPasswordHash(reqBodyParams.Password, user.HashedPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		log.Println(err)
		fmt.Fprintf(w, "invalid email or password provided")
		return
	}

	jsonData := usersResponseBody{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	jsonRes, err := json.Marshal(jsonData)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(200)
	w.Write(jsonRes)

}

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type reqBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var reqBodyParams reqBody

	w.Header().Set("Content-Type", "appliation/json")

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&reqBodyParams)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("error: ", err.Error())
		fmt.Fprintf(w, `{ "error": "invalid param(s) provided}`)
		return
	}

	password, err := auth.HashPassword(reqBodyParams.Password)
	if err != nil {
		if err.Error() == "invalid password provided" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error encountered please try agin"))
		log.Println(err)
		return
	}

	createUserParam := database.CreateUserParams{
		ID:             uuid.NewString(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          reqBodyParams.Email,
		HashedPassword: password,
	}

	createdUser, err := cfg.db.CreateUser(context.Background(), createUserParam)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println("error: ", err.Error())
		fmt.Fprintf(w, `{ "error": "could not create user in db"}`)
		return
	}

	jsonData := usersResponseBody{
		ID:        createdUser.ID,
		CreatedAt: createdUser.CreatedAt,
		UpdatedAt: createdUser.UpdatedAt,
		Email:     createdUser.Email,
	}

	jsonRes, err := json.Marshal(jsonData)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(201)
	w.Write(jsonRes)
}

func (cfg *apiConfig) handleChirp(w http.ResponseWriter, r *http.Request) {
	type ReqBody struct {
		Body   string `json:"body"`
		UserId string `json:"user_id"`
	}
	reqParams := ReqBody{}

	w.Header().Set("Content-Type", "application/json")
	jsonDecoder := json.NewDecoder(r.Body)
	err := jsonDecoder.Decode(&reqParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatalln(err)
		return
	}

	user, err := cfg.db.GetUserById(context.Background(), reqParams.UserId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("error encountered could not retrieve user"))
		return
	}

	validChirpBody, err := validateChirpBody(reqParams.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Println(err)
		w.Write([]byte("invalid chirp body provided"))
		return
	}

	createChirpParams := database.CreateChirpParams{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Body:      validChirpBody,
		UserID:    user.ID,
	}

	createdChirp, err := cfg.db.CreateChirp(context.Background(), createChirpParams)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		w.Write([]byte("error encountered could not create chirp"))
		return
	}

	jsonData := chirpsResponseBody{
		ID:        createdChirp.ID,
		CreatedAt: createdChirp.CreatedAt,
		UpdatedAt: createdChirp.UpdatedAt,
		Body:      createdChirp.Body,
		UserId:    createdChirp.UserID,
	}

	jsonRes, err := json.Marshal(jsonData)
	if err != nil {
		log.Fatalln(err)
	}
	w.WriteHeader(201)
	w.Write(jsonRes)

}

func validateChirpBody(chirp string) (string, error) {
	if len(chirp) > 140 {
		return "", fmt.Errorf("chirp is too long")
	}

	unWantedWords := map[string]string{
		"kerfuffle": "****",
		"sharbert":  "****",
		"fornax":    "****",
	}

	reqBodyWords := strings.Split(chirp, " ")
	for i, word := range reqBodyWords {
		if hiddenWord, exists := unWantedWords[strings.ToLower(word)]; exists {
			reqBodyWords[i] = hiddenWord
		}
	}

	return strings.Join(reqBodyWords, " "), nil
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
