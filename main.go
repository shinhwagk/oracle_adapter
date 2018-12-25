package main

import (
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/golang/snappy"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/prompb"
)

type noOpWriter struct{}

type writer interface {
	Write(samples model.Samples) error
	Name() string
}

// type reader interface {
// 	Read(req *prompb.ReadRequest) (*prompb.ReadResponse, error)
// 	Name() string
// 	HealthCheck() error
// }

func (no *noOpWriter) Name() string {
	return "noopWriter"
}

func protoToSamples(req *prompb.WriteRequest) model.Samples {
	var samples model.Samples
	for _, ts := range req.Timeseries {
		metric := make(model.Metric, len(ts.Labels))
		for _, l := range ts.Labels {
			metric[model.LabelName(l.Name)] = model.LabelValue(l.Value)
		}

		for _, s := range ts.Samples {
			samples = append(samples, &model.Sample{
				Metric:    metric,
				Value:     model.SampleValue(s.Value),
				Timestamp: model.Time(s.Timestamp),
			})
		}
	}
	return samples
}

func sendSamples(w writer, samples model.Samples) error {
	begin := time.Now()
	err := w.Write(samples)
	duration := time.Since(begin).Seconds()
	// if err != nil {
	// 	failedSamples.WithLabelValues(w.Name()).Add(float64(len(samples)))
	// 	return err
	// }
	// sentSamples.WithLabelValues(w.Name()).Add(float64(len(samples)))
	// sentBatchDuration.WithLabelValues(w.Name()).Observe(duration)
	return nil
}

func write(writer writer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compressed, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("msg", "Read error", "err", err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		reqBuf, err := snappy.Decode(nil, compressed)
		if err != nil {
			log.Error("msg", "Decode error", "err", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req prompb.WriteRequest
		if err := proto.Unmarshal(reqBuf, &req); err != nil {
			log.Error("msg", "Unmarshal error", "err", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		samples := protoToSamples(&req)
		// receivedSamples.Add(float64(len(samples)))

		err = sendSamples(writer, samples)
		if err != nil {
			log.Warn("msg", "Error sending samples to remote storage", "err", err, "storage", writer.Name(), "num_samples", len(samples))
		}

		counter, err := sentSamples.GetMetricWithLabelValues(writer.Name())
		if err != nil {
			log.Warn("msg", "Couldn't get a counter", "labelValue", writer.Name(), "err", err)
		}
		writeThroughtput.SetCurrent(getCounterValue(counter))

		select {
		case d := <-writeThroughtput.Values:
			log.Info("msg", "Samples write throughput", "samples/sec", d)
		default:
		}
	})
}

func buildClients(cfg *config) writer {
	pgClient := pgprometheus.NewClient(&cfg.pgPrometheusConfig)
	if cfg.readOnly {
		return &noOpWriter{}, pgClient
	}
	return pgClient
}

func main() {
	http.Handle(cfg.telemetryPath, prometheus.Handler())
	log.Init(cfg.logLevel)

	writer, reader := buildClients(cfg)

	http.Handle("/write", timeHandler("write", write(writer)))
	http.HandleFunc("/read", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("read")) })
	http.Handle("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("healthz")) })

	log.Info("msg", "Starting up...")
	log.Info("msg", "Listening", "addr", cfg.listenAddr)

	err := http.ListenAndServe(cfg.listenAddr, nil)

	if err != nil {
		log.Error("msg", "Listen failure", "err", err)
		os.Exit(1)
	}
}
