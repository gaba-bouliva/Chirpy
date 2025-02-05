package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits *atomic.Int32
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
	cfg.fileserverHits.Swap(0)
	w.Header().Set("Content-Type", "text/plain; charset-utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("Hits: %d", cfg.fileserverHits.Load())))
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	apiCfg := apiConfig{fileserverHits: &atomic.Int32{}}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /admin/metrics", apiCfg.countHits)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", validate_chirp)
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	fmt.Printf("running server on localhost%s", server.Addr)
	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
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
