package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const IGNORE = "^hourly-[0-2][0-9]$"

var (
	verbose *bool
	strip   *int
	mbuffer *bool
)

type Snap struct {
	Dataset  string
	Snapshot string
}

type SnapTime struct {
	Snapshot string
	Creation int64 // UNIX time, seconds
}

type Chain struct {
	Start SnapTime
	Snaps []SnapTime
}

func dest_filename(dataset string, target string) string {
	split := strings.SplitN(dataset, "/", *strip+1)
	if len(split) < *strip {
		log.Fatal("ZFS dataset ", dataset, " does not have ", *strip,
			" components to strip")
	}
	return path.Join(target, path.Join(split[*strip:len(split)]...))
}

func zfs_send(dataset string, chain Chain, target string, nomount bool) {
	prev := chain.Start
	for _, next := range chain.Snaps {
		if next.Creation < prev.Creation {
			log.Println("skipping src snapshot", dataset+"@"+next.Snapshot,
				"as it is older than", target+"@"+prev.Snapshot)
			continue
		}
		dest := dest_filename(dataset, target)

		var cmd, incr, mbuf, canmount string
		if prev.Snapshot != "" {
			incr = " -i " + prev.Snapshot
		}
		if *mbuffer {
			mbuf = "mbuffer -q -s 128k -m 1G|"
		}
		if nomount {
			canmount = " -o canmount=off"
		}

		cmd = fmt.Sprintf(
			"zfs send -c -v%s %s@%s|%szfs receive%s %s@%s",
			incr, dataset, next.Snapshot, mbuf, canmount, dest, next.Snapshot,
		)
		log.Println(cmd)
		prev = next
		zfs := exec.Command("/bin/sh", "-c", cmd)
		zfs.Stdout = os.Stdout
		zfs.Stderr = os.Stderr
		if err := zfs.Run(); err != nil {
			log.Fatal("failed to run: ", cmd, ": ", err)
		}
	}
}

func filesystems() map[string]bool {
	zfs := exec.Command(
		"/usr/sbin/zfs", "list", "-H", "-t", "filesystem", "-o", "name")
	stdout, err := zfs.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if err := zfs.Start(); err != nil {
		log.Fatal(err)
	}
	fs := make(map[string]bool)
	for scanner.Scan() {
		line := scanner.Text()
		fs[line] = true
	}
	return fs
}

func schedule(source []string, target string) map[string]Chain {
	zfs := exec.Command(
		"/usr/sbin/zfs", "list", "-H", "-t", "snapshot", "-o", "name,creation")
	stdout, err := zfs.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if err := zfs.Start(); err != nil {
		log.Fatal(err)
	}
	exists := make(map[Snap]bool)
	src_snaps := make(map[string][]SnapTime)
	ignore, err := regexp.Compile(IGNORE)
	if err != nil {
		log.Fatal("could not compile ignore regex ", IGNORE, ": ", err)
	}
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println("line:", line)
		val := strings.SplitN(line, "\t", 2)
		if len(val) != 2 {
			log.Fatal("could not parse line: ", line)
		}
		// fmt.Println("val:", val, len(val))
		name := strings.SplitN(val[0], "@", 2)
		if ignore.MatchString(name[1]) {
			if *verbose {
				fmt.Println("skipping", val[0])
			}
			continue
		}
		snapshot := Snap{name[0], name[1]}
		if strings.HasPrefix(snapshot.Dataset, target) {
			exists[snapshot] = true
		}
		match := false
		for _, fs := range source {
			if strings.HasPrefix(snapshot.Dataset, fs) {
				match = true
				break
			}
		}
		if !match {
			continue
		}
		snaps, ok := src_snaps[snapshot.Dataset]
		if !ok {
			snaps = make([]SnapTime, 0)
			src_snaps[snapshot.Dataset] = snaps
		}
		// Linux format
		unix, err := strconv.ParseInt(val[1], 10, 64)
		if err == nil {
			snaps = append(snaps, SnapTime{snapshot.Snapshot, unix})
		} else {
			// Illumos format
			ts, err := time.Parse("Mon Jan 2 15:04 2006", val[1])
			if err != nil {
				log.Fatal("could not parse: ", val[1], ": ", err)
			}
			snaps = append(snaps, SnapTime{snapshot.Snapshot, ts.Unix()})
		}
		src_snaps[snapshot.Dataset] = snaps
	}
	// find the oldest snapshot already present on target, filter out
	// snapshots already present on the target and sort chronologically
	todo := make(map[string]Chain)
	for dataset, snaps := range src_snaps {
		var newest SnapTime

		filtered := make([]SnapTime, 0)

		for _, snap := range snaps {
			if exists[Snap{dest_filename(dataset, target), snap.Snapshot}] {
				if snap.Creation > newest.Creation {
					newest = snap
				}
			} else {
				filtered = append(filtered, snap)
			}
			// sort filtered by ts
			sort.Sort(ByCreation(filtered))

			todo[dataset] = Chain{newest, filtered}
		}
	}
	if *verbose {
		fmt.Println("existing: ", exists)
	}

	return todo
}

type ByCreation []SnapTime

func (a ByCreation) Len() int           { return len(a) }
func (a ByCreation) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByCreation) Less(i, j int) bool { return a[i].Creation < a[j].Creation }

func main() {
	target := flag.String("target", "", "target zpool to replicate to")
	verbose = flag.Bool("v", false, "Verbose error reporting")
	mbuffer = flag.Bool("m", false, "Use mbuffer")
	strip = flag.Int("strip", 0, "Number of pathname components to strip from the source when replicating to the target, e.g. zfsvault --strip 1 -target vault zones/majid will replicate zones/majid to vault/majid, not vault/zones/majid")
	flag.Parse()
	// XXX validate target and args for valid ZFS dataset names here

	if *target == "" {
		log.Fatal("must specify a destination target zpool using --target")
	}

	todo := schedule(flag.Args(), *target)
	fs := filesystems()

	sorted := make([]string, 0)
	for dataset, _ := range todo {
		sorted = append(sorted, dataset)
	}
	sort.Strings(sorted)
	for _, dataset := range sorted {
		chain := todo[dataset]
		zfs_send(dataset, chain, *target, fs[dataset])
	}
}
