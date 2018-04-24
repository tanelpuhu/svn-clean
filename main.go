package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// VERSION ...
const VERSION string = "0.0.2"

var flagVersion bool

func fileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func getDirSize(path string) int64 {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return size

}

func chdir(path string) {
	if err := os.Chdir(path); err != nil {
		panic(err)
	}
}

func runGC(path string) time.Duration {
	wd, _ := os.Getwd()
	defer chdir(wd)
	chdir(path)
	start := time.Now()
	args := []string{"cleanup", "--quiet"}
	if err := exec.Command("svn", args...).Run(); err != nil {
		log.Fatal(err)
	}
	return time.Now().Sub(start)
}

func fmtInt(size int64) string {
	unit := "b"
	// i know, if if if'y
	if size >= 1024 {
		unit, size = "Kb", size/1024
	}
	if size >= 1024 {
		unit, size = "Mb", size/1024
	}
	if size >= 1024 {
		unit, size = "Gb", size/1024
	}
	result := fmt.Sprintf("%d%s", size, unit)
	return result
}

func sizeAndRunGC(path string) {
	sizeBefore := getDirSize(path)
	fmt.Printf("%-64s %11s -> ", path, fmtInt(sizeBefore))
	elapsed := runGC(path)
	sizeAfter := getDirSize(path)
	fmt.Printf("%-14s %10s %8s\n",
		fmtInt(sizeAfter),
		fmt.Sprintf("%.2f%%", 100*float32(sizeAfter)/float32(sizeBefore)),
		fmt.Sprintf("%s", elapsed.Truncate(time.Millisecond).String()),
	)
}

func walkCallback(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatal(err)
	}
	if info.IsDir() && info.Name() == ".svn" {
		basepath, err := filepath.Abs(filepath.Dir(path))
		if err != nil {
			log.Fatal(err)
		}
		if fileExists(filepath.Join(path, "wc.db")) {
			sizeAndRunGC(basepath)
		}
		return filepath.SkipDir
	}
	return nil
}

func checkExec() {
	_, err := exec.LookPath("svn")
	if err != nil {
		log.Fatal("svn not present in $PATH")
	}
}

func init() {
	flag.BoolVar(&flagVersion, "V", false, "Print version")
	flag.Parse()
}

func main() {
	if flagVersion {
		fmt.Printf("svn-clean %v\n", VERSION)
		return
	}
	checkExec()

	args := flag.Args()
	if len(flag.Args()) == 0 {
		args = append(args, ".")
	}
	start := time.Now()
	for _, arg := range args {
		filepath.Walk(arg, walkCallback)
	}
	fmt.Printf("%-105s %8s\n",
		"",
		fmt.Sprintf("%s", time.Now().Sub(start).Truncate(time.Millisecond).String()),
	)

}
