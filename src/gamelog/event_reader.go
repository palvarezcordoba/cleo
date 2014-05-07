package gamelog

import "os"
import "fmt"
import "proto"
import gproto "code.google.com/p/goprotobuf/proto"

func Read(filename string) *proto.GameLog {
	// Read the file and unmarshal
	// https://code.google.com/p/goprotobuf/source/browse/README
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Cannot find file.")
		return &proto.GameLog{}
	}
	
	fi, err := file.Stat()
	buffer := make([]byte, fi.Size())
	
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Printf("Cannot read full file.")
	}
	
	event_log := &proto.GameLog{}
	gproto.Unmarshal(buffer, event_log) 
	
	// Return a GameLog
	return event_log
}
