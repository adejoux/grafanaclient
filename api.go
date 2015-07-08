// Copyright © 2015 Alain Dejoux <adejoux@djouxtech.net>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package grafanaclient provide a simple API to manage Grafana 2.0 DataSources and Dashboards in Go.
// It's using Grafana 2.0 REST API.
package grafanaclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/naoina/toml"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"reflect"
)

// GrafanaError is a error structure to handle error messages in this library
type GrafanaError struct {
	Code        int
	Description string
}

// A GrafanaMessage contains the json error message received when http request failed
type GrafanaMessage struct {
	Message string `json:"message"`
}

// Error generate a text error message.
// If Code is zero, we know it's not a http error.
func (h GrafanaError) Error() string {
	if h.Code != 0 {
		return fmt.Sprintf("HTTP %d: %s", h.Code, h.Description)
	}
	return fmt.Sprintf("ERROR: %s", h.Description)
}

// Session contains user credentials, url and a pointer to http client session.
type Session struct {
	client   *http.Client
	User     string
	Password string
	url      string
}

// A Login contains the json structure of Grafana authentication request
type Login struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// A DataSource contains the json structure of Grafana DataSource
type DataSource struct {
	ID                int    `json:"Id"`
	OrgID             int    `json:"orgId"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Access            string `json:"access"`
	URL               string `json:"url"`
	Password          string `json:"password"`
	User              string `json:"user"`
	Database          string `json:"database"`
	BasicAuth         bool   `json:"basicAuth"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
	IsDefault         bool   `json:"isDefault"`
}

// A DashboardUploader encapsulates a complete Dashboard
type DashboardUploader struct {
	Dashboard Dashboard `json:"dashboard"`
	Overwrite bool      `json:"overwrite"`
}

// A DashboardResult contains the response from Grafana when requesting a Dashboard.
// It contains the Dashboard itself and the meta data.
type DashboardResult struct {
	Meta  Meta      `json:"meta"`
	Model Dashboard `json:"model"`
}

// A Meta contains a Dashboard metadata.
type Meta struct {
	Created    string `json:"created"`
	Expires    string `json:"expires"`
	IsHome     bool   `json:"isHome"`
	IsSnapshot bool   `json:"isSnapshot"`
	IsStarred  bool   `json:"isStarred"`
	Slug       string `json:"slug"`
}

// A Dashboard contains the Dashboard structure.
type Dashboard struct {
	Editable        bool          `json:"editable"`
	HideControls    bool          `json:"hideControls"`
	ID              int           `json:"id"`
	OriginalTitle   string        `json:"originalTitle"`
	Refresh         bool          `json:"refresh"`
	Annotations     Annotation    `json:"annotations"`
	SchemaVersion   int           `json:"schemaVersion"`
	SharedCrosshair bool          `json:"sharedCrosshair"`
	Style           string        `json:"style"`
	Templating      Template      `json:"templating"`
	Tags            []interface{} `json:"tags"`
	GTime           GTime         `json:"time"`
	Rows            []Row         `json:"rows" toml:"row"`
	Title           string        `json:"title"`
	Version         int           `json:"version"`
	Timezone        string        `json:"timezone"`
}

func (d *Dashboard) UnmarshalTOML([]byte) error {
	_, err := fmt.Printf("tata\n")
	os.Exit(1)
	return err
}

// A Template is a part of Dashboard
type Template struct {
	List []interface{} `json:"list"`
}

// A GTime contains the Dadhboard informations on the time frame of the data.
type GTime struct {
	From string `json:"from"`
	Now  bool   `json:"now"`
	To   string `json:"to"`
}

// A Annotation contains the current annotations of a dashboard
type Annotation struct {
	Enable bool          `json:"enable"`
	List   []interface{} `json:"list"`
}

// A Row is a dashboard Row it can contains multiple panels
type Row struct {
	Collapse bool    `json:"collapse"`
	Editable bool    `json:"editable"`
	Height   string  `json:"height"`
	Panels   []Panel `json:"panels" toml:"panel"`
	Title    string  `json:"title"`
}

func (row *Row) UnmarshalTOML(data []byte) error {
	fmt.Printf("toto\n")
	os.Exit(1)

	r := NewRow()
	err := toml.Unmarshal(data, &r)
	if err != nil {
		return err
	}
	*row = r
	return err
}

