package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var secp = flag.Int("sec", 60*60, "Time between checks in seconds")
var userp = flag.String("user", "", "User for chown")
var groupp = flag.String("group", "", "Group for chgrp")
var delme = flag.Bool("DELME", false, "Remove files/directories named delme")
var verbose = false

var permf = OctalFlag(00660)
var permd = OctalFlag(02770)
var fmode, dmode uint32
var uid = -1
var gid = -1

type OctalFlag uint32

func (o *OctalFlag) String() string {
	return fmt.Sprintf("0%o", uint32(*o))
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
	flag.BoolVar(&verbose, "v", false, "Verbose output")
}

type walkState struct {
	timeSec int64
}

func main() {
	flag.Parse()

	fmode = uint32(permf)
	dmode = uint32(permd)

	var err error
	uid, err = users.Lookup(*userp)
	if err != nil {
		log.Fatalf("Error looking up user %q: %v", *userp, err)
	}

	gid, err = groups.Lookup(*groupp)
	if err != nil {
		log.Fatalf("Error looking up group: %q: %v", *groupp, err)
	}

	log.Printf("Set uid=%d gid=%d dmode=%o fmode=%o", uid, gid, dmode, fmode)

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
	var ws walkState
	for {
		ws.timeSec = int64(time.Now().Unix()) + (15 * 60)
		err := filepath.Walk(dir, ws.walker)
		if err != nil {
			log.Println("Walk", dir, err)
		}
		time.Sleep(time.Duration(*secp) * time.Second)
	}
}

func doesIdNeedChange(want int, have uint32) bool {
	return want >= 0 && want != int(have)
}

func chown(path string, uid int, gid int) {
	if verbose {
		log.Printf("chown %q %d %d", path, uid, gid)
	}
	err := syscall.Chown(path, uid, gid)
	if err != nil {
		log.Printf("chown(%q,%d,%d) => ERROR: %v", path, uid, gid, err)
	}
}

func chmod(path string, mode uint32) {
	if verbose {
		log.Printf("chmod %q %o", path, mode)
	}
	err := syscall.Chmod(path, mode)
	if err != nil {
		log.Printf("chmod(%q,%o) => ERROR: %v", path, mode, err)
	}
}

func deleteDir(path string) {
	err := os.RemoveAll(path)
	if err != nil {
		time.Sleep(100 * time.Millisecond)
		err = os.RemoveAll(path)
	}
	log.Printf("Delete path %q => %v", path, err)
}

func (ws *walkState) walker(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Println("Walk error: ", err)
		return nil
	}
	st := info.Sys().(*syscall.Stat_t)
	if doesIdNeedChange(uid, st.Uid) || doesIdNeedChange(gid, st.Gid) {
		chown(path, uid, gid)
	}
	var mode = fmode
	if info.IsDir() {
		mode = dmode
	}
	if mode != (st.Mode & 07777) {
		chmod(path, mode)
	}
	if st.Atim.Sec > ws.timeSec || st.Ctim.Sec > ws.timeSec || st.Mtim.Sec > ws.timeSec {
		log.Println("Fixing time in future: %q", path)
		tv := syscall.NsecToTimeval(time.Now().UnixNano())
		w := []syscall.Timeval{tv,tv,tv}
		err := syscall.Utimes(path, w)
		if err != nil {
			log.Println("Fixing time in future failed %q: %v", path, err)
		}
	}
	if *delme && info.IsDir() && strings.HasSuffix(path, "/DELME") {
		deleteDir(path)
	}
	return nil
}
