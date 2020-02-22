package internal

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"time"
)

var basicHTMLTemplate = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8">
	<title>Simple Service Registry</title>
	<style>
		table, th, td {border: 1px solid black;}
		th, td {padding: 4px;}
	</style>
  </head>
  <body>
<:CONTENT:>
  </body>
</html>`

type rowData struct {
	ServiceURL     string
	SpecURL        string
	Description    string
	Tags           string
	HealthcheckURL string
	Available      bool
	LastChecked    string
	LastAvailable  string
}

func rowDataFromService(service *Service) *rowData {
	lastChecked := "never"
	if !service.LastChecked.IsZero() {
		lastChecked = fmt.Sprintf("%s ago", time.Since(service.LastChecked))
	}

	lastAvailable := "never"
	if !service.LastAvailable.IsZero() {
		lastAvailable = fmt.Sprintf("%s ago", time.Since(service.LastAvailable))
	}

	return &rowData{
		ServiceURL:     service.ServiceURL,
		SpecURL:        service.SpecURL,
		Description:    service.Description,
		Tags:           strings.Join(service.Tags, ","),
		HealthcheckURL: service.HealthcheckURL,
		Available:      service.Available,
		LastChecked:    lastChecked,
		LastAvailable:  lastAvailable,
	}
}

func createServiceHTML(services []*Service) ([]byte, error) {
	var rows bytes.Buffer
	t, _ := template.New("tableRow").Parse(
		`          <tr>
			<td><a href={{.ServiceURL}}>{{.ServiceURL}}</a></td>
			<td><a href={{.SpecURL}}>{{.SpecURL}}</a></td>
			<td>{{.Description}}</td>
			<td>{{.Tags}}</td>
			<td>{{.Available}}</td>
			<td>{{.LastChecked}}</td>
			<td>{{.LastAvailable}}</td>
          </tr>`)

	rows.WriteString("    <h1>simple service registry</h1>\n")
	rows.WriteString("    <table>\n")
	rows.WriteString("      <tr><th>Base URL</th><th>Spec URL</th><th>Description</th><th>Tags</th><th>Available</th><th>Last Checked</th><th>Last Available</th></tr>\n")

	for _, service := range services {

		t.Execute(&rows, rowDataFromService(service))
	}

	rows.WriteString("\n    </table>\n")

	return []byte(strings.Replace(basicHTMLTemplate, "<:CONTENT:>", rows.String(), -1)), nil
}