func NewRow() Row {
	fmt.Printf("toto\n")
	return Row{Height: "200px", Editable: true}
}

// A Panel is a component of a Row. It can be a chart, a text or a single stat panel
type Panel struct {
	Content    string   `json:"content"`
	Editable   bool     `json:"editable"`
	Error      bool     `json:"error"`
	ID         int      `json:"id"`
	Mode       string   `json:"mode"`
	Span       int      `json:"span"`
	Style      struct{} `json:"style"`
	Title      string   `json:"title"`
	Type       string   `json:"type"`
	DataSource string   `json:"datasource"`
	Fill       int      `json:"fill"`
	Targets    []Target `json:"targets" toml:"target"`
}

func NewPanel() Panel {
	return Panel{Span: 6,
		Type:       "graph",
		Editable:   true,
		DataSource: "nmon2influxdb",
		Fill:       0}
}

// A Target specify the metrics used by the Panel
type Target struct {
	Alias    string   `json:"alias"`
	Column   string   `json:"column"`
	Function string   `json:"function"`
	Hide     bool     `json:"hide"`
	Query    string   `json:"query"`
	RawQuery bool     `json:"rawQuery"`
	Series   string   `json:"series"`
	Fields   []string `json:"-"`
}

func (target *Target) UnmarshalTOML(b []byte) (err error) {
	t := NewTarget()
	err = toml.Unmarshal(b, &t)
	if err != nil {
		return
	}
	*target = t
	return
}

func NewTarget() Target {
	return Target{Function: "mean", RawQuery: false}
}

// NewSession creates a new http connection .
// It includes a cookie jar used to keep session cookies.
// The URL url specifies the host and request URI.
//
// It returns a Session struct pointer.
func NewSession(user string, password string, url string) *Session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	return &Session{client: &http.Client{Jar: jar}, User: user, Password: password, url: url}
}

// httpRequest handle the request to Grafana server.
//It returns the response body and a error if something went wrong
func (s *Session) httpRequest(method string, url string, body io.Reader) (result io.Reader, err error) {
	request, err := http.NewRequest(method, url, body)
	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return result, GrafanaError{0, "Unable to perform the http request"}
	}

	//	defer response.Body.Close()
	if response.StatusCode != 200 {
		dec := json.NewDecoder(response.Body)
		var gMess GrafanaMessage
		dec.Decode(&gMess)

		return result, GrafanaError{response.StatusCode, gMess.Message}
	}
	result = response.Body
	return
}

// DoLogon uses  a new http connection using the credentials stored in the Session struct.
// It returns a error if it cannot perform the login.
func (s *Session) DoLogon() (err error) {
	reqURL := s.url + "/login"

	login := Login{User: s.User, Password: s.Password}
	jsonStr, _ := json.Marshal(login)

	_, err = s.httpRequest("POST", reqURL, bytes.NewBuffer(jsonStr))

	return
}

// CreateDataSource creates a Grafana DataSource.
// It take a DataSource struct in parameter.
// It returns a error if it cannot perform the creation.
func (s *Session) CreateDataSource(ds DataSource) (err error) {
	reqURL := s.url + "/api/datasources"

	jsonStr, _ := json.Marshal(ds)
	_, err = s.httpRequest("PUT", reqURL, bytes.NewBuffer(jsonStr))

	return
}

// DeleteDataSource deletes a Grafana DataSource.
// It take a existing DataSource struct in parameter.
// It returns a error if it cannot perform the deletion.
func (s *Session) DeleteDataSource(ds DataSource) (err error) {

	reqURL := fmt.Sprintf("%s/api/datasources/%d", s.url, ds.ID)

	jsonStr, _ := json.Marshal(ds)
	_, err = s.httpRequest("DELETE", reqURL, bytes.NewBuffer(jsonStr))

	return
}

// GetDataSourceList return a listof existing Grafana DataSources.
// It return a array of DataSource struct.
// It returns a error if it cannot get the DataSource list.
func (s *Session) GetDataSourceList() (ds []DataSource, err error) {
	reqURL := s.url + "/api/datasources"

	body, err := s.httpRequest("GET", reqURL, nil)
	if err != nil {
		return
	}
	dec := json.NewDecoder(body)
	err = dec.Decode(&ds)
	return
}

