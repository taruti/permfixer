package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
)

var groups = readUserMap("/etc/group")

type userMapFile struct {
	m map[string]int
}

func readUserMap(filename string) userMapFile {
	um := userMapFile{m: map[string]int{}}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return um
	}
	rows := bytes.Split(content, []byte("\n"))
	for _, row := range rows {
		cols := bytes.Split(row, []byte(":"))
		if len(cols) >= 4 {
			if v, err := strconv.Atoi(string(cols[2])); err == nil {
				um.m[string(cols[0])] = v
			}
		}
	}
	return um
}

func (u *userMapFile) Lookup(name string) (int, error) {
	if name == "" {
		return -1, nil
	}
	id, ok := u.m[name]
	if !ok {
		if len(u.m) == 0 {
			return 0, errors.New("Initializing user/group information failed")
		}
		return 0, errors.New("User/Group not found")
	}
	return id, nil
}
