package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const date_fmt = "[02/Jan/2006:15:04:05 -0700]"
const iso_8601 = "2006-01-02 15:04:05"
const RETENTION = 14 // retention period in days

func create(snap string) {
	zfs := exec.Command("/usr/sbin/zfs", "snapshot", "-r", snap)
	zfs.Stdout = os.Stdout
	zfs.Stderr = os.Stderr
	if err := zfs.Run(); err != nil {
		log.Fatal("error creating snapshot ", snap, ": ", err)
	}
}

func destroy(snap string) {
	zfs := exec.Command("/usr/sbin/zfs", "destroy", "-r", snap)
	zfs.Stdout = os.Stdout
	zfs.Stderr = os.Stderr
	if err := zfs.Run(); err != nil {
		log.Fatal("error destroying snapshot ", snap, ": ", err)
	}
}

func destroy_old_snaps(datasets []string, snaps map[string]bool) {
	zfs := exec.Command(
		"/usr/sbin/zfs", "list", "-H", "-t", "snapshot", "-o", "name")
	stdout, err := zfs.StdoutPipe()
	if err != nil {
		log.Fatal("error piping listing snapshots: ", err)
	}
	scanner := bufio.NewScanner(stdout)
	if err := zfs.Start(); err != nil {
		log.Fatal("error starting listing snapshots: ", err)
	}
	todo := make([]string, 0)
	for scanner.Scan() {
		snapshot := scanner.Text()
		match := false
		for _, fs := range datasets {
			if strings.HasPrefix(snapshot, fs) {
				match = true
			}
		}
		if !match {
			continue
		}
		snaps[snapshot] = true
		//fmt.Println(snapshot)
		i := strings.Index(snapshot, "@daily-")
		if i != -1 {
			date, err := time.Parse("@daily-2006-01-02", snapshot[i:])
			if err == nil {
				if time.Since(date) > time.Duration(RETENTION*24*time.Hour) {
					todo = append(todo, snapshot)
				}
			}
		}
	}
	if err := zfs.Wait(); err != nil {
		log.Fatal("error waiting to exit: ", err)
	}
	sort.Strings(todo)
	for i := range todo {
		snapshot := todo[len(todo)-1-i]
		//fmt.Println("destroying", snapshot)
		destroy(snapshot)
	}
}

func main() {
	flag.Parse()
	datasets := flag.Args()
	snaps := make(map[string]bool)

	destroy_old_snaps(datasets, snaps)

	now := time.Now()
	for _, fs := range datasets {
		//fmt.Println(i, " ", fs)
		hourly := fmt.Sprintf("%s@%s", fs, now.Format("hourly-15"))
		if snaps[hourly] {
			//fmt.Println("destroying", hourly)
			destroy(hourly)
		}
		//fmt.Println("creating", hourly)
		create(hourly)

		daily := fmt.Sprintf("%s@%s", fs, now.Format("daily-2006-01-02"))
		if !snaps[daily] {
			//fmt.Println("creating", daily)
			create(daily)
		}
	}
}
