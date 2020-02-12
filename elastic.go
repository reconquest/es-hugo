package main

import (
	"os"

	"github.com/reconquest/karma-go"
	"github.com/reconquest/pkg/log"
	"github.com/zazab/zhash"
	"gopkg.in/resty.v1"
)

type Elastic struct {
	rest     *resty.Client
	config   *Config
	newIndex string
	filename string
}

func NewElastic(config *Config, newIndex, filename string) *Elastic {
	return &Elastic{
		rest:     resty.New().SetHostURL(config.Elastic),
		config:   config,
		newIndex: newIndex,
		filename: filename,
	}
}

func (elastic *Elastic) GetAliases() ([]string, error) {
	var response map[string]interface{}
	_, err := elastic.rest.R().
		SetResult(&response).
		Get("/" + elastic.config.Index + "/_alias/")
	if err != nil {
		return nil, err
	}

	keys := []string{}
	for key, _ := range response {
		keys = append(keys, key)
	}

	return keys, nil
}

func (elastic *Elastic) Bulk() error {
	file, err := os.Open(elastic.filename)
	if err != nil {
		return err
	}
	defer file.Close()

	log.Infof(nil, "writing index: %s", elastic.newIndex)

	_, err = elastic.rest.R().
		SetBody(file).
		SetHeader("Content-Type", "application/x-ndjson").
		Post("/_bulk")

	return err
}

func (elastic *Elastic) Alias() error {
	_, err := elastic.rest.R().
		Put("/" + elastic.newIndex + "/_alias/" + elastic.config.Index)
	return err
}

func (elastic *Elastic) DeleteAlias(name string) error {
	_, err := elastic.rest.R().
		Delete("/" + name + "/_alias/" + elastic.config.Index)
	return err
}

func (elastic *Elastic) DeleteIndex(name string) error {
	_, err := elastic.rest.R().Delete("/" + name)
	return err
}

func (elastic *Elastic) Prepare() error {
	err := elastic.Bulk()
	if err != nil {
		return karma.Format(
			err,
			"unable to insert new dataset",
		)
	}

	err = elastic.Alias()
	if err != nil {
		return karma.Format(
			err,
			"unable to create new alias",
		)
	}

	aliases, err := elastic.GetAliases()
	if err != nil {
		return err
	}

	for _, alias := range aliases {
		if alias == elastic.newIndex {
			continue
		}

		log.Infof(nil, "removing index and alias: %s", alias)

		err := elastic.DeleteAlias(alias)
		if err != nil {
			return karma.Format(
				err,
				"unable to delete alias: %s", alias,
			)
		}

		err = elastic.DeleteIndex(alias)
		if err != nil {
			return karma.Format(
				err,
				"unable to delete index: %s", alias,
			)
		}
	}

	return nil
}

func (elastic *Elastic) Search(query string) ([]map[string]interface{}, error) {
	body := map[string]interface{}{
		"query": map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":  "*" + query + "*",
				"fields": []string{"content"},
			},
		},
	}

	result := map[string]interface{}{}
	_, err := elastic.rest.R().
		SetBody(body).
		SetResult(&result).
		Post("/_search")
	if err != nil {
		return nil, err
	}

	hash := zhash.HashFromMap(result)

	entries, err := hash.GetMapSlice("hits", "hits")
	if err != nil {
		if zhash.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	sources := []map[string]interface{}{}
	for _, entry := range entries {
		sources = append(sources, entry["_source"].(map[string]interface{}))
	}

	return sources, nil
}
