package config

import (
	"encoding/json"
	"io/ioutil"
	"strconv"

	"github.com/Sirupsen/logrus"
)

var log = logrus.WithFields(logrus.Fields{"app": "fullerite", "pkg": "config"})

// Config type holds the global Fullerite configuration.
type Config struct {
	Prefix                string                            `json:"prefix"`
	Interval              interface{}                       `json:"interval"`
	DiamondCollectorsPath string                            `json:"diamond_collectors_path"`
	DiamondCollectors     map[string]map[string]interface{} `json:"diamond_collectors"`
	Handlers              map[string]map[string]interface{} `json:"handlers"`
	Collectors            map[string]map[string]interface{} `json:"collectors"`
	DefaultDimensions     map[string]string                 `json:"defaultDimensions"`
	InternalMetricsPort   string                            `json:"internalMetricsPort"`
}

// ReadConfig reads a fullerite configuration file
func ReadConfig(configFile string) (c Config, e error) {
	log.Info("Reading configuration file at ", configFile)
	contents, e := ioutil.ReadFile(configFile)
	if e != nil {
		log.Error("Config file error: ", e)
		return c, e
	}
	err := json.Unmarshal(contents, &c)
	if err != nil {
		log.Error("Invalid JSON in config: ", err)
		return c, err
	}
	return c, nil
}

// GetAsFloat parses a string to a float or returns the float if float is passed in
func GetAsFloat(value interface{}, defaultValue float64) (result float64) {
	result = defaultValue

	switch value.(type) {
	case string:
		fromString, err := strconv.ParseFloat(value.(string), 64)
		if err != nil {
			log.Warn("Failed to convert value", value, "to a float64. Falling back to default", defaultValue)
			result = defaultValue
		} else {
			result = fromString
		}
	case float64:
		result = value.(float64)
	}

	return
}

// GetAsInt parses a string/float to an int or returns the int if int is passed in
func GetAsInt(value interface{}, defaultValue int) (result int) {
	result = defaultValue

	switch value.(type) {
	case string:
		fromString, err := strconv.ParseInt(value.(string), 10, 64)
		if err == nil {
			result = int(fromString)
		} else {
			log.Warn("Failed to convert value", value, "to an int")
		}
	case int:
		result = value.(int)
	case int32:
		result = int(value.(int32))
	case int64:
		result = int(value.(int64))
	case float64:
		result = int(value.(float64))
	}

	return
}
