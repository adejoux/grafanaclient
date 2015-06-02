package grafanaclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
)

//
// HTTP session struct
//

type Session struct {
	client   *http.Client
	User     string
	Password string
	url      string
}

type Login struct {
	User     string `json:"user"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type DataSource struct {
	Id                int
	OrgId             int    `json:"orgId"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	Access            string `json:"access"`
	Url               string `json:"url"`
	Password          string `json:"password"`
	User              string `json:"user"`
	Database          string `json:"database"`
	BasicAuth         bool   `json:"basicAuth"`
	BasicAuthUser     string `json:"basicAuthUser"`
	BasicAuthPassword string `json:"basicAuthPassword"`
	IsDefault         bool   `json:"isDefault"`
}

func NewSession(user string, password string, url string) *Session {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	return &Session{client: &http.Client{Jar: jar}, User: user, Password: password, url: url}
}

func (s *Session) DoLogon() (err error) {
	reqUrl := s.url + "/login"

	login := Login{User: s.User, Password: s.Password}
	jsonStr, _ := json.Marshal(login)

	request, err := http.NewRequest("POST", reqUrl, bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return
	} else {
		defer response.Body.Close()
		if response.StatusCode != 200 {
			error_message := fmt.Sprintf("%d", response.StatusCode)
			return errors.New(error_message)
		}
	}
	return
}

func (s *Session) CreateDataSource(ds DataSource) (err error) {
	reqUrl := s.url + "/api/datasources"

	jsonStr, _ := json.Marshal(ds)
	fmt.Println(string(jsonStr))

	request, err := http.NewRequest("PUT", reqUrl, bytes.NewBuffer(jsonStr))
	request.Header.Set("Content-Type", "application/json")

	response, err := s.client.Do(request)
	if err != nil {
		return
	} else {
		defer response.Body.Close()
		if response.StatusCode != 200 {
			error_message := fmt.Sprintf("%d", response.StatusCode)
			return errors.New(error_message)
		}
	}
	return
}

// func (s *Session) DeleteDataSource(ds DataSource) (err error) {
// 	reqUrl := s.url + "/api/datasources"

// 	jsonStr, _ := json.Marshal(ds)
// 	fmt.Println(string(jsonStr))

// 	request, err := http.NewRequest("DELETE", reqUrl, bytes.NewBuffer(jsonStr))
// 	request.Header.Set("Content-Type", "application/json")

// 	response, err := s.client.Do(request)
// 	if err != nil {
// 		return
// 	} else {
// 		defer response.Body.Close()
// 		if response.StatusCode != 200 {
// 			error_message := fmt.Sprintf("%d", response.StatusCode)
// 			return errors.New(error_message)
// 		}
// 	}
// 	return
// }

func (s *Session) GetDataSourceList() (ds []DataSource, err error) {
	reqUrl := s.url + "/api/datasources"

	response, err := s.client.Get(reqUrl)
	if err != nil {
		return
	} else {
		defer response.Body.Close()
		if response.StatusCode != 200 {
			error_message := fmt.Sprintf("%d", response.StatusCode)
			return ds, errors.New(error_message)
		}
	}

	dec := json.NewDecoder(response.Body)
	dec.Decode(&ds)
	return
}

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

func (s *Session) GetDashboard(name string) (dashboard *bytes.Buffer) {
	dashUrl := s.url + "/api/dashboards/db/" + name
	response, err := s.client.Get(dashUrl)
	if err != nil {
		return
	} else {
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return
		}
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	return buf
}

func (s *Session) UploadDashboard(dashboard string) {
	dashUrl := s.url + "/api/dashboards/db"

	content := `{ "dashboard": ` + dashboard + `,"overwrite": true }`

	request, _ := http.NewRequest("POST", dashUrl, bytes.NewBuffer([]byte(content)))
	request.Header.Set("Content-Type", "application/json")

	s.client.Do(request)

}

func (s *Session) DeleteDashboard(name string) {

}
