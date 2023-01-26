package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gregdel/pushover"
	"gopkg.in/yaml.v3"
)

type servers struct {
	URL   string
	Label string
}

type pushoverCredential struct {
	Token          string `yaml:"app_token"`
	RecipientToken string `yaml:"recipient_token"`
}

type Config struct {
	Servers    []servers
	UserAgent  string             `yaml:"ua"`
	Credential pushoverCredential `yaml:"pushover"`
}

var conf Config
var confFile string

func main() {
	flag.StringVar(&confFile, "config", "conf.yaml", "config filename")
	flag.Parse()

	f, err := os.ReadFile(confFile)
	if err != nil {
		log.Fatalln(err)
	}
	err = yaml.Unmarshal(f, &conf)
	if err != nil {
		log.Fatalln(err)
	}

	client := new(http.Client)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client.Transport = tr

	var msg []string
	for _, i := range conf.Servers {
		// fmt.Println(i.URL)
		req, err := http.NewRequest("GET", i.URL, nil)
		if err != nil {
			log.Fatalln(err)
		}
		req.Header.Add("User-Agent", conf.UserAgent)
		resp, err := client.Do(req)
		if err != nil {
			msg = append(msg, fmt.Sprintf("%s is seems down.", i.Label))
			continue
		}
		if resp.Status != "200 OK" {
			msg = append(msg, fmt.Sprintf("%s is seems down.", i.Label), "Invalid Status Code")
		}
		_, err = io.Copy(io.Discard, resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		resp.Body.Close()
	}
	if len(msg) > 0 {
		err = sendPushover(strings.Join(msg, "<br>"), 0)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func sendPushover(message string, priority int) error {
	app := pushover.New(conf.Credential.Token)
	_, err := app.SendMessage(
		&pushover.Message{
			Message:  message,
			Title:    "uptime checker",
			Priority: priority,
			HTML:     true,
		},
		pushover.NewRecipient(conf.Credential.RecipientToken),
	)
	return err
}
