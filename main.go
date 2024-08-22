package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/showwin/speedtest-go/speedtest"
)

func speedTestHandler(w http.ResponseWriter, r *http.Request) {
	// Obtendo a lista de servidores
	serverList, err := speedtest.FetchServers()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Limitar o número de servidores a serem testados
	const maxServers = 3
	if len(serverList) > maxServers {
		serverList = serverList[:maxServers]
	}

	// Pingando os servidores em paralelo
	var wg sync.WaitGroup
	for _, server := range serverList {
		wg.Add(1)
		go func(s *speedtest.Server) {
			defer wg.Done()
			_ = s.PingTest(nil)
		}(server)
	}
	wg.Wait()

	// Ordenar os servidores pela latência
	sort.Slice(serverList, func(i, j int) bool {
		return serverList[i].Latency < serverList[j].Latency
	})

	// Escolhendo o servidor com menor latência
	targetServer := serverList[0]

	var totalPing, totalDownload, totalUpload float64
	numTests := 1 // Reduzindo para 1 teste para acelerar

	// Medindo o tempo de execução
	start := time.Now()

	for i := 0; i < numTests; i++ {
		err = targetServer.PingTest(func(latency time.Duration) {
			totalPing += float64(latency.Milliseconds())
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
		totalDownload += float64(targetServer.DLSpeed) // Já em bps

		err = targetServer.UploadTest()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		totalUpload += float64(targetServer.ULSpeed)
	}

	// Time de execução
	duration := time.Since(start)
	fmt.Printf("Tempo de execução: %v\n", duration)

	// Calculando as médias
	avgPing := totalPing / float64(numTests)
	avgDownload := (totalDownload / float64(numTests)) / (8 * 1024 * 1024) // Convertendo para MB/s
	avgUpload := (totalUpload / float64(numTests)) / (8 * 1024 * 1024)     // Convertendo para MB/s

	// Formatando os resultados com 2 casas decimais
	result := map[string]interface{}{
		"ping":     fmt.Sprintf("%.2f ms", avgPing),
		"download": fmt.Sprintf("%.2f MB/s", avgDownload),
		"upload":   fmt.Sprintf("%.2f MB/s", avgUpload),
		"time":     fmt.Sprintf("%.2f s", duration.Seconds()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/speedtest", speedTestHandler)
	fmt.Println("Servidor rodando na porta 8080")
	http.ListenAndServe(":8080", nil)
}
