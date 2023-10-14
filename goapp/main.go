package main

import (
	"github.com/emmanuelviniciusdev/imersao-fullcycle-opentelemetry/opentelemetry"
	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var tracer trace.Tracer

func main() {
	ot := opentelemetry.NewOpenTel()

	ot.ServiceName = "GoApp"
	ot.ServiceVersion = "0.1"
	ot.ExporterEndpoint = "http://localhost:9411/api/v2/spans"

	tracer = ot.GetTracer()

	router := mux.NewRouter()

	router.Use(otelmux.Middleware(ot.ServiceName))

	router.HandleFunc("/", homeHandler)

	http.ListenAndServe(":3000", router)
}

func homeHandler(writer http.ResponseWriter, request *http.Request) {
	ctx := baggage.ContextWithoutBaggage(request.Context())

	// Rotina 1: "process-file"
	ctx, processFile := tracer.Start(ctx, "process-file")

	time.Sleep(time.Millisecond * 100)

	processFile.End()

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// Rotina 2: "request-remote-json"
	ctx, httpCall := tracer.Start(ctx, "request-remote-json")

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:3001", nil)

	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		log.Fatalln(err)
	}

	time.Sleep(time.Millisecond * 300)

	httpCall.End()

	// Rotina 3: "render-content"
	ctx, renderContent := tracer.Start(ctx, "render-content")

	time.Sleep(time.Millisecond * 100)

	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(body))

	renderContent.End()
}
