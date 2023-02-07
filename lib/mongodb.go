package mpmongodb

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	mp "github.com/mackerelio/go-mackerel-plugin-helper"
	"github.com/mackerelio/golib/logging"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var logger = logging.GetLogger("metrics.plugin.mongodb")

func getFloatValue(s map[string]interface{}, keys []string) (float64, error) {
	var val float64
	sm := s
	var err error
	for i, k := range keys {
		if i+1 < len(keys) {
			switch sm[k].(type) {
			case bson.M:
				sm = sm[k].(bson.M)
			default:
				return 0, fmt.Errorf("Cannot handle as a hash for %s", k)
			}
		} else {
			val, err = strconv.ParseFloat(fmt.Sprint(sm[k]), 64)
			if err != nil {
				return 0, err
			}
		}
	}

	return val, nil
}

// MongoDBPlugin mackerel plugin for mongo
type MongoDBPlugin struct {
	URL       string
	Username  string
	Password  string
	Source    string
	KeyPrefix string
	Verbose   bool
}

func (m MongoDBPlugin) fetchStatus() (bson.M, error) {
	ctx := context.Background()
	auth := options.Credential{
		Username:   m.Username,
		Password:   m.Password,
		AuthSource: m.Source,
	}
	timeout := 10 * time.Second
	opts := options.Client().ApplyURI(m.URL).SetDirect(true).SetAuth(auth).SetTimeout(timeout)

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	var serverStatus bson.M
	command := bson.D{{Key: "serverStatus", Value: 1}}
	err = client.Database("admin").RunCommand(ctx, command).Decode(&serverStatus)
	if err != nil {
		return nil, err
	}
	if m.Verbose {
		str, err := json.Marshal(serverStatus)
		if err != nil {
			fmt.Println(fmt.Errorf("Marshaling error: %s", err.Error()))
		}
		fmt.Println(string(str))
	}
	return serverStatus, nil
}

// FetchMetrics interface for mackerelplugin
func (m MongoDBPlugin) FetchMetrics() (map[string]interface{}, error) {
	serverStatus, err := m.fetchStatus()
	if err != nil {
		return nil, err
	}
	return m.parseStatus(serverStatus)
}

func (m MongoDBPlugin) parseStatus(serverStatus bson.M) (map[string]interface{}, error) {
	stat := make(map[string]interface{})

	//Adapt to version 3.2 or higher.
	//Check in version 3.6.
	metricPlace := map[string][]string{
		"connections_current": {"connections", "current"},
		"opcounters_insert":   {"opcounters", "insert"},
		"opcounters_query":    {"opcounters", "query"},
		"opcounters_update":   {"opcounters", "update"},
		"opcounters_delete":   {"opcounters", "delete"},
		"opcounters_getmore":  {"opcounters", "getmore"},
		"opcounters_command":  {"opcounters", "command"},
	}

	for k, v := range metricPlace {
		val, err := getFloatValue(serverStatus, v)
		if err != nil {
			logger.Warningf("Cannot fetch metric %s: %s", v, err)
		}

		stat[k] = val
	}

	return stat, nil
}

// GraphDefinition interface for mackerelplugin
func (m MongoDBPlugin) GraphDefinition() map[string]mp.Graphs {
	labelPrefix := m.LabelPrefix()

	// Adapt to version 3.2 or higher.
	// Check in version 3.6.
	return map[string]mp.Graphs{
		"connections": {
			Label: labelPrefix + " Connections",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "connections_current", Label: "current"},
			},
		},
		"opcounters": {
			Label: labelPrefix + " opcounters",
			Unit:  "integer",
			Metrics: []mp.Metrics{
				{Name: "opcounters_insert", Label: "Insert", Diff: true, Type: "uint64"},
				{Name: "opcounters_query", Label: "Query", Diff: true, Type: "uint64"},
				{Name: "opcounters_update", Label: "Update", Diff: true, Type: "uint64"},
				{Name: "opcounters_delete", Label: "Delete", Diff: true, Type: "uint64"},
				{Name: "opcounters_getmore", Label: "Getmore", Diff: true, Type: "uint64"},
				{Name: "opcounters_command", Label: "Command", Diff: true, Type: "uint64"},
			},
		},
	}
}

const defaultPrefix = "mongodb"

// MetricKeyPrefix returns the metrics key prefix
func (m MongoDBPlugin) MetricKeyPrefix() string {
	if m.KeyPrefix == "" {
		m.KeyPrefix = defaultPrefix
	}
	return m.KeyPrefix
}

func (m MongoDBPlugin) LabelPrefix() string {
	return cases.Title(language.Und, cases.NoLower).String(strings.Replace(m.MetricKeyPrefix(), defaultPrefix, "MongoDB", -1))
}

// Do the plugin
func Do() {
	optHost := flag.String("host", "localhost", "Hostname")
	optPort := flag.String("port", "27017", "Port")
	optUser := flag.String("username", "", "Username")
	optPass := flag.String("password", os.Getenv("MONGODB_PASSWORD"), "Password")
	optSource := flag.String("source", "", "authenticationDatabase")
	optVerbose := flag.Bool("v", false, "Verbose mode")
	optTempfile := flag.String("tempfile", "", "Temp file name")
	optKeyPrefix := flag.String("metric-key-prefix", "", "Metric key prefix")
	flag.Parse()

	var mongodb MongoDBPlugin
	mongodb.Verbose = *optVerbose
	mongodb.URL = fmt.Sprintf("mongodb://%s", net.JoinHostPort(*optHost, *optPort))
	mongodb.Username = *optUser
	mongodb.Password = *optPass
	mongodb.Source = *optSource
	mongodb.KeyPrefix = *optKeyPrefix

	helper := mp.NewMackerelPlugin(mongodb)
	if *optTempfile != "" {
		helper.Tempfile = *optTempfile
	} else {
		helper.SetTempfileByBasename(fmt.Sprintf("mackerel-plugin-mongodb-%s-%s", *optHost, *optPort))
	}

	helper.Run()
}
