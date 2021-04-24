package main

import (
	"log"
	"os"
	//"fmt"
	"context"
	"time"
	"net/http"
	"os/signal"
	"github.com/gorilla/mux"
	"github.com/nicholasjackson/env"
	"github.com/go-openapi/runtime/middleware"
	input_registerHl "github.com/bybrisk/input-register-api/handlers"
)

func main() {
	/*port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}*/

	port := "80"
	var bindAddress = env.String("BIND_ADDRESS", false, ":"+port,"Bind address for the server")

	env.Parse()

	l := log.New(os.Stdout,"input-api",log.LstdFlags)
	
	//registering all handlers
	input_Register_Handler := input_registerHl.New_Input_Register(l)

	serveMux := mux.NewRouter()

	//Get Router
	getRouter := serveMux.Methods(http.MethodGet).Subrouter()
	/*getRouter.HandleFunc("/account/{id}",accountHandler.GetAccountDetail)
	getRouter.HandleFunc("/account/check/{username}",accountHandler.CheckAvailableUsername)*/

	//Post Router
	postRouter := serveMux.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/input/register/create", input_Register_Handler.Register_User)
	postRouter.HandleFunc("/input/register/business", input_Register_Handler.Subscribe_User)

	// handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"} 
	sh := middleware.Redoc(opts, nil)
	
	getRouter.Handle("/docs", sh)
	getRouter.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	//Create a new server
	s := http.Server{
		Addr: *bindAddress,			//configure the bind address
		Handler: serveMux,	//set the default handler
		ErrorLog: l,	//set the logger for the server
		ReadTimeout: 5 * time.Second,	//max time for read request for the client
		WriteTimeout: 10 * time.Second,	//Max time for write response for client
		IdleTimeout: 120 * time.Second, //Max timme for connection using TCP Keep alive 
	}

	//start the server
	go func () {
		l.Println("Starting bybrisk input-api server")
		err := s.ListenAndServe()	
		if err!=nil {
			l.Printf("Error starting bybrisk input-api server: %s\n",err)
			os.Exit(1)
		}
	}()

	//trap sigterm or intrupt and gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, os.Kill)

	sig := <-c
	l.Println("Recieved terminate, graceful shutdown bybrisk input-api server", sig)

	tc,_ := context.WithTimeout(context.Background(), 30*time.Second)
	s.Shutdown(tc)

}