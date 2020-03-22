package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gopkg.in/oauth2.v3/store"
)

var (
	clientStore = store.NewClientStore()
	manager = manage.NewDefaultManager()
	srv = server.NewDefaultServer(manager)
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Home"))
}

func protected(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Protected"))
}

func cred(w http.ResponseWriter, r *http.Request) {
	clientId := uuid.New().String()[:8]
	clientSecret := uuid.New().String()[:8]
	err := clientStore.Set(clientId, &models.Client{
		ID:		clientId,
		Secret: clientSecret,
		Domain: "http://localhost:9094",
	})
	if err != nil {
		log.Println(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"CLIENT_ID": clientId, "CLIENT_SECRET": clientSecret})
}

func token(w http.ResponseWriter, r *http.Request) {
	srv.HandleTokenRequest(w, r)
}

func validateToken(f http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	   _, err := srv.ValidationBearerToken(r)
	   if err != nil {
		  http.Error(w, err.Error(), http.StatusBadRequest)
		  return
	   }
 
	   f.ServeHTTP(w, r)
	})
}

func main() {

	manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	manager.MustTokenStorage(store.NewMemoryTokenStore())
	manager.MapClientStorage(clientStore)
	manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

	srv.SetAllowGetAccessRequest(true)
   	srv.SetClientInfoHandler(server.ClientFormHandler)
	
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func (re *errors.Response) {
		log.Println("Reponse Error:", re.Error.Error())
		return
	})

	homeHandler := http.HandlerFunc(home)
	ProtectedHandler := http.HandlerFunc(protected)
	CredHandler := http.HandlerFunc(cred)
	TokenHandler := http.HandlerFunc(token)

	http.Handle("/", homeHandler)
	http.Handle("/protected", validateToken(ProtectedHandler))
	http.Handle("/credentials", CredHandler)
	http.Handle("/token", TokenHandler)
	log.Fatal(http.ListenAndServe(":9096", nil))
}