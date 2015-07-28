package storage

import (
	"gopkg.in/mgo.v2"
)

type MongoRunner func(*mgo.Collection) error

type MongoSession struct {
	serverStr string
	database  string
	password  string
	session   *mgo.Session
}

func NewMongoSession(serverStr, database, password string) *MongoSession {
	return &MongoSession{
		serverStr: serverStr,
		database:  database,
		password:  password,
	}
}

func (this *MongoSession) Session() *mgo.Session {
	if this.session == nil {
		var err error
		this.session, err = mgo.Dial(this.serverStr)
		if err != nil {
			panic(err) // no, not really
		}
	}
	return this.session.Clone()
}

func (this *MongoSession) Do(collection string, f MongoRunner) error {
	session := this.Session()
	defer session.Close()
	c := session.DB(this.database).C(collection)
	return f(c)
}
