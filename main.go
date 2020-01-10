package main

import (
	"fmt"
	"github.com/nlopes/slack"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
)

func getToken() (string, error) {
	f, err := os.Open("./.token")
	if err != nil {
		return "", err
	}

	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimRight(string(b), "\n"), nil
}

func main() {
	token, err := getToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	api := slack.New(token)

	param := slack.ListFilesParameters{
		Limit: 1000,
	}

	files, _, err := api.ListFiles(param)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	for _, file := range files {
		wg.Add(1)
		go func(f slack.File) {
			defer wg.Done()
			if isNeededChannelID(f) {
				err := DL(api, f.URLPrivateDownload, genFilename(f))
				if err != nil {
					log.Println(err)
				}
			}
		}(file)
	}
	wg.Wait()
}

func genFilename(f slack.File) string {
	t := f.Created.Time()
	return fmt.Sprintf("%04d%02d%02d_%s.%s",
		t.Year(), int(t.Month()), t.Day(), f.ID, f.Filetype)
}

func isNeededChannelID(f slack.File) bool {
	needed := []string{
		"CJG26EM5W",
	}
	for _, c := range needed {
		if elemStr(f.Channels, c) {
			return true
		}
	}
	return false
}

func elemStr(xs []string, e string) bool {
	for _, x := range xs {
		if x == e {
			return true
		}
	}
	return false
}

func DL(api *slack.Client, url, filename string) error {
	dir := "./photo/"
	filepath := dir + filename

	err := safeMkdir(dir, 0777)
	if err != nil {
		return err
	}

	if _, err := os.Stat(filepath); !os.IsNotExist(err) {
		return err
	}

	toFile, err := os.Create(dir + filename)
	if err != nil {
		return err
	}

	api.GetFile(url, toFile)

	return nil
}

func safeMkdir(dir string, perm os.FileMode) error {
	err := os.Mkdir(dir, perm)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}
