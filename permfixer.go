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
var permf = OctalFlag(0)
var permd = OctalFlag(0)
var uid, gid, fmode, dmode uint32

type OctalFlag uint32

func (o *OctalFlag) String() string {
	return fmt.Sprint(uint32(*o))
}
func (o *OctalFlag) Set(s string) error {
	v, err := strconv.ParseInt(s, 8, 32)
	if err != nil {
		return err
	}
	*o = OctalFlag(uint32(v))
	return nil
}

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s: [flags] {directories}\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Var(&permf, "permf", "Permissions for chmod in octal for files")
	flag.Var(&permd, "permd", "Permissions for chmod in octal for directories")
}

func main() {
	flag.Parse()

	fmode = uint32(permf)
	dmode = uint32(permd)

	var err error
	if *userp != "" {
		uid, err = users.Lookup(*userp)
		if err != nil {
			log.Fatalf("Error looking up user %q: %v", *userp, err)
		}
	}

	if *groupp != "" {
		gid, err = groups.Lookup(*groupp)
		if err != nil {
			log.Fatalf("Error looking up group: %q: %v", *groupp, err)
		}
	}

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}

	for _, dir := range flag.Args() {
		dir := dir
		go work(dir)
	}

	select {}
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
