// +build !js !wasm

// Copyright (c) 2020 OSRAM; Licensed under the MIT license.
// This is the web server script of the application
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
)

type webHandler func(string, []byte) ([]byte, error)

type webRoute struct {
	Name    string
	Method  string
	Path    string
	Handler webHandler
}

type webErrorReply struct {
	Error  string `json:"error,omitempty"`
	Result string `json:"result,omitempty"`
}

func webHandlerWrapper(handler webHandler, name string, logger *log.Logger) http.HandlerFunc {
	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		logger.Printf("INFO: %s %s %s", request.Method, request.RequestURI, name)
		response.Header().Set("Content-Type", "application/json; charset=UTF-8")
		status := http.StatusOK
		var bufferOut []byte
		bufferIn, fail := ioutil.ReadAll(request.Body)
		if fail == nil {
			bufferOut, fail = handler(name, bufferIn)
		}
		if fail != nil {
			reply := webErrorReply{fail.Error(), string(bufferOut)}
			if bufferOut, fail = json.Marshal(&reply); fail != nil {
				logger.Printf("ERROR: Failed to marshal an error reply (%s)", fail)
				bufferOut = []byte(reply.Error)
			}
			status = http.StatusInternalServerError
		}
		response.WriteHeader(status)
		written, fail := response.Write(bufferOut)
		if fail != nil {
			logger.Printf("ERROR: Failed to write a reply (%s)", fail)
		} else if written != len(bufferOut) {
			logger.Printf("ERROR: Failed to write a full reply")
		}
	})
}

// Launches a web server for PHYTOFY RL
func webLaunch(port uint16, routes []webRoute, includeUI bool, logger *log.Logger) error {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		handler := webHandlerWrapper(route.Handler, route.Name, logger)
		router.Methods(route.Method).Path(route.Path).Name(route.Name).Handler(handler)
	}
	if includeUI {
		router.PathPrefix("/").Handler(http.FileServer(rice.MustFindBox("assets").HTTPBox()))
	}
	return http.ListenAndServe(fmt.Sprintf(":%d", port), router)
}
