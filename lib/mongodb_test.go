package mpmongodb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestGraphDefinition(t *testing.T) {
	var mongodb MongoDBPlugin

	graphdef := mongodb.GraphDefinition()
	if len(graphdef) != 2 {
		t.Errorf("GetTempfilename: %d should be 4", len(graphdef))
	}
}

func TestParse80(t *testing.T) {
	var mongodb MongoDBPlugin
	stub, err := os.ReadFile("testdata/stub8_0_0.json")
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	var v any
	err = json.Unmarshal(stub, &v)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	bsonStats, err := bson.Marshal(v)
	if err != nil {
		t.Errorf("Error: %s", err.Error())
	}
	var m bson.M
	decoder := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(bsonStats)))
	decoder.DefaultDocumentM()
	if err = decoder.Decode(&m); err != nil {
		t.Error(err)
	}

	stat, err := mongodb.parseStatus(m)
	fmt.Println(stat)
	assert.Nil(t, err)
	// Mongodb Stats
	assert.EqualValues(t, reflect.TypeOf(stat["opcounters_command"]).String(), "float64")
	assert.EqualValues(t, stat["opcounters_command"], 22)
	assert.EqualValues(t, stat["connections_current"], 3)
}

func TestMetricKeyPrefix(t *testing.T) {
	var m MongoDBPlugin
	prefix := m.MetricKeyPrefix()
	assert.Equal(t, "mongodb", prefix)

	m.KeyPrefix = "test"
	prefix = m.MetricKeyPrefix()
	assert.Equal(t, "test", prefix)
}

func TestLabelPrefix(t *testing.T) {
	var m MongoDBPlugin
	label := m.LabelPrefix()
	assert.Equal(t, "MongoDB", label)

	m.KeyPrefix = "test"
	label = m.LabelPrefix()
	assert.Equal(t, "Test", label)
}
