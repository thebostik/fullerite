package internalserver

import (
	"fullerite/config"
	"fullerite/handler"

	"encoding/json"
	"net/http"
	"runtime"

	l "github.com/Sirupsen/logrus"
)

const (
	DefaultPort        = 19090
	DefaultMetricsPath = "/metrics"
)

type internalServer struct {
	log      *l.Entry
	handlers *[]handler.Handler
}

type ResponseFormat struct {
	Memory   handler.InternalMetrics
	Handlers map[string]handler.InternalMetrics
}

func RunServer(cfg *config.Config, handlers *[]handler.Handler) {
	srv := internalServer{}
	srv.log = l.WithFields(l.Fields{"app": "fullerite", "pkg": "internal_server"})
	srv.handlers = handlers

	// port := config.GetAsInt(cfg.InternalMetricsPort, DefaultPort)
	http.HandleFunc("/"+DefaultMetricsPath, srv.handleInternalMetricsRequest)
}

// this is what services the request. The response will be JSON formatted like this:
// 	{
// 		"memory": {
// 			"counters": {
//				"TotalAlloc": 43.2,
//				"NumGoRoutine": 12.3
//			},
//			"gauges": {
//				"Alloc": 23.4,
//				"Sys": 12.43
//			}
//		},
//		"handlers": {
//			"somehandler": {
//				"counters": {
//					"totalEmissions": 12332,
//				},
//				"gauges": {
//					"averageEmissionTiming": 1.34,
//				}
//			}
//		}
//	}
//
func (srv internalServer) handleInternalMetricsRequest(writer http.ResponseWriter, req *http.Request) {
	srv.log.Debug("Starting to handle request for internal metrics, checking ", len(*srv.handlers), " handlers")

	// rspString := srv.buildResponse()

	// TODO
}

// responsible for querying each handler and serializing the total response
func (srv internalServer) buildResponse() *[]byte {
	memoryStats := getMemoryStats()

	handlerStats := make(map[string]handler.InternalMetrics)
	for _, inst := range *srv.handlers {
		handlerStats[inst.Name()] = inst.InternalMetrics()
	}

	rsp := ResponseFormat{}
	rsp.Handlers = handlerStats
	rsp.Memory = *memoryStats

	asString, err := json.Marshal(rsp)
	if err != nil {
		srv.log.Warn("Failed to marshal response ", rsp, " because of error ", err)
	}

	return &asString
}

// gets the actual memory stats
func memoryStats() *runtime.MemStats {
	stats := new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	return stats
}

// converts the memory stats to a map. The response is in the form like this: {counters: [], gauges: []}
func getMemoryStats() *handler.InternalMetrics {
	m := memoryStats()

	counters := map[string]float64{
		"NumGoroutine": float64(runtime.NumGoroutine()),
		"TotalAlloc":   float64(m.TotalAlloc),
		"Lookups":      float64(m.Lookups),
		"Mallocs":      float64(m.Mallocs),
		"Frees":        float64(m.Frees),
		"PauseTotalNs": float64(m.PauseTotalNs),
		"NumGC":        float64(m.NumGC),
	}

	gauges := map[string]float64{
		"Alloc":        float64(m.Alloc),
		"Sys":          float64(m.Sys),
		"HeapAlloc":    float64(m.HeapAlloc),
		"HeapSys":      float64(m.HeapSys),
		"HeapIdle":     float64(m.HeapIdle),
		"HeapInuse":    float64(m.HeapInuse),
		"HeapReleased": float64(m.HeapReleased),
		"HeapObjects":  float64(m.HeapObjects),
		"StackInuse":   float64(m.StackInuse),
		"StackSys":     float64(m.StackSys),
		"MSpanInuse":   float64(m.MSpanInuse),
		"MSpanSys":     float64(m.MSpanSys),
		"MCacheInuse":  float64(m.MCacheInuse),
		"MCacheSys":    float64(m.MCacheSys),
		"BuckHashSys":  float64(m.BuckHashSys),
		"GCSys":        float64(m.GCSys),
		"OtherSys":     float64(m.OtherSys),
		"NextGC":       float64(m.NextGC),
		"LastGC":       float64(m.LastGC),
	}

	rsp := handler.InternalMetrics{
		Counters: counters,
		Gauges:   gauges,
	}
	return &rsp
}
