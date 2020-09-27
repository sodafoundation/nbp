package main

import (
	"github.com/gorilla/mux"
	"github.com/sodafoundation/nbp/csi-plug-n-play/sidecars/soda-proxy/pkg/controller/profile"
	"log"
	"net/http"
)

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/getprofile/{id}", profile.GetProfile)
	log.Fatal(http.ListenAndServe("0.0.0.0:50029", myRouter))
}

func main() {
	handleRequests()
}
