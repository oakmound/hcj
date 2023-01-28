package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	testDirectories := []string{
		//"https://www.w3.org/Style/CSS/Test/CSS3/Color/current/html4/",
		//"https://www.w3.org/Style/CSS/Test/CSS3/Selectors/current/html/",
		"http://test.csswg.org/suites/css-ui-3_dev/nightly-unstable/html/chapter-3.htm",
	}
	for _, dir := range testDirectories {
		dirResp, err := http.DefaultClient.Get(dir)
		if err != nil {
			return err
		}
		if dirResp.StatusCode < 200 || dirResp.StatusCode > 299 {
			dirResp.Body.Close()
			return fmt.Errorf("unexpected response code for directory %v", dirResp.StatusCode)
		}
		respData, err := io.ReadAll(dirResp.Body)
		if err != nil {
			dirResp.Body.Close()
			return fmt.Errorf("failed to read response data: %v", err)
		}
		dirResp.Body.Close()
		// TODO: this might be too strict
		// regex for color html
		rg := regexp.MustCompile(`\<a href=\"([^\>]*\.htm)\"`)
		// for Selectors
		//rg := regexp.MustCompile(`\<a href=\"(tests[^\>]*\.html)\"`)
		matches := rg.FindAllSubmatch(respData, -1)
		u, err := url.Parse(dir)
		if err != nil {
			return fmt.Errorf("failed to read input as url: %v", err)
		}
		u.Path = path.Dir(u.Path)
		fmt.Println(u.Path)
		dirRoot := u.String()
		for _, matchList := range matches {
			for i, match := range matchList {
				if i == 0 {
					continue
				}
				if bytes.Contains(match, []byte("reference")) {
					continue
				}
				htmResp, err := http.DefaultClient.Get(dirRoot + "/" + string(match))
				if err != nil {
					return fmt.Errorf("failed to query test data: %v", err)
				}
				if htmResp.StatusCode < 200 || htmResp.StatusCode > 299 {
					htmResp.Body.Close()
					fmt.Println(dirRoot + "/" + string(match))
					return fmt.Errorf("unexpected response code for test %v", htmResp.StatusCode)
				}
				created, err := os.Create(filepath.Join("htmlin", string(match)))
				if err != nil {
					htmResp.Body.Close()
					return fmt.Errorf("unable to create testdata file: %v", err)
				}
				_, err = io.Copy(created, htmResp.Body)
				if err != nil {
					htmResp.Body.Close()
					created.Close()
					return fmt.Errorf("unable to write to testdata file: %v", err)
				}
				htmResp.Body.Close()
				created.Close()
			}
		}
	}
	return nil
}