// GetDataSource get a existing DataSource by name.
// It return a DataSource struct.
// It returns a error if a problem occurs when trying to retrieve the DataSource.
func (s *Session) GetDataSource(name string) (ds DataSource, err error) {
	dslist, err := s.GetDataSourceList()
	if err != nil {
		return
	}

	for _, elem := range dslist {
		if elem.Name == name {
			ds = elem
		}
	}
	return
}

// GetDashboard get a existing Dashboard by name.
// It takes a name string in parameter.
// It return a bytes.Buffer pointer.
// It returns a error if a problem occurs when trying to retrieve the DataSource.
func (s *Session) GetDashboard(name string) (dashboard DashboardResult, err error) {
	reqURL := s.url + "/api/dashboards/db/" + name
	body, err := s.httpRequest("GET", reqURL, nil)
	if err != nil {
		return
	}
	dec := json.NewDecoder(body)
	err = dec.Decode(&dashboard)
	return
}

// UploadDashboardString upload a new Dashboard.
// It takes a string cotnaining the json structure in parameter.
// This string will be decoded against a Dashboard struct for validation.
// If valid, the dashboard structure will be sent to UploadDashboard.
// overwrite parameter define if it overwrite existing dashboard.
// It returns a error if a problem occurs when trying to create the dashboard.
func (s *Session) UploadDashboardString(dashboard string, overwrite bool) (err error) {
	dec := json.NewDecoder(bytes.NewBuffer([]byte(dashboard)))
	var ds Dashboard
	err = dec.Decode(&ds)
	if err != nil {
		return GrafanaError{0, "dashboard template in wrong format"}
	}
	err = s.UploadDashboard(ds, overwrite)
	return
}

// UploadDashboard upload a new Dashboard.
// It takes a dashboard structure in parameter.
// It encapsulate it in a DashboardUploader structure.
// overwrite parameter define if it overwrite existing dashboard.
// It returns a error if a problem occurs when creating the dashboard.
func (s *Session) UploadDashboard(dashboard Dashboard, overwrite bool) (err error) {
	reqURL := s.url + "/api/dashboards/db"

	var content DashboardUploader
	content.Dashboard = dashboard
	content.Overwrite = overwrite
	jsonStr, _ := json.Marshal(content)

	_, err = s.httpRequest("POST", reqURL, bytes.NewBuffer(jsonStr))
	return
}

//DeleteDashboard delete a Grafana Dashboard.
// First, it try to retrieve it. And if successful, delete it using the slug attribute
// It returns a error if a problem occurs when deleting the dashboard.
func (s *Session) DeleteDashboard(name string) (err error) {
	dashRes, err := s.GetDashboard(name)
	if err != nil {
		return
	}

	slug := dashRes.Meta.Slug
	reqURL := fmt.Sprintf("%s/api/dashboards/db/%s", s.url, slug)
	_, err = s.httpRequest("DELETE", reqURL, nil)
	return
}

func ConvertTemplate(file string) (dashboard Dashboard, err error) {
	f, err := os.Open(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	if err := toml.Unmarshal(buf, &dashboard); err != nil {
		panic(err)
	}

	fmt.Println(dashboard)
	// for _, tmplRow := range template.Rows {
	// 	var row Row

	// 	for _, tmplPanel := range tmplRow.Panels {
	// 		var panel Panel

	// 		for _, field := range tmplPanel.Fields {
	// 			var target Target
	// 			target.Alias = field
	// 			target.Column = field
	// 			target.Series = tmplPanel.Serie
	// 			target.Query = fmt.Sprintf("select mean(\"%s\") from \"%s\" where $timeFilter group by time($interval) order asc", field, tmplPanel.Serie)
	// 			panel.Targets = append(panel.Targets, target)
	// 		}
	// 		row.Panels = append(row.Panels, panel)
	// 	}
	// 	dashboard.Rows = append(dashboard.Rows, row)

	// }

	// var gtime GTime

	// gtime.To = "now"
	// gtime.From = "now - 1d"

	// dashboard.GTime = gtime
	return

}
