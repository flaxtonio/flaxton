package fxBackend

import (
	"lib"
	"gopkg.in/mgo.v2"
	"time"
	"gopkg.in/mgo.v2/bson"
)

// This file for adding functionality as a server task Stack
// It will allow to distribute any task to specific group of servers or to current server

type TaskStack struct {
	DaemonID string 	`bson:"daemon_id" json:"daemon_id"`
	Tasks []lib.Task 	`bson:"tasks,omitempty" json:"tasks"`
}

var (
	DbSession *mgo.Session
	DbObject *mgo.Database
	TasksCollection *mgo.Collection
)

func NewTaskStack() (ts TaskStack) {
	var err error
	DbSession, err = mgo.Dial("127.0.0.1/flaxton")
	if err != nil {
		panic(err)
		return
	}
	DbSession.SetMode(mgo.Monotonic, true)
	DbObject = DbSession.DB("flaxton")
	TasksCollection = DbObject.C("daemon_tasks")
	go func() {
		// Just keeping connection with database by making ping request every 10 second
		var ping_error error
		for {
			ping_error = DbSession.Ping()
			if ping_error != nil {
				// If there is an error in pinging to database then we need to reconnect
				DbSession, err = mgo.Dial("127.0.0.1")
				if err != nil {
					panic(err)
					return
				}
			}
			time.Sleep(time.Second * 10)
		}
	}()
	return
}

func (stack *TaskStack) Add(daemon_id string, t lib.Task) {
	TasksCollection.Find(bson.M{"daemon_id": daemon_id}).One()
}

func (stack *TaskStack) Remove(t lib.Task) {

}