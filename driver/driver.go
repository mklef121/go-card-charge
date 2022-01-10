package driver

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

/*
To access databases in Go, you use a `sql.DB`. You use this type to
create statements and transactions, execute queries, and fetch results.

The first thing you should know is that a sql.DB isn’t a database connection.
It also doesn’t map to any particular database software’s notion of a “database” or “schema.”
It’s an abstraction of the interface and existence of a database,

sql.DB performs some important tasks for you behind the scenes:

- It opens and closes connections to the actual underlying database, via the driver.

- It manages a pool of connections as needed, which may be a variety of things as mentioned.

To use database/sql you’ll need the package itself, as well as a driver for the specific database you want to use.

import _ "github.com/go-sql-driver/mysql" //make it anonymous

You generally shouldn’t use driver packages directly, although some drivers encourage you to do so.
(In our opinion, it’s usually a bad idea.) Instead, your code should only refer to
types defined in database/sql, if possible. This helps avoid making your code dependent on the driver,
so that you can change the underlying driver (and thus the database you’re accessing) with minimal code changes

Now that you’ve loaded the driver package, you’re ready to create a database object, a sql.DB.

*/
func OpenDb(dsn string) (*sql.DB, error) {
	//We need to create the sql.DB struct, and the Open function does it for us.
	// We pass in the name registerd by the driver and the DSN to the Open function
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		fmt.Println(err)

		return nil, err
	}

	return db, nil
}
