package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
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

func LookupUser(uname *string) (string, int, error) {
	if uname == nil || *uname=="" {
		return "", -1, nil
	}
	uids, ok := users[*uname]
	if !ok {
		if len(users) == 0 {
			return "", 0, errors.New("Initializing user information failed")
		}
		return "", 0, errors.New("User not found")
	}
	uid, err := strconv.ParseInt(uids, 10, 32)
	return uids, int(uid), err
}
