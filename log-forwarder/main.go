package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/andrewmarklloyd/pi-brew/internal/pkg/config"
	dd "github.com/andrewmarklloyd/pi-brew/internal/pkg/datadog"
)

type syslog struct {
	Identifier string `json:"SYSLOG_IDENTIFIER"`
	Message    string `json:"MESSAGE"`
	Error      error
}

var datadogClient dd.Client

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	datadogClient = dd.NewDatadogClient(conf.DatadogApiKey, conf.DatadogAppKey)

	go sendLogHearbeat()

	logChannel := make(chan syslog)
	go tailSystemdLogs(logChannel)
	for log := range logChannel {
		if log.Error != nil {
			fmt.Printf("error receiving logs from journalctl channel: %s\n", log.Error)
			break
		}

		// fmt.Println(log)
		err := sendLog(log.Message)
		if err != nil {
			fmt.Println("error sending logs:", err)
		}
	}
}

func tailSystemdLogs(ch chan syslog) error {
	cmd := exec.Command("journalctl", "-u", "pi-brew", "-f", "-n 0", "--output", "json")
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("creating command stdout pipe: %s", err)
	}

	scanner := bufio.NewScanner(cmdReader)
	go func() {
		for scanner.Scan() {
			var s syslog
			if err := json.Unmarshal([]byte(scanner.Text()), &s); err != nil {
				s.Error = fmt.Errorf("unmarshalling log: %s, original log text: %s", err, scanner.Text())
				ch <- s
				break
			}
			if s.Message != "" && s.Identifier != "systemd" && !strings.Contains(s.Message, "Logs begin at") {
				ch <- s
			}
		}
	}()

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting command: %s", err)
	}

	if err := cmd.Wait(); err != nil {
		close(ch)
		return fmt.Errorf("waiting for command: %s", err)
	}

	return nil
}

func filterOutLogs(log string) string {
	if strings.Contains(log, "CommandCompleteEP") {
		return ""
	}

	if strings.Contains(log, "Finished scanning") {
		return ""
	}

	if strings.Contains(log, "Scanning for") {
		return ""
	}

	return log
}

func sendLog(log string) error {
	if msg := filterOutLogs(log); msg == "" {
		return nil
	}

	body := []datadogV2.HTTPLogItem{
		{
			Ddsource: datadog.PtrString("pi-brew"),
			Ddtags:   datadog.PtrString("env:prod"),
			Hostname: datadog.PtrString("tiltpi"),
			Message:  log,
			Service:  datadog.PtrString("pi-brew"),
		},
	}
	ctx := datadog.NewDefaultContext(context.Background())
	configuration := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(configuration)
	api := datadogV2.NewLogsApi(apiClient)
	_, r, err := api.SubmitLog(ctx, body, *datadogV2.NewSubmitLogOptionalParameters().WithContentEncoding(datadogV2.CONTENTENCODING_DEFLATE))

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error when calling `LogsApi.SubmitLog`: %v\n", err)
		fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
		return err
	}

	return nil
}

func sendLogHearbeat() {
	metric := "log_forwarder_heartbeat"
	tags := map[string]string{}

	if err := datadogClient.PublishMetric(context.Background(), metric, tags); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		if err := datadogClient.PublishMetric(context.Background(), metric, tags); err != nil {
			fmt.Println(err)
		}
	}
}
