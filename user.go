package main

import (
	"bytes"
	"errors"
	"io/ioutil"
)

var users = map[string]string{}

func init() {
	content, err := ioutil.ReadFile("/etc/passwd")
	if err != nil {
		return
	}
	rows := bytes.Split(content, []byte("\n"))
	for _, row := range rows {
		cols := bytes.Split(row, []byte(":"))
		if len(cols) >= 4 {
			users[string(cols[0])] = string(cols[2])
		}
	}
}

func LookupUser(uname string) (string, error) {
	uid, ok := users[uname]
	if !ok {
		if len(users) == 0 {
			return "", errors.New("Initializing user information failed")
		}
		return "", errors.New("User not found")
	}
	return uid, nil
}
