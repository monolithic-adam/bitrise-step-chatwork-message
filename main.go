package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
)

// Config ...
type Config struct {
	// Message
	APIToken       stepconf.Secret `env:"api_token"`
	RoomID         stepconf.Secret `env:"room_id"`
	AppTitle       string          `env:"app_title"`
	GitBranch      string          `env:"git_branch"`
	GitMessage     string          `env:"git_message"`
	BuildURL       string          `env:"build_url"`
	InstallPageURL string          `env:"install_page_url"`
}

var success = os.Getenv("BITRISE_BUILD_STATUS") == "0"

func validate(conf *Config) error {
	if conf.APIToken == "" && conf.RoomID == "" {
		return fmt.Errorf("The API Token and the Room ID are empty. Please Enter them both to continue")
	}

	return nil
}

// postMessage sends a message to a channel.
func postMessage(conf Config, msg string) error {
	client := &http.Client{}
	testStr := url.QueryEscape(msg)
	println(testStr)
	apiURL := fmt.Sprintf("https://api.chatwork.com/v2/rooms/%s/messages?body=%s", string(conf.RoomID), url.QueryEscape(msg))
	data := url.Values{}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-ChatWorkToken", string(conf.APIToken))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send the request: %s", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("server error: %s, failed to read response: %s", resp.Status, err)
		}
		return fmt.Errorf("server error: %s, response: %s", resp.Status, body)
	}

	return nil
}

func createMessage(conf Config) string {
	var messageString = "[info][title]Notification from Bitrise[/title]"
	if !success {
		messageString += fmt.Sprintf(";( Build Error 🚫 - %s (%s) \nCommit message: %s \nBuild logs: %s", conf.AppTitle, conf.GitBranch, conf.GitMessage, conf.BuildURL)
		return messageString
	}

	messageString += fmt.Sprintf("(dance) Build Passed ✅ - %s (%s) \nCommit message: %s \nBuild logs: %s", conf.AppTitle, conf.GitBranch, conf.GitMessage, conf.BuildURL)
	if conf.InstallPageURL != "" {
		messageString += fmt.Sprintf(" \nInstall Page: %s", conf.InstallPageURL)
	}
	messageString += "[/info]"
	return messageString
}

func main() {
	var conf Config

	if err := stepconf.Parse(&conf); err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}
	stepconf.Print(conf)

	if err := validate(&conf); err != nil {
		log.Errorf("Error: %s\n", err)
		os.Exit(1)
	}

	if err := postMessage(conf, createMessage(conf)); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Donef("\nChatwork message successfully sent! 🍺\n")
	os.Exit(0)
}
