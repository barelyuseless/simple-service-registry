package internal

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

func StartAPI(address string, services ServiceStore) *http.Server {
	router := httprouter.New()

	router.GET("/health", logLatency(cors(handleGetHealth())))
	router.GET("/spec", logLatency(cors(handleGetSpec())))

	router.GET("/services", logLatency(cors(handleGetServices(services))))
	router.POST("/services", logLatency(cors(handlePostServices(services))))

	server := &http.Server{Addr: address, Handler: router}

	go server.ListenAndServe()

	return server
}

func handleGetHealth() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetSpec() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/x-yaml")
		http.ServeFile(w, r, "spec.yml")
	}
}

func handleGetServices(services ServiceStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		serviceSlice, err := services.getServices()
		if err != nil {
			log.Err(err).Msg("Error retrieving services from serviceStore")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var serviceData []byte
		if r.Header.Get("accept") == "application/json" {
			serviceData, err = json.Marshal(serviceSlice)
			if err != nil {
				log.Err(err).Msg("Error JSON encoding service data")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
		} else {
			serviceData, err = yaml.Marshal(serviceSlice)
			if err != nil {
				log.Err(err).Msg("Error YAML encoding service data")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/x-yaml")
		}

		_, err = w.Write(serviceData)
		if err != nil {
			log.Err(err).Msg("Error writing service data")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func handlePostServices(services ServiceStore) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		var service Service
		err := json.NewDecoder(r.Body).Decode(&service)
		if err != nil {
			log.Err(err).Msg("Error decoding new service json")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !service.validate() {
			log.Error().Msg("Invalid service submitted, rejecting")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err = services.addService(&service)
		if err != nil {
			log.Err(err).Msg("Error adding new service to serviceStore")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func logLatency(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		t0 := time.Now()
		handle(w, r, p)
		log.Info().Float64("latency", time.Since(t0).Seconds()).Str("url", r.RequestURI).Send()
	}
}

func cors(handle httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		handle(w, r, p)
	}
}
