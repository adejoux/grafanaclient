# grafanaclient
--
    import "github.com/adejoux/grafanaclient"

Package grafanaclient provide a simple API to manage Grafana 2.0 DataSources and
Dashboards in Go. It's using Grafana 2.0 REST API.

## Usage

#### type Annotation

```go
type Annotation struct {
	Enable bool          `json:"enable"`
	List   []interface{} `json:"list"`
}
```

A Annotation contains the current annotations of a dashboard

#### type Dashboard

```go
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
	Rows            []Row         `json:"rows"`
	Title           string        `json:"title"`
	Version         int           `json:"version"`
	Timezone        string        `json:"timezone"`
}
```

A Dashboard contains the Dashboard structure.

#### type DashboardResult

```go
type DashboardResult struct {
	Meta  Meta      `json:"meta"`
	Model Dashboard `json:"model"`
}
```

A DashboardResult contains the response from Grafana when requesting a
Dashboard. It contains the Dashboard itself and the meta data.

#### type DashboardUploader

```go
type DashboardUploader struct {
	Dashboard Dashboard `json:"dashboard"`
	Overwrite bool      `json:"overwrite"`
}
```

A DashboardUploader encapsulates a complete Dashboard

#### type DataSource

```go
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
```

A DataSource contains the json structure of Grafana DataSource

#### type GTime

```go
type GTime struct {
	From string `json:"from"`
	Now  bool   `json:"now"`
	To   string `json:"to"`
}
```

A GTime contains the Dadhboard informations on the time frame of the data.

#### type GrafanaError

```go
type GrafanaError struct {
	Code        int
	Description string
}
```

GrafanaError is a error structure to handle error messages in this library

#### func (GrafanaError) Error

```go
func (h GrafanaError) Error() string
```
Error generate a text error message. If Code is zero, we know it's not a http
error.

#### type GrafanaMessage

```go
type GrafanaMessage struct {
	Message string `json:"message"`
}
```

A GrafanaMessage contains the json error message received when http request
failed

#### type Login

```go
type Login struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
```

A Login contains the json structure of Grafana authentication request

#### type Meta

```go
type Meta struct {
	Created    string `json:"created"`
	Expires    string `json:"expires"`
	IsHome     bool   `json:"isHome"`
	IsSnapshot bool   `json:"isSnapshot"`
	IsStarred  bool   `json:"isStarred"`
	Slug       string `json:"slug"`
}
```

A Meta contains a Dashboard metadata.

#### type Panel

```go
type Panel struct {
	Content  string   `json:"content"`
	Editable bool     `json:"editable"`
	Error    bool     `json:"error"`
	ID       int      `json:"id"`
	Mode     string   `json:"mode"`
	Span     int      `json:"span"`
	Style    struct{} `json:"style"`
	Title    string   `json:"title"`
	Type     string   `json:"type"`
	Targets  []Target `json:"targets"`
}
```

A Panel is a component of a Row. It can be a chart, a text or a single stat
panel

#### type Row

```go
type Row struct {
	Collapse bool    `json:"collapse"`
	Editable bool    `json:"editable"`
	Height   string  `json:"height"`
	Panels   []Panel `json:"panels"`
	Title    string  `json:"title"`
}
```

A Row is a dashboard Row it can contians multiple panels

#### type Session

```go
type Session struct {
	User     string
	Password string
}
```

Session contains user credentials, url and a pointer to http client session.

#### func  NewSession

```go
func NewSession(user string, password string, url string) *Session
```
NewSession creates a new http connection . It includes a cookie jar used to keep
session cookies. The URL url specifies the host and request URI.

It returns a Session struct pointer.

#### func (*Session) CreateDataSource

```go
func (s *Session) CreateDataSource(ds DataSource) (err error)
```
CreateDataSource creates a Grafana DataSource. It take a DataSource struct in
parameter. It returns a error if it cannot perform the creation.

#### func (*Session) DeleteDashboard

```go
func (s *Session) DeleteDashboard(name string) (err error)
```
DeleteDashboard delete a Grafana Dashboard. First, it try to retrieve it. And if
successful, delete it using the slug attribute It returns a error if a problem
occurs when deleting the dashboard.

#### func (*Session) DeleteDataSource

```go
func (s *Session) DeleteDataSource(ds DataSource) (err error)
```
DeleteDataSource deletes a Grafana DataSource. It take a existing DataSource
struct in parameter. It returns a error if it cannot perform the deletion.

#### func (*Session) DoLogon

```go
func (s *Session) DoLogon() (err error)
```
DoLogon uses a new http connection using the credentials stored in the Session
struct. It returns a error if it cannot perform the login.

#### func (*Session) GetDashboard

```go
func (s *Session) GetDashboard(name string) (dashboard DashboardResult, err error)
```
GetDashboard get a existing Dashboard by name. It takes a name string in
parameter. It return a bytes.Buffer pointer. It returns a error if a problem
occurs when trying to retrieve the DataSource.

#### func (*Session) GetDataSource

```go
func (s *Session) GetDataSource(name string) (ds DataSource, err error)
```
GetDataSource get a existing DataSource by name. It return a DataSource struct.
It returns a error if a problem occurs when trying to retrieve the DataSource.

#### func (*Session) GetDataSourceList

```go
func (s *Session) GetDataSourceList() (ds []DataSource, err error)
```
GetDataSourceList return a listof existing Grafana DataSources. It return a
array of DataSource struct. It returns a error if it cannot get the DataSource
list.

#### func (*Session) UploadDashboard

```go
func (s *Session) UploadDashboard(dashboard Dashboard, overwrite bool) (err error)
```
UploadDashboard upload a new Dashboard. It takes a dashboard structure in
parameter. It encapsulate it in a DashboardUploader structure. overwrite
parameter define if it overwrite existing dashboard. It returns a error if a
problem occurs when creating the dashboard.

#### func (*Session) UploadDashboardString

```go
func (s *Session) UploadDashboardString(dashboard string, overwrite bool) (err error)
```
UploadDashboardString upload a new Dashboard. It takes a string cotnaining the
json structure in parameter. This string will be decoded against a Dashboard
struct for validation. If valid, the dashboard structure will be sent to
UploadDashboard. overwrite parameter define if it overwrite existing dashboard.
It returns a error if a problem occurs when trying to create the dashboard.

#### type Target

```go
type Target struct {
	Alias    string `json:"alias"`
	Column   string `json:"column"`
	Function string `json:"function"`
	Hide     bool   `json:"hide"`
	Query    string `json:"query"`
	RawQuery bool   `json:"rawQuery"`
	Series   string `json:"series"`
}
```

A Target specify the metrics used by the Panel

#### type Template

```go
type Template struct {
	List []interface{} `json:"list"`
}
```

A Template is a part of Dashboard

## License

Copyright © 2015 Alain Dejoux <adejoux@djouxtech.net>.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

