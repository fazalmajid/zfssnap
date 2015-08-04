package main

import (
	"log"
	"os/exec"
	"bufio"
	"fmt"
	"flag"
	"time"
	"strings"
)

const date_fmt = "[02/Jan/2006:15:04:05 -0700]"
const iso_8601 = "2006-01-02 15:04:05"
const RETENTION = 14 // retention period in days

func create(snap string) {
	zfs := exec.Command("/usr/sbin/zfs", "snapshot", "-r", snap)
	if err := zfs.Run(); err != nil {
		log.Fatal(err)
	}
}

func destroy(snap string) {
	zfs := exec.Command("/usr/sbin/zfs", "destroy", "-r", snap)
	if err := zfs.Run(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	flag.Parse()
	
	zfs := exec.Command("/usr/sbin/zfs", "list", "-H", "-t", "snapshot", "-o", "name")
	stdout, err := zfs.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if err := zfs.Start(); err != nil {
		log.Fatal(err)
	}
	var snaps map[string]bool
	snaps = make(map[string]bool)
	for scanner.Scan() {
		snapshot := scanner.Text()
		snaps[snapshot] = true
		//fmt.Println(snapshot)
		i := strings.Index(snapshot, "@daily-")
		if i != -1 {
			date, err := time.Parse("@daily-2006-01-02", snapshot[i:])
			if err == nil {
				if time.Since(date) > time.Duration(RETENTION * 24 * time.Hour) {
					//fmt.Println("destroying", snapshot)
					destroy(snapshot)
				}
			}
		}
	}
	now := time.Now()
	for _, fs := range(flag.Args()) {
		//fmt.Println(i, " ", fs)
		hourly := fmt.Sprintf("%s@%s", fs, now.Format("hourly-15"))
		if snaps[hourly] {
			//fmt.Println("destroying", hourly)
			destroy(hourly)
		}
		//fmt.Println("creating", hourly)
		create(hourly)
		
		daily := fmt.Sprintf("%s@%s", fs, now.Format("daily-2006-01-02"))
		if ! snaps[daily] {
			//fmt.Println("creating", daily)
			create(daily)
		}
	}
	
	if err := zfs.Wait(); err != nil {
		log.Fatal(err)
	}
}
