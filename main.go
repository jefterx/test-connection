package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

func speedTestHandler(w http.ResponseWriter, r *http.Request) {
	serverList, err := speedtest.FetchServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Escolhendo o primeiro servidor da lista
	targetServer := serverList[0]

	// Passando uma função de callback adequada para PingTest
	err = targetServer.PingTest(func(latency time.Duration) {
		fmt.Printf("Latência: %v\n", latency)
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = targetServer.DownloadTest()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = targetServer.UploadTest()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"ping":     targetServer.Latency,
		"download": targetServer.DLSpeed,
		"upload":   targetServer.ULSpeed,
	}

	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/speedtest", speedTestHandler)
	fmt.Println("Servidor rodando na porta 8080")
	http.ListenAndServe(":8080", nil)
}
