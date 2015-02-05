package main

import (
	"bytes"
	"errors"
	"io/ioutil"
)

var groups = map[string]string{}

func init() {
	content, err := ioutil.ReadFile("/etc/group")
	if err != nil {
		return
	}
	rows := bytes.Split(content, []byte("\n"))
	for _, row := range rows {
		cols := bytes.Split(row, []byte(":"))
		if len(cols) == 4 {
			groups[string(cols[0])] = string(cols[2])
		}
	}
}

func LookupGroup(gname string) (string, error) {
	gid, ok := groups[gname]
	if !ok {
		if len(groups) == 0 {
			return "", errors.New("Initializing group information failed")
		}
		return "", errors.New("Group not found")
	}
	return gid, nil
}
