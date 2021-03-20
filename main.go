package main

import (
	"errors"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type browser struct {
	OpenBrowser       bool `yaml:"openBrowser"`
	CustomBrowserPath bool `yaml:"customBrowserPath"`
	Path              string
	Option            string
}

type replacementData struct {
	ReplaceTarget string `yaml:"replaceTarget"`
	Method        string
	Option        string
}

func (u *replacementData) replaceNew() (*string, error) {
	if u.Method != "DATETIME" {
		return nil, errors.New("not datetime error")
	}
	t := time.Now().Format(u.Option)

	// TODO: go-sed sample
	// engine, err := sed.New(strings.NewReader(u.SED))
	// if err != nil {
	// 	return nil, err
	// }
	// output, err := engine.RunString(t)
	// if err != nil {
	// 	return nil, err
	// }
	// // Remove line feed
	// if output[len(output)-1] == '\n' {
	// 	output = output[:len(output)-1]
	// }

	return &t, nil
}

type template struct {
	Name        string
	Endpoint    string
	Replacement []replacementData
	TO          string
	CC          string
	BCC         string
	Subject     string
	Body        string
}

type config struct {
	Browser  browser
	Template []template
}

func openDefaultBrowser(url string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		return exec.Command("open", url).Start()
	}
	return errors.New("No such as GOOS")
}

func open(url string, execPath string, options ...string) error {
	args := append([]string{url}, options...)
	return exec.Command(execPath, args...).Start()
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func copyFile(srcName string, dstName string) {
	src, err := os.Open(srcName)
	if err != nil {
		panic(err)
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		panic(err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		panic(err)
	}
}

func loadConfigForYaml() (*config, error) {
	target := "local.yml"
	if !fileExists(target) {
		copyFile("default.yml", target)
	}
	f, err := os.Open(target)
	if err != nil {
		log.Fatal("loadConfigForYaml os.Open err:", err)
		return nil, err
	}
	defer f.Close()

	var cfg config
	err = yaml.NewDecoder(f).Decode(&cfg)
	return &cfg, err
}

func main() {
	cfg, err := loadConfigForYaml()
	if err != nil {
		log.Fatal("loadConfigForYaml os.Open err:", err)
		panic(err)
	}
	log.Print(cfg)

	mainTemplate := cfg.Template[0]
	subject := mainTemplate.Subject
	body := mainTemplate.Body
	for _, targetVariable := range mainTemplate.Replacement {
		replaceNew, _ := targetVariable.replaceNew()
		subject = strings.Replace(subject, targetVariable.ReplaceTarget, *replaceNew, -1)
		body = strings.Replace(body, targetVariable.ReplaceTarget, *replaceNew, -1)
	}
	log.Print("Subject: " + subject)
	log.Print("Body: " + body)

	u, err := url.Parse(mainTemplate.Endpoint)
	if err != nil {
		log.Fatal("Error on endpoint url:", err)
		panic(err)
	}

	q := u.Query()
	q.Set("view", "cm")
	q.Set("fs", "1")
	q.Set("tf", "1")
	q.Set("to", mainTemplate.TO)
	q.Set("cc", mainTemplate.CC)
	q.Set("bcc", mainTemplate.BCC)
	q.Set("su", subject)
	q.Set("body", body)
	u.RawQuery = q.Encode()
	log.Print(u.String())

	if !cfg.Browser.OpenBrowser {
		os.Exit(0)
	}

	err = nil
	if cfg.Browser.CustomBrowserPath {
		log.Print("openCustomPath")
		err = open(u.String(), cfg.Browser.Path, cfg.Browser.Option)
	} else {
		log.Print("openDefaultBrowser")
		err = openDefaultBrowser(u.String())
	}
	if err != nil {
		log.Fatal("Failed open browser:", err)
		log.Panic(err)
	}

}
