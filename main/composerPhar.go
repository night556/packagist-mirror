package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/ioutil"
	"time"
)

var versions = make(map[string][]stable)
var versionsCache []byte

type stable struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	MinPhp  int    `json:"min-php"`
}

func composerPhar(name string, num int) {

	processName := getProcessName(name, num)

	for {
		// Sleep
		time.Sleep(1 * time.Second)

		// Get latest stable version
		versionUrl := "https://getcomposer.org/versions"
		resp, err := get(versionUrl, processName)
		if err != nil {
			continue
		}

		if resp.StatusCode != 200 {
			continue
		}

		content, err := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			fmt.Println(processName, versionUrl, err.Error())
			continue
		}

		if bytes.Equal(versionsCache, content) {
			fmt.Println(processName, "Update to date:", versionUrl)
			continue
		}
		versionsCache = content

		// Sync versions
		options := []oss.Option{
			oss.ContentType("application/json"),
		}
		_ = putObject(processName, "versions", bytes.NewReader(content), options...)

		// JSON Decode
		err = json.Unmarshal(content, &versions)
		if err != nil {
			errHandler(err)
			continue
		}

		// Like https://getcomposer.org/download/1.9.1/composer.phar
		phar, err := get("https://getcomposer.org"+versions["stable"][0].Path, processName)
		if err != nil {
			continue
		}

		if phar.StatusCode != 200 {
			continue
		}

		composerPhar, err := ioutil.ReadAll(phar.Body)
		_ = putObject(processName, "composer.phar", bytes.NewReader(composerPhar))
		_ = putObject(processName, "download/"+versions["stable"][0].Version+"/composer.phar", bytes.NewReader(composerPhar))
		_ = phar.Body.Close()

		// Like https://getcomposer.org/download/1.9.1/composer.phar.sig

		options = []oss.Option{
			oss.ContentType("application/json"),
		}

		sig, err := get("https://getcomposer.org"+versions["stable"][0].Path+".sig", processName)
		if err != nil {
			continue
		}

		if sig.StatusCode != 200 {
			continue
		}

		composerPharSig, err := ioutil.ReadAll(sig.Body)
		_ = putObject(processName, "composer.phar.sig", bytes.NewReader(composerPharSig), options...)
		_ = putObject(processName, "download/"+versions["stable"][0].Version+"/composer.phar.sig", bytes.NewReader(composerPharSig), options...)
		_ = sig.Body.Close()

		// Sleep
		time.Sleep(6000 * time.Second)
	}
}
