package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type result struct {
	ServiceName string `json:"serviceName"`
	Status      string `json:"status"`
	Reason      string `json:"reason"`
}

const (
	notaryServiceName = "Developer ID Notary Service"
)

func status(serviceName string) string {
	if serviceName == "" {
		serviceName = notaryServiceName
	}

	log.Printf("text: %s", serviceName)

	resp, err := http.Get("https://www.apple.com/support/systemstatus/data/developer/system_status_en_US.js")
	if err != nil && resp.StatusCode != http.StatusOK {
		return formatResult(serviceName, "na", "failed to get status")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return formatResult(serviceName, "na", "failed to read status")
	}

	re := regexp.MustCompile(`\((.*?)\)`)
	rs := re.FindStringSubmatch(string(body))

	var response response
	if err := json.Unmarshal([]byte(rs[1]), &response); err != nil {
		return formatResult(serviceName, "na", "failed to parse status")
	}

	for _, s := range response.Services {
		if strings.EqualFold(serviceName, s.ServiceName) {
			if len(s.Events) == 0 {
				return formatResult(serviceName, "up", "service is up")
			}
			firstEvent := s.Events[0]
			startTime, err := time.Parse("01/02/2006 15:04 MST", firstEvent.StartDate)
			if err != nil {
				return formatResult(serviceName, "na",
					fmt.Sprintf("failed to parse start time: %s", err.Error()))
			}
			endTime, err := time.Parse("01/02/2006 15:04 MST", firstEvent.EndDate)
			if err != nil {
				return formatResult(serviceName, "na",
					fmt.Sprintf("failed to parse start time: %s", err.Error()))
			}
			now := time.Now()

			log.Printf("start: %s", startTime)
			log.Printf("end  : %s", endTime)
			log.Printf("now  : %s", now)

			if now.After(startTime) && now.Before(endTime) {
				return formatResult(serviceName, s.Events[0].EventStatus, s.Events[0].Message)
			}
			return formatResult(serviceName, "up", "service is up")
		}
	}
	return formatResult(serviceName, "na", "service not found")
}

type response struct {
	DrPost    bool      `json:"drpost"`
	DrMessage string    `json:"drMessage"`
	Services  []service `json:"services"`
}

type service struct {
	RedirectURL string  `json:"redirectUrl"`
	ServiceName string  `json:"serviceName"`
	Events      []event `json:"events"`
}

type event struct {
	EventStatus string `json:"eventStatus"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Message     string `json:"message"`
}

func formatResult(service, status, reason string) string {
	r := result{
		ServiceName: service,
		Status:      status,
		Reason:      reason,
	}
	b, _ := json.MarshalIndent(r, "", "  ")
	return fmt.Sprintf("%s\n", string(b))
}
