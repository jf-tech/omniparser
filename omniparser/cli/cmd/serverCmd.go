package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/spf13/cobra"

	"github.com/jf-tech/omniparser/jsons"
	"github.com/jf-tech/omniparser/omniparser"
	"github.com/jf-tech/omniparser/omniparser/transformctx"
)

var (
	serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Launches op into HTTP server mode that does transform on its REST API.",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			doServer()
		},
	}
	port int
)

func init() {
	serverCmd.Flags().IntVarP(&port, "port", "p", 8080, "the listening HTTP port")
}

const (
	contentTypeHeader = "Content-Type"
	contentTypeJSON   = "application/json"
)

func doServer() {
	transformRouter := chi.NewRouter()
	transformRouter.Use(middleware.RealIP)
	transformRouter.Use(middleware.AllowContentType(contentTypeJSON))
	transformRouter.Post("/", httpPostTransform)

	rootRouter := chi.NewRouter()
	rootRouter.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.FileServer(http.Dir(filepath.Join(serverCmdDir(), "web"))).ServeHTTP(w, req)
	})
	rootRouter.Mount("/transform", transformRouter)

	envPort, found := os.LookupEnv("PORT")
	if found {
		var err error
		port, err = strconv.Atoi(envPort)
		if err != nil {
			panic(err)
		}
	}
	log.Printf("Listening on port %d ...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), rootRouter))
}

func serverCmdDir() string {
	_, filename, _, _ := runtime.Caller(1)
	absDir, _ := filepath.Abs(filepath.Dir(filename))
	return absDir
}

type req struct {
	Schema     string            `json:"schema"`
	Input      string            `json:"input"`
	Properties map[string]string `json:"properties"`
}

func httpPostTransform(w http.ResponseWriter, r *http.Request) {
	log.Printf("Serving request from %s ... ", r.RemoteAddr)

	writeError := func(msg string, code int) {
		http.Error(w, msg, code)
		log.Print(code)
	}
	writeBadRequest := func(msg string) {
		writeError(msg, http.StatusBadRequest)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(fmt.Sprintf("bad request: unable to read request body. err: %s", err))
		return
	}
	var req req
	err = json.Unmarshal(b, &req)
	if err != nil {
		writeBadRequest(fmt.Sprintf("bad request: invalid request body. err: %s", err))
		return
	}
	p, err := omniparser.NewParser("test-schema", strings.NewReader(req.Schema))
	if err != nil {
		writeBadRequest(fmt.Sprintf("bad request: invalid schema. err: %s", err))
		return
	}
	op, err := p.GetTransformOp("test-input", strings.NewReader(req.Input),
		&transformctx.Ctx{ExternalProperties: req.Properties})
	if err != nil {
		writeBadRequest(fmt.Sprintf("bad request: unable to init transform. err: %s", err))
		return
	}
	var records []string
	for op.Next() {
		b, err := op.Read()
		if err != nil {
			writeBadRequest(fmt.Sprintf("bad request: transform failed. err: %s", err))
			return
		}
		records = append(records, string(b))
	}
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(jsons.BPJ("[" + strings.Join(records, ",") + "]")))
	log.Print(http.StatusOK)
	log.Print(jsons.BPM(req))
}
