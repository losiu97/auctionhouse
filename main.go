package main

import (
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang-poc/app"
	"golang-poc/controllers"
	"golang-poc/dao"
	"golang-poc/p2p"
	"net/http"
	"os"
)

func main() {

	router := mux.NewRouter()

	router.HandleFunc("/user", controllers.CreateAccount).Methods("POST")
	router.HandleFunc("/user/me", controllers.GetCurrentUser).Methods("GET")
	router.HandleFunc("/user/login", controllers.Authenticate).Methods("POST")

	router.HandleFunc("/auction", controllers.CreateAuction).Methods("POST")
	router.HandleFunc("/auction/{id:[0-9]+}", controllers.GetAuctionById).Methods("GET")
	router.HandleFunc("/auction", controllers.GetAllAuctions).Methods("GET")

	router.HandleFunc("/auction/{id:[0-9]+}/file", controllers.CreateAuctionFile).Methods("POST")
	router.HandleFunc("/auction/{id:[0-9]+}/file/{fileId:[0-9]+}", controllers.GetAuctionFileById).Methods("GET")

	router.Use(app.JwtAuthentication) //attach JWT auth middleware

	router.NotFoundHandler = app.NotFoundHandler

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" //localhost
	}

	for _, auction := range dao.GetAllAuctions(0, "") {
		p2p.RebootHost(*auction)
	}

	err := http.ListenAndServe(":"+port, handlers.CORS(headersOk, originsOk, methodsOk)(router)) //Launch the app, visit localhost:8000
	if err != nil {
		fmt.Print(err)
	}
}
