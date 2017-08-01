package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strconv"
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

func LookupGroup(gname *string) (string, int, error) {
	if gname == nil || *gname == "" {
		return "", -1, nil
	}
	gids, ok := groups[*gname]
	if !ok {
		if len(groups) == 0 {
			return "", 0, errors.New("Initializing group information failed")
		}
		return "", 0, errors.New("Group not found")
	}
	gid, err := strconv.ParseInt(gids, 10, 32)
	return gids, int(gid), err
}
