package grafanaclient

import "github.com/stretchr/testify/assert"
import "testing"

var url = "http://localhost:3000"

// func Test_BadConnect(t *testing.T) {
// 	testDB := NewInfluxDB()
// 	err := testDB.InitSession("locallhost", "8087", "testdb", "admin", "admin")
// 	assert.NotNil(t, err, "We are expecting error and didn't got one")
// }

func Test_DoLogon(t *testing.T) {
	session := NewSession("admin", "admin", url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one")
}

func Test_CreateDataSource(t *testing.T) {
	session := NewSession("admin", "admin", url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one")
	ds := DataSource{Name: "testme",
		Type:      "influxdb_08",
		Access:    "direct",
		Url:       "http://localhost:8086",
		User:      "root",
		Password:  "root",
		Database:  "test",
		IsDefault: true}
	err = session.CreateDataSource(ds)
	assert.Nil(t, err, "We are expecting no error and got one when creating DataSource")
}

// func Test_GetDataSourceList(t *testing.T) {
// 	session := NewSession("admin", "admin", url)
// 	err := session.DoLogon()
// 	assert.Nil(t, err, "We are expecting no error and got one when Login")
// 	dslist, err := session.GetDataSourceList()
// 	assert.Nil(t, err, "We are expecting no error and got one getting DataSource")
// 	var check bool
// 	for _, ds := range dslist {
// 		if ds.Name == "test" {
// 			check = true
// 		}
// 	}

// 	fmt.Println(dslist)
// 	assert.Equal(t, true, check, "We didn't find the test datasource")
// }
