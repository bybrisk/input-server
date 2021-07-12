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
	input_convoHl "github.com/bybrisk/input-convo-starter-api/handlers"
	input_actionHl "github.com/bybrisk/input-action-api/handlers"
	input_archiveHl "github.com/bybrisk/input-archive-api/handlers"
	input_schemaHl "github.com/bybrisk/input-schema-api/handlers"
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
	input_Convo_Starter_Handler := input_convoHl.New_Input_Convo(l)
	input_Action_Handler := input_actionHl.New_Input_Action(l)
	input_Archive_Handler := input_archiveHl.New_Input_Archive(l)
	input_Schema_Handler := input_schemaHl.New_Input_Schema(l)


	serveMux := mux.NewRouter()

	//Get Router
	getRouter := serveMux.Methods(http.MethodGet).Subrouter()
	getRouter.HandleFunc("/user/{PhoneNumber}",input_Register_Handler.GetUserIDByPhone)
	/*getRouter.HandleFunc("/account/check/{username}",accountHandler.CheckAvailableUsername)*/

	//Post Router
	postRouter := serveMux.Methods(http.MethodPost).Subrouter()
	postRouter.HandleFunc("/user/register", input_Register_Handler.Register_User)
	postRouter.HandleFunc("/user/subscribe", input_Register_Handler.Subscribe_User)

	postRouter.HandleFunc("/commence/get", input_Convo_Starter_Handler.Get_Conversation_Initials)
	postRouter.HandleFunc("/commence/action", input_Convo_Starter_Handler.Get_Action_Handlers)

	postRouter.HandleFunc("/action/order", input_Action_Handler.Make_Action_Order)

	postRouter.HandleFunc("/archive/create", input_Archive_Handler.Create_Archive)
	postRouter.HandleFunc("/archive/get/page", input_Archive_Handler.Get_Archive_Pagewise)
	postRouter.HandleFunc("/schema/create", input_Schema_Handler.Create_Bot_Schema)

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