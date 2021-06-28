package main

import (
	"github.com/gorilla/mux"
	"github.com/sodafoundation/nbp/csi-plug-n-play/sidecars/soda-proxy/pkg/controller/profile"
	"github.com/sodafoundation/nbp/csi-plug-n-play/sidecars/soda-proxy/pkg/controller/snapshot"

	"log"
	"net/http"
)

func handleRequests() {
	sodaProxyRouter := mux.NewRouter().StrictSlash(true)
	sodaProxyRouter.HandleFunc("/getprofile/{id}", profile.GetProfile)
	sodaProxyRouter.HandleFunc("/snapshot/{id}", snapshot.CreateSnapshot)
	log.Fatal(http.ListenAndServe("0.0.0.0:50029", sodaProxyRouter))

}

func main() {
	handleRequests()
}
