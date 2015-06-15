package main

import (
	"fmt"
	"time"
	"github.com/siddontang/ledisdb/config"
	"github.com/siddontang/ledisdb/ledis"
	"log"
	"encoding/json"
)

type TestType struct {
	FieldA string `json:"fieldA"`
	FieldB int `json:"fieldB"`
}

func main() {
	conf := config.NewConfigDefault()
	l, _ := ledis.Open(conf)
	db, _ := l.Select(0)
	k := []byte("test_key")
	d, _ := StructToByte(TestType{FieldA: "This is a test LedisDB example for Lists", FieldB: 100})
	start := time.Now()

	db.LPush(k, d)

	elapsed := time.Since(start)
	log.Printf("Done in %s", elapsed)

	start = time.Now()

	dx, _ := db.LPop(k)
	rd := TestType{}
	ByteToStruct(dx, &rd)
	elapsed = time.Since(start)
	log.Printf("Done in %s", elapsed)

	fmt.Println(rd.FieldB, rd.FieldA)
}

// Encoding Strcutures to byte array

func StructToByte(data interface{}) (ret_data []byte, err error) {
	ret_data, err = json.Marshal(data)
	return
}

func ByteToStruct(b []byte, ret_data interface{}) (err error) {
	err = json.Unmarshal(b, &ret_data)
	return
}