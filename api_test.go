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

func Test_GetDataSourceList(t *testing.T) {
	session := NewSession("admin", "admin", url)
	err := session.DoLogon()
	assert.Nil(t, err, "We are expecting no error and got one")
	err = session.GetDataSourceList()
	assert.Nil(t, err, "We are expecting no error and got one")
}
