package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	rancher "github.com/rancher/go-rancher/v2"
)

func main() {
	for {
		setMemoryLimit()
		time.Sleep(time.Hour)
	}
}

func setMemoryLimit() {
	opts := rancher.ClientOpts{
		Url:       os.Getenv("CATTLE_CONFIG_URL"),
		AccessKey: os.Getenv("CATTLE_ACCESS_KEY"),
		SecretKey: os.Getenv("CATTLE_SECRET_KEY"),
	}
	client, err := rancher.NewRancherClient(&opts)
	if err != nil {
		log.Fatal(err)
	}

	hosts, err := client.Host.List(nil)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Get("http://rancher-metadata/latest/self/host/uuid")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	uuidBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	uuid := string(uuidBytes)

	var host *rancher.Host
	for _, h := range hosts.Data {
		if h.Uuid == uuid {
			host = &h
			break
		}
	}

	if host == nil {
		log.Fatal("Could not find current host")
	}

	info := host.Info.(map[string]interface{})
	memoryInfo := info["memoryInfo"].(map[string]interface{})
	memTotal := memoryInfo["memTotal"].(float64)
	memTotalBytes := int64(memTotal) * 1024 * 1024
	limit := memTotalBytes - 512*1024*1024
	host.Memory = limit

	host, err = client.Host.Update(host, host)
	if err != nil {
		log.Fatal(err)
	}

	memLimitFileName := os.Getenv("MEM_LIMIT_FILE")
	if memLimitFileName == "" {
		memLimitFileName = "/host-sys/fs/cgroup/memory/docker/memory.limit_in_bytes"
	}

	f, err := os.OpenFile(memLimitFileName, os.O_WRONLY, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()
	_, err = f.WriteString(strconv.FormatInt(limit, 10))
	if err != nil {
		log.Fatal(err)
	}
}
