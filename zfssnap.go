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

const (
	date_fmt = "[02/Jan/2006:15:04:05 -0700]"
	iso_8601 = "2006-01-02 15:04:05"
)

var (
	verbose   *bool
	retention *int64
)

func create(snap string) {
	if *verbose {
		log.Println("creating ", snap)
	}
	zfs := exec.Command("/usr/sbin/zfs", "snapshot", "-r", snap)
	zfs.Stdout = os.Stdout
	zfs.Stderr = os.Stderr
	if err := zfs.Run(); err != nil {
		log.Fatal("error creating snapshot ", snap, ": ", err)
	}
}

func destroy(snap string) {
	if *verbose {
		log.Println("destroying ", snap)
	}
	zfs := exec.Command("/usr/sbin/zfs", "destroy", "-r", snap)
	zfs.Stdout = os.Stdout
	zfs.Stderr = os.Stderr
	if err := zfs.Run(); err != nil {
		log.Fatal("error destroying snapshot ", snap, ": ", err)
	}
}

type Snaps []string

func (a Snaps) Len() int      { return len(a) }
func (a Snaps) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a Snaps) Less(i, j int) bool {
	si := strings.SplitN(a[i], "@", 2)
	sj := strings.SplitN(a[j], "@", 2)
	switch {
	case si[0] < sj[0]:
		return true
	case si[0] > sj[0]:
		return false
	default:
		return si[1] < sj[1]
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
		i := strings.Index(snapshot, "@daily-")
		if i != -1 {
			date, err := time.Parse("@daily-2006-01-02", snapshot[i:])
			if err == nil {
				if time.Since(date) > time.Duration(*retention)*24*time.Hour {
					todo = append(todo, snapshot)
				}
			}
		}
	}
	if err := zfs.Wait(); err != nil {
		log.Fatal("error waiting to exit: ", err)
	}
	sort.Sort(sort.Reverse(Snaps(todo)))
	for _, snapshot := range todo {
		destroy(snapshot)
	}
}

func main() {
	verbose = flag.Bool("v", false, "Verbose logging")
	retention = flag.Int64("d", 14, "retention for daily snapshots in days")
	flag.Parse()
	datasets := flag.Args()
	snaps := make(map[string]bool)

	destroy_old_snaps(datasets, snaps)

	now := time.Now()
	for _, fs := range datasets {
		hourly := fmt.Sprintf("%s@%s", fs, now.Format("hourly-15"))
		if snaps[hourly] {
			destroy(hourly)
		}
		create(hourly)

		daily := fmt.Sprintf("%s@%s", fs, now.Format("daily-2006-01-02"))
		if !snaps[daily] {
			create(daily)
		}
	}
}
