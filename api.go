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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/imdario/mergo"
	"github.com/naoina/toml"
)

const timeout = 5

var protocolRegexp = regexp.MustCompile(`^https://`)

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

// A DataSourcePlugin contains the json structure of Grafana DataSource plugin
type DataSourcePlugin struct {
	Annotations struct {
		Enable bool          `json:"enable"`
		List   []interface{} `json:"list"`
	} `json:"annotations"`
	Module      string `json:"module"`
	Name        string `json:"name"`
	Partials    PluginPartial
	PluginType  string `json:"pluginType"`
	ServiceName string `json:"serviceName"`
	Type        string `json:"type"`
}

//Plugin is a Grafana 3.0 structure for plugins
type Plugin struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Pinned  bool   `json:"pinned"`
	Info    struct {
		Author struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"author"`
		Description string      `json:"description"`
		Links       interface{} `json:"links"`
		Logos       struct {
			Small string `json:"small"`
			Large string `json:"large"`
		} `json:"logos"`
		Screenshots interface{} `json:"screenshots"`
		Version     string      `json:"version"`
		Updated     string      `json:"updated"`
	} `json:"info"`
	LatestVersion string `json:"latestVersion"`
	HasUpdate     bool   `json:"hasUpdate"`
}

// Plugins is an array of Plugin
type Plugins []Plugin

// A DataSourcePlugins contains a map of DataSourcePlugin
type DataSourcePlugins map[string]DataSourcePlugin

// A PluginPartial contains the json structure of Grafana DataSource Plugin Partial
type PluginPartial struct {
	Annotations string `json:"annotations"`
	Config      string `json:"config"`
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
	Refresh         string        `json:"refresh"`
	Annotations     Annotation    `json:"annotations"`
	SchemaVersion   int           `json:"schemaVersion"`
	SharedCrosshair bool          `json:"sharedCrosshair"`
	Style           string        `json:"style"`
	Templating      Templating    `json:"templating,omitempty" toml:"templates"`
	Tags            []interface{} `json:"tags"`
	GTime           GTime         `json:"time" toml:"time"`
	Rows            []Row         `json:"rows" toml:"row"`
	Title           string        `json:"title"`
	Version         int           `json:"version"`
	Timezone        string        `json:"timezone"`
}

// A GTime contains the Dadhboard informations on the time frame of the data.
type GTime struct {
	From string `json:"from"`
	Now  bool   `json:"now"`
	To   string `json:"to"`
}

// A Templating contains a List of Templates usable in Dashboard
type Templating struct {
	List Templates `json:"list" toml:"template"`
}

//Template define a variable usable in Grafana
type Template struct {
	AllFormat string `json:"allFormat"`
	Current   struct {
		Tags  []interface{} `json:"tags"`
		Text  string        `json:"text"`
		Value interface{}   `json:"value"`
	} `json:"current,omitempty"`
	Datasource  string `json:"datasource"`
	IncludeAll  bool   `json:"includeAll"`
	Multi       bool   `json:"multi"`
	MultiFormat string `json:"multiFormat"`
	Name        string `json:"name"`
	Options     []struct {
		Selected bool   `json:"selected"`
		Text     string `json:"text"`
		Value    string `json:"value"`
	} `json:"options,omitempty"`
	Query         string `json:"query"`
	Refresh       string `json:"refresh"`
	RefreshOnLoad bool   `json:"refresh_on_load"`
	Regex         string `json:"regex"`
	Type          string `json:"type"`
}

//Templates is an Array of Template
type Templates []Template

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

// A Panel is a component of a Row. It can be a chart, a text or a single stat panel
type Panel struct {
	Content         string           `json:"content"`
	Editable        bool             `json:"editable"`
	Error           bool             `json:"error"`
	ID              int              `json:"id"`
	Mode            string           `json:"mode"`
	Span            int              `json:"span"`
	Style           struct{}         `json:"style"`
	Title           string           `json:"title"`
	Type            string           `json:"type"`
	Fill            int              `json:"fill"`
	Stack           bool             `json:"stack"`
	Targets         []Target         `json:"targets" toml:"target"`
	Metrics         []Metric         `json:"-" toml:"metric"`
	SeriesOverrides []SeriesOverride `json:"seriesOverrides,omitempty" toml:"override"`
	Tooltip         Tooltip          `json:"tooltip,omitempty"`
	PageSize        int              `json:"pageSize,omitempty" toml:"pageSize,omitempty"`
	Legend          Legend           `json:"legend,omitempty"`
	LeftYAxisLabel  string           `json:"leftYAxisLabel,omitempty"`
	RightYAxisLabel string           `json:"rightYAxisLabel,omitempty"`
	DataSource      string           `json:"datasource,omitempty"`
	NullPointMode   string           `json:"nullPointMode,omitempty"`
	ValueName       string           `json:"valueName,omitempty"`
	Lines           bool             `json:"lines,omitempty"`
	Linewidth       int              `json:"linewidth,omitempty"`
	Points          bool             `json:"points,omitempty"`
	Pointradius     int              `json:"pointradius,omitempty"`
	Bars            bool             `json:"bars,omitempty"`
	Percentage      bool             `json:"percentage,omitempty"`
	SteppedLine     bool             `json:"steppedLine,omitempty"`
	TimeFrom        interface{}      `json:"timeFrom,omitempty"`
	TimeShift       interface{}      `json:"timeShift,omitempty"`
}

// A Target specify the metrics used by the Panel
type Target struct {
	Alias       string    `json:"alias"`
	Hide        bool      `json:"hide"`
	Measurement string    `json:"measurement"`
	GroupBy     []GroupBy `json:"groupBy"`
	Select      []Selects `json:"select,omitempty"`
	Tags        []Tag     `json:"tags"`
	DsType      string    `json:"dsType,omitempty"`
	Transform   string    `json:"transform,omitempty" toml:"transform,omitempty"`
}

// Selects array of Select struct
type Selects []Select

// A Select specify the criteria to perform selection
type Select struct {
	Type   string   `json:"type"`
	Params []string `json:"params"`
}

// A Legend specify the legend options used by the Panel
type Legend struct {
	Show         bool `json:"show"`
	Values       bool `json:"values"`
	Min          bool `json:"min"`
	Max          bool `json:"max"`
	Current      bool `json:"current"`
	Total        bool `json:"total"`
	Avg          bool `json:"avg"`
	AlignAsTable bool `json:"alignAsTable"`
}

// A GroupBy struct is used to setup the group by part of the query
type GroupBy struct {
	Type     string   `json:"type"`
	Interval string   `json:"interval,omitempty"`
	Params   []string `json:"params"`
}

//NewGroupBy initialize a GroupBy structure
func NewGroupBy() []GroupBy {
	return []GroupBy{GroupBy{Type: "time", Interval: "auto"}}
}

// TagKeys returns a array of keys
func (target *Target) TagKeys() []string {

	keys := make([]string, len(target.Tags))

	for i, tag := range target.Tags {
		keys[i] = tag.Key
		i++
	}

	return keys
}

// A Metric is only used in TOML templates to define the targets to create
type Metric struct {
	Measurement string
	Fields      []string
	Hosts       []string
	Alias       []string
}

// A SeriesOverride allows to setup specific override by serie
type SeriesOverride struct {
	Alias     string `json:"alias"`
	Stack     bool   `json:"stack"`
	Fill      int    `json:"fill"`
	Transform string `json:"transform"`
}

// A Tag allows to filter the values
type Tag struct {
	Condition string `json:"condition"`
	Key       string `json:"key"`
	Value     string `json:"value"`
}

// A Tooltip allow to setup some graphic display options
type Tooltip struct {
	ValueType string `json:"value_type" toml:"value_type"`
}

// NewRow create a new Grafana row with default values
func NewRow() Row {
	return Row{Height: "200px", Editable: true}
}

// NewPanel create a new Grafana panel with default values
func NewPanel() Panel {
	return Panel{Span: 6,
		Type:          "graph",
		Editable:      true,
		Fill:          0,
		Legend:        NewLegend(),
		NullPointMode: "connected",
	}
}

// NewTarget create a new Grafana target with default values
func NewTarget() Target {
	return Target{Alias: "$tag_host $tag_name", DsType: "influxdb"}
}

// NewLegend create a new Grafana legend with default values
func NewLegend() Legend {
	return Legend{Show: true}
}

// NewSeriesOverride create a new Grafana series override using the specified alias
func NewSeriesOverride(alias string) SeriesOverride {
	return SeriesOverride{Alias: alias}
}

// NewGTime create a default time window for Grafana
func NewGTime() GTime {
	return GTime{From: "now-24h", To: "now"}
}

// NewTemplate create a default template for Grafana
func NewTemplate() Template {
	return Template{Type: "query", Refresh: "1", AllFormat: "regex values", MultiFormat: "regex values"}
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

	client := http.Client{Jar: jar, Timeout: time.Second * timeout}

	if protocolRegexp.MatchString(url) {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.Transport = tr
	}

	return &Session{client: &client, User: user, Password: password, url: url}
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
	_, err = s.httpRequest("POST", reqURL, bytes.NewBuffer(jsonStr))

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

// GetDataSourcePlugins return a list of existing Grafana DataSources.
// It return a array of DataSource struct.
// It returns a error if it cannot get the DataSource list.
func (s *Session) GetDataSourcePlugins() (plugins DataSourcePlugins, err error) {
	reqURL := s.url + "/api/datasources/plugins"

	body, err := s.httpRequest("GET", reqURL, nil)

	if err != nil {
		return
	}

	dec := json.NewDecoder(body)
	err = dec.Decode(&plugins)
	return
}

//GetPlugins get the list of plugins by PluginType
func (s *Session) GetPlugins(pluginType string) (plugins Plugins, err error) {
	reqURL := s.url + "/api/plugins?type=" + pluginType

	body, err := s.httpRequest("GET", reqURL, nil)

	if err != nil {
		return
	}

	dec := json.NewDecoder(body)
	err = dec.Decode(&plugins)
	return
}

// GetDataSourceList return a list of existing Grafana DataSources.
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

// AddRow add a row to an existing dashboard.
// It takes a Row struct in parameter.
func (db *Dashboard) AddRow(row Row) {
	db.Rows = append(db.Rows, row)
}

// SetTimeFrame setup the dashboard timeframe.
func (db *Dashboard) SetTimeFrame(from time.Time, to time.Time) {
	db.GTime = GTime{From: from.Format(time.RFC3339), To: to.Format(time.RFC3339)}
}

// AddPanel add a panel to an existing row.
// It takes a Row struct in parameter.
func (row *Row) AddPanel(panel Panel) {
	row.Panels = append(row.Panels, panel)
}

// AddTarget add a target to an existing panel.
// It takes a Panel struct in parameter.
func (panel *Panel) AddTarget(target Target) {
	panel.Targets = append(panel.Targets, target)
}

// FilterByTag add a Tag to an existing target.
// It takes a name and value strings in parameter.
func (target *Target) FilterByTag(name string, value string) {
	target.Tags = append(target.Tags, Tag{Key: name, Value: value})
}

// GroupByTag add a group by selection to the Target
// It takes a string in parameter specifying the tag name
func (target *Target) GroupByTag(tag string) {
	if len(target.GroupBy) == 0 {
		target.GroupBy = NewGroupBy()
	}
	target.GroupBy = append(target.GroupBy, GroupBy{Type: "tag", Params: []string{tag}})
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

//ConvertTemplate converts a string to a dashboard structure
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

	dashboard.Editable = true
	if tomlErr := toml.Unmarshal(buf, &dashboard); tomlErr != nil {
		//try to convert a json template
		if jsonErr := json.Unmarshal(buf, &dashboard); jsonErr != nil {
			fmt.Printf("not a JSON template: %s\n", err.Error())
			err = fmt.Errorf("Unable to parse template:\nTOML error: %s\nJSON error: %s\n", tomlErr.Error(), jsonErr.Error())
			return dashboard, err
		}
		//cleanup existing dashboard ID
		dashboard.ID = 0
	}
	defRow := NewRow()
	defPanel := NewPanel()

	if len(dashboard.Templating.List) > 0 {
		defTemplate := NewTemplate()
		for i := range dashboard.Templating.List {
			template := &dashboard.Templating.List[i]
			if err := mergo.Merge(template, defTemplate); err != nil {
				panic(err)
			}
		}
	}

	for i := range dashboard.Rows {
		row := &dashboard.Rows[i]
		if err := mergo.Merge(row, defRow); err != nil {
			panic(err)
		}
		for j := range row.Panels {
			panel := &row.Panels[j]
			if err := mergo.Merge(panel, defPanel); err != nil {
				panic(err)
			}
			for _, metric := range panel.Metrics {
				target := NewTarget()
				fields := strings.Join(metric.Fields, "|")
				hosts := strings.Join(metric.Hosts, "|")

				target.Measurement = metric.Measurement

				// adding tags
				hostTag := Tag{Key: "host", Value: "/" + hosts + "/"}
				target.Tags = append(target.Tags, hostTag)
				fieldsTag := Tag{Key: "name", Value: "/" + fields + "/", Condition: "AND"}
				target.Tags = append(target.Tags, fieldsTag)
				target.GroupBy = NewGroupBy()
				target.GroupBy = append(target.GroupBy, GroupBy{Type: "tag", Params: []string{"name"}})
				target.GroupBy = append(target.GroupBy, GroupBy{Type: "tag", Params: []string{"host"}})
				panel.Targets = append(panel.Targets, target)
			}
		}
	}

	if dashboard.GTime == (GTime{}) {
		dashboard.GTime = NewGTime()
	}
	return

}
