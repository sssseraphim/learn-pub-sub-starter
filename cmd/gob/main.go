package main

import (
	"bytes"
	"encoding/gob"
	"time"
)

func encode(gl GameLog) ([]byte, error) {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	encoder.Encode(gl)
	return res.Bytes(), nil
}

func decode(data []byte) (GameLog, error) {
	var res GameLog
	var input bytes.Buffer
	input.Write(data)
	decoder := gob.NewDecoder(&input)
	decoder.Decode(&res)
	return res, nil
}

// don't touch below this line

type GameLog struct {
	CurrentTime time.Time
	Message     string
	Username    string
}
