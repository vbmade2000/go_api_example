package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nathanmalishev/taskmanager/common"
	"github.com/nathanmalishev/taskmanager/controllers"
	"github.com/nathanmalishev/taskmanager/models"
	"github.com/urfave/negroni"
)

func dummy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Yet to be implemented")
		return
	})
}

// With db wraps each controller that needs the db with a new session
// this is important to handle requests concurrently
func WithDb(store *models.DataStore, fn func(*models.DataStore, http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newStore := store.GetStore() // when we return the store, we copy the session
		defer newStore.Close()       // must close the session, or we will leave connections open
		fn(newStore, w, r)
	})
}

func InitRoutes(store *models.DataStore, authModule *common.Auth) http.Handler {
	router := mux.NewRouter().StrictSlash(false)

	/* User routes */
	router.Handle("/users/register", WithDb(store, controllers.Register)).Methods("POST")
	router.Handle("/users/login", dummy()).Methods("POST")

	/* Task routes  */
	taskRouter := mux.NewRouter().StrictSlash(false)
	taskRouter.Handle("/tasks", WithDb(store, controllers.GetAllTasks)).Methods("GET")
	taskRouter.Handle("/tasks/{id}", dummy()).Methods("GET")
	taskRouter.Handle("/tasks/{id}", dummy()).Methods("DELETE")
	taskRouter.Handle("/tasks", dummy()).Methods("POST")
	taskRouter.Handle("/tasks/{id}", dummy()).Methods("PUT")
	taskRouter.Handle("/tasks/users/{id}", dummy()).Methods("GET")

	/* Notes routes  */
	notesRouter := mux.NewRouter().StrictSlash(false)
	notesRouter.Handle("/notes", dummy()).Methods("GET")
	notesRouter.Handle("/notes/{id}", dummy()).Methods("GET")
	notesRouter.Handle("/notes/{id}", dummy()).Methods("DELETE")
	notesRouter.Handle("/notes", dummy()).Methods("POST")
	notesRouter.Handle("/notes/{id}", dummy()).Methods("PUT")
	notesRouter.Handle("/notes/tasks/{id}", dummy()).Methods("GET")

	/* middleware */
	commonMidleware := negroni.New(
		negroni.NewLogger(),
	) // will add auth middleware to these routes soon
	router.PathPrefix("/notes").Handler(negroni.New(
		common.WithAuth(authModule),
		negroni.Wrap(notesRouter),
	))
	router.PathPrefix("/tasks").Handler(negroni.New(
		//common.WithAuth(authModule),
		negroni.Wrap(taskRouter),
	))
	// common wraps all routes in default middleware
	// this includes all API hits
	commonMidleware.UseHandler(router)

	return commonMidleware
}
