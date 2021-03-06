package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var dirp = flag.String("dir", ".", "Directory to fix permissions in recursively")
var secp = flag.Int("sec", 60, "Time between checks in seconds")
var userp = flag.String("user", "", "User for chown")
var groupp = flag.String("group", "", "Group for chgrp")
var permfp = flag.String("permf", "", "Permissions for chmod in octal for files")
var permdp = flag.String("permd", "", "Permissions for chmod in octal for directories")
var uid, gid int
var fmode, dmode uint32

func main() {
	flag.Parse()

	t, e := strconv.ParseInt(*permfp, 8, 32)
	if e != nil {
		log.Fatal("Error parsing octal parameter permf", e)
	}
	fmode = uint32(t)
	t, e = strconv.ParseInt(*permdp, 8, 32)
	if e != nil {
		log.Fatal("Error parsing octal parameter permd", e)
	}
	dmode = uint32(t)

	_, uid, e = LookupUser(userp)
	if e != nil {
		log.Fatal("Error looking up user", *userp, e)
	}

	_, gid, e = LookupGroup(groupp)
	if e != nil {
		log.Fatal("Error looking up group", *groupp, e)
	}

	for {
		e = filepath.Walk(*dirp, walker)
		if e != nil {
			log.Println("Walk", *dirp, e)
		}
		time.Sleep(time.Duration(*secp) * time.Second)
	}
}

func walker(path string, info os.FileInfo, err error) error {
	st := info.Sys().(*syscall.Stat_t)
	if (uid != -1 && int(st.Uid) != uid) || (gid != -1 && int(st.Gid) != gid) {
		e := syscall.Chown(path, uid, gid)
		if e != nil {
			log.Println("chown", path, e)
		}
	}
	var mode = fmode
	if info.IsDir() {
		mode = dmode
	}
	if mode != st.Mode {
		e := syscall.Chmod(path, mode)
		if e != nil {
			log.Println("chmod", path, e)
		}
	}
	return nil
}
