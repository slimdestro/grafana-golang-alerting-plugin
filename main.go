package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/grafana/grafana-plugin-sdk-go/backend/alerting"
	"github.com/grafana/grafana-plugin-sdk-go/backend/alerting/eval"
	"github.com/spf13/viper"
)

func main() {
	// .env load
	loadConfig()
	alerting.RegisterPlugin(newAlertingPlugin())
}

type alertingPlugin struct{}

func newAlertingPlugin() *alertingPlugin {
	return &alertingPlugin{}
}

func (p *alertingPlugin) NewAlertQuery() alerting.Query {
	return &alerting.QueryModel{
		JSON: `{"selectedField": "` + viper.GetString("selectedField") + `", "condition": "` + viper.GetString("condition") + `", "threshold": ` + strconv.FormatFloat(viper.GetFloat64("threshold"), 'f', -1, 64) + `}`,
	}
}

func (p *alertingPlugin) Init() error {
	return nil
}

func (p *alertingPlugin) RunQuery(dsInfo *alerting.DataSourceInstanceSettings, queryModel *alerting.QueryModel) (alerting.Result, error) {
	csvFile, err := os.Open(viper.GetString("csvPath"))
	if err != nil {
		log.Fatal(err)
	}
	defer csvFile.Close()

	// CSV parsin
	reader := csv.NewReader(csvFile)
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	selectedField := viper.GetString("selectedField")
	condition := viper.GetString("condition")
	threshold := viper.GetFloat64("threshold")
	var alerts []eval.Alert
	for _, record := range records {
		value := record[selectedField]
		numericValue, err := parseFloat(value)
		if err != nil {
			log.Printf("Error parsing value: %v", err)
			continue
		}

		switch condition {
		case ">":
			if numericValue > threshold {
				alerts = append(alerts, alerting.Alert{
					State:   alerting.AlertStateAlerting,
					Message: fmt.Sprintf("Alert condition met for %s: %s %s %f", selectedField, value, condition, threshold),
				})
			}
		}
	}

	return alerting.Result{
		Alerts: alerts,
	}, nil
}

func parseFloat(s string) (float64, error) {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, err
	}
	return value, nil
}

func loadConfig() {
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading .env file: %s", err)
	}
}
