// Package p contains an HTTP Cloud Function.
package p

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func ProcessQuest(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	var msg struct {
		Message string `json:"message"`
	}

	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body) // Log the request body

	if err := json.Unmarshal(body, &msg); err != nil {
		log.Fatalf("json.NewDecoder: Request %v", err)
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if msg.Message == "" {
		log.Println("No message?")
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	sha := hash(msg.Message)
	log.Printf("Msg: %s, sha: %s\n", string(body), sha)

	payload, err := createPayload(msg.Message)
	if err != nil {
		log.Fatalf("Error creating a payload: sha: %s, err: %v\n", sha, err)
		return
	}

	url := createUrl(sha)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("Error creating a request sha: %s err: %v\n", sha, err)
		return
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", os.Getenv("TOKEN")))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request %s, e: %v\n", sha, err)
		return
	}

	defer resp.Body.Close()
	cnt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading a response for request %+v sha %s, e: %v\n", req, sha, err)
		return
	}
	log.Printf("Request %s, response %+v, status %+v\n", sha, string(cnt), resp.StatusCode)
}

func hash(msg string) string {
	h := sha1.New()
	h.Write([]byte(msg))
	bs := h.Sum(nil)
	return hex.EncodeToString(bs)
}

func createPayload(msg string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"message":  fmt.Sprintf("new response at %v", time.Now().Unix()),
		"commiter": `{"name": "lacon-bot", "email": "10778553+lacon-bot@users.noreply.github.com"}`,
		"content":  base64.StdEncoding.EncodeToString([]byte(msg)),
	})
}

func createUrl(sha string) string {
	return fmt.Sprintf(
		"https://api.github.com/repos/%s/%s/contents/%s",
		os.Getenv("USER"),
		os.Getenv("REPO"),
		sha,
	)
}
