package db

import (
	"github.com/gocql/gocql"
)

type Database struct {
	db *gocql.Session
}

func NewDatabaseConn() (*Database, error) {
	cluster := gocql.NewCluster("localhost")
	cluster.Keyspace = "clique"
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: "admin",
		Password: "admin",
	}

	session, err := cluster.CreateSession()

	if err != nil {
		return nil, err
	}

	return &Database{db: session}, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetDb() *gocql.Session {
	return d.db
}
