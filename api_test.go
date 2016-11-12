package grafanaclient

import "github.com/stretchr/testify/assert"
import "testing"

var url = "http://192.168.56.101:3000"
var user = "admin"
var pass = "admin"

var ds = DataSource{Name: "testme",
	Type:      "influxdb",
	Access:    "proxy",
	URL:       "http://localhost:8086",
	User:      "root",
	Password:  "root",
	Database:  "test",
	IsDefault: true}

var dashboard = `{
        "id": null,
        "title": "new dashboard",
        "tags": [ "templated" ],
        "timezone": "browser",
        "rows": [
          {
          }
        ],
        "schemaVersion": 6,
        "version": 0
      }`

func Test_DoLogon(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one")
}

// func Test_GetPlugins(t *testing.T) {
// 	session := NewSession(pass, pass, url)
// 	err := session.DoLogon()
// 	assert.Nil(t, err, "We are expecting no error and got one")
// 	plugins, err := session.GetPlugins("datasource")
// 	assert.Nil(t, err, "We are expecting no error and got one when getting DataSource Plugins")
// 	assert.NotNil(t, plugins["influxdb"], "We didn't find a plugin for InfluxDB in DataSource")
// }

func Test_CreateDataSource(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one")
	err = session.CreateDataSource(ds)
	assert.Nil(t, err, "We are expecting no error and got one when creating DataSource")
}

func Test_GetDataSourceList(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")
	dslist, err := session.GetDataSourceList()
	assert.Nil(t, err, "We are expecting no error and got one getting DataSource")
	var check bool
	for _, ds := range dslist {
		if ds.Name == "testme" {
			check = true
		}
	}

	assert.Equal(t, true, check, "We didn't find the test datasource")
}

func Test_GetDataSource(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	resDs, _ := session.GetDataSource("testme")

	assert.Equal(t, "testme", resDs.Name, "We are expecting to retrieve testme DataSource and didn't get it")
}

func Test_CreateDashboard(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	err = session.UploadDashboardString(dashboard, true)
	assert.Nil(t, err, "We are expecting no error and got one when Uploading")
}

func Test_GetDashboard(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	dashboard, err := session.GetDashboard("new-dashboard")
	assert.Nil(t, err, "We are expecting no error and got one when getting dashboard")
	assert.NotNil(t, dashboard, "We are expecting to receive a dashboard")

}

func Test_ConvertTemplate(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	dashboard, err := ConvertTemplate("example.toml")
	assert.Nil(t, err, "We are expecting no error and got one when Converting template")
	assert.NotNil(t, dashboard, "We are expecting to receive a dashboard")
}

func Test_UploadDasboardFromTemplate(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	dashboard, err := ConvertTemplate("example.toml")
	assert.Nil(t, err, "We are expecting no error and got one when Converting template")
	assert.NotNil(t, dashboard, "We are expecting to receive a dashboard")
	err = session.UploadDashboard(dashboard, true)
	assert.Nil(t, err, "We are expecting no error and got one when Uploading")
}

func Test_DeleteDataSource(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	resDs, err := session.GetDataSource("testme")

	err = session.DeleteDataSource(resDs)
	assert.Nil(t, err, "We are expecting no error and got one when Deleting")
}

func Test_DeleteDashboard(t *testing.T) {
	session := NewSession(user, pass, url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one when Login")

	err = session.DeleteDashboard("new-dashboard")
	assert.Nil(t, err, "We are expecting no error and got one when Deleting")
}
