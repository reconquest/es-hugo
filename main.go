package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"time"

	"github.com/docopt/docopt-go"
	"github.com/reconquest/executil-go"
	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
	"github.com/reconquest/pkg/web"
)

var (
	version = "[manual build]"
	usage   = "es-hugo " + version + `

Requires kovetskiy/hugo-elasticsearch

Usage:
  es-hugo [options]
  es-hugo -h | --help
  es-hugo --version

Options:
  -c --config <path>  Path to config file. [default: /etc/es-hugo.conf]
  -h --help           Show this screen.
  --version           Show version.
`
)

type Handler struct {
	elastic *Elastic
}

func main() {
	args, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	log.Infof(karma.Describe("version", version), "started es-hugo")

	log.SetLevel(log.LevelDebug)

	config, err := getConfig(args["--config"].(string))
	if err != nil {
		log.Fatal(err)
	}

	log.Infof(nil, "generating dataset")

	newIndex, filename, err := generateDataset(config)
	if err != nil {
		log.Fatal(err)
	}

	elastic := NewElastic(config, newIndex, filename)

	err = elastic.Prepare()
	if err != nil {
		log.Fatal(err)
	}

	handler := &Handler{elastic: elastic}

	web := web.New()
	web.Use(cors)
	web.Get("/", web.ServeFunc(handler.Search))

	log.Infof(nil, "listening and serving")
	err = http.ListenAndServe(config.Listen, web)
	if err != nil {
		log.Fatal(err)
	}
}

func (handler *Handler) Search(ctx *web.Context) web.Status {
	param := ctx.GetQueryParam("query")
	if param == "" {
		return ctx.NotFound()
	}

	items, err := handler.elastic.Search(param)
	if err != nil {
		return ctx.InternalError(err, "unable to search for specified query")
	}

	err = json.NewEncoder(ctx.GetResponseWriter()).Encode(items)
	if err != nil {
		return ctx.InternalError(err, "unable to marshal error")
	}

	return ctx.OK()
}

func generateDataset(config *Config) (string, string, error) {
	index := config.Index + "_" + fmt.Sprint(time.Now().Unix())
	output := "es-hugo.json"

	cmd := exec.Command(
		"hugo-elasticsearch",
		"--input", config.Input,
		"--output", output,
		"--language", config.Language,
		"--delimiter", config.Delimiter,
		"--index-name", index,
	)

	_, _, err := executil.Run(cmd)

	return index, output, err
}

func cors(handler web.Handler) web.Handler {
	return func(context *web.Context) web.Status {
		origin := context.GetRequest().Header.Get("Origin")
		if origin == "" {
			// data, _ := httputil.DumpRequest(context.GetRequest(), true)
			// log.Errorf(nil, "Origin for some reason is empty: %s", string(data))
			origin = "*"
		}

		context.GetResponseWriter().Header().Add(
			"Access-Control-Allow-Origin",
			origin,
		)

		context.GetResponseWriter().Header().Add(
			"Access-Control-Allow-Methods",
			"GET, POST, OPTIONS, PUT, PATCH, DELETE",
		)

		//* (wildcard) The value "*" only counts as a special wildcard value for
		// requests without credentials (requests without HTTP cookies or HTTP
		// authentication information). In requests with credentials, it is treated
		// as the literal header name "*" without special semantics. Note that the
		// Authorization header can't be wildcarded and always needs to be listed
		// explicitly.
		context.GetResponseWriter().Header().Add(
			"Access-Control-Allow-Headers",
			"*",
		)

		return handler(context)
	}
}
