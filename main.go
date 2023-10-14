package main

import (
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/oauth2"
	"log"
	"net/http"
)

var env, _ = godotenv.Read(".env")

var (
	azureADClientID     = env["AZURE_CLIENT_ID"]
	azureADClientSecret = env["AZURE_CLIENT_SECRET"]
	azureADTenantID     = env["AZURE_TENANT_ID"]
)

var oauth2Config = &oauth2.Config{
	ClientID:     azureADClientID,
	ClientSecret: azureADClientSecret,
	RedirectURL:  "http://localhost:8080/callback",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://login.microsoftonline.com/" + azureADTenantID + "/oauth2/v2.0/authorize",
		TokenURL: "https://login.microsoftonline.com/" + azureADTenantID + "/oauth2/v2.0/token",
	},
	Scopes: []string{"User.Read", "profile", "openid", "email"},
}

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		url := oauth2Config.AuthCodeURL("", oauth2.AccessTypeOffline)
		http.Redirect(w, r, url, http.StatusFound)
	})

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := oauth2Config.Exchange(context.TODO(), code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		userInfo, err := getUserInfo(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tokenJson, err := json.Marshal(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]interface{}{
			"token": tokenJson,
			"user":  userInfo,
		})
		if err != nil {
			return
		}

	})

	log.Fatal(http.ListenAndServe(":8080", nil))

}

func getUserInfo(token *oauth2.Token) (map[string]interface{}, error) {
	client := oauth2Config.Client(context.TODO(), token)
	resp, err := client.Get("https://graph.microsoft.com/v1.0/me") // Endpoint para obter informações do usuário
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}
