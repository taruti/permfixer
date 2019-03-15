package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

var secp = flag.Int("sec", 60*60, "Time between checks in seconds")
var userp = flag.String("user", "", "User for chown")
var groupp = flag.String("group", "", "Group for chgrp")
var permfp = flag.String("permf", "", "Permissions for chmod in octal for files")
var permdp = flag.String("permd", "", "Permissions for chmod in octal for directories")
var uid, gid, fmode, dmode uint32

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: [flags] [directories]\n", os.Args[0])
		flag.PrintDefaults()
	}
}

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

	u, e := LookupUser(*userp)
	if e != nil {
		log.Fatal("Error looking up user", *userp, e)
	}
	t, e = strconv.ParseInt(u, 10, 32)
	if e != nil {
		log.Fatal("Error parsing uid for user", *userp, e)
	}
	uid = uint32(t)

	g, e := LookupGroup(*groupp)
	if e != nil {
		log.Fatal("Error looking up group", *groupp, e)
	}
	t, e = strconv.ParseInt(g, 10, 32)
	if e != nil {
		log.Fatal("Error parsing gid for group", *groupp, e)
	}
	gid = uint32(t)
	for _, dir := range flag.Args() {
		dir := dir
		go work(dir)
	}
}

func work(dir string) {
	for {
		err := filepath.Walk(dir, walker)
		if err != nil {
			log.Println("Walk", dir, err)
		}
		time.Sleep(time.Duration(*secp) * time.Second)
	}
}

func walker(path string, info os.FileInfo, err error) error {
	st := info.Sys().(*syscall.Stat_t)
	if st.Uid != uid || st.Gid != gid {
		e := syscall.Chown(path, int(uid), int(gid))
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
