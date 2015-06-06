package grafanaclient_test

import (
	"github.com/adejoux/grafanaclient"
)

func ExampleSession_CreateDataSource() {
	myurl := "http://localhost:3000"

	myds := grafanaclient.DataSource{Name: "testgf",
		Type:     "influxdb_08",
		Access:   "direct",
		URL:      "http://localhost:8086",
		User:     "root",
		Password: "root",
		Database: "test",
	}

	session := grafanaclient.NewSession("admin", "admin", myurl)
	session.DoLogon()
	session.CreateDataSource(myds)
	// Output:
	//

}

func ExampleSession_DeleteDataSource() {
	myurl := "http://localhost:3000"

	session := grafanaclient.NewSession("admin", "admin", myurl)
	session.DoLogon()
	ds, _ := session.GetDataSource("testgf")
	session.DeleteDataSource(ds)
	// Output:
	//

}
