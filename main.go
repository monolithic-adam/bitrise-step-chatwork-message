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
	APIToken stepconf.Secret `env:"api_token"`
	RoomID   stepconf.Secret `env:"room_id"`
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
	apiURL := fmt.Sprintf("https://api.chatwork.com/v2/rooms/%s/messages?body=%s", conf.RoomID.String(), msg)
	data := url.Values{}
	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	req.Header.Add("Content-Type", "application/json; charset=utf-8")
	req.Header.Add("X-ChatWorkToken", conf.APIToken.String())

	client := &http.Client{}
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

func createMessage() string {
	if !success {
		return "[info][title]Notification from Bitrise[/title](dance) Build Passed ✅ - ${BITRISE_APP_TITLE} (${BITRISE_GIT_BRANCH}) \nCommit message: ${BITRISE_GIT_MESSAGE} \nBuild logs: ${BITRISE_BUILD_URL}[/info]"
	}
	return "[info][title]Notification from Bitrise[/title];( Build Error 🚫 - ${BITRISE_APP_TITLE} (${BITRISE_GIT_BRANCH}) \nCommit message: ${BITRISE_GIT_MESSAGE} \nBuild logs: ${BITRISE_BUILD_URL}[/info]"
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

	if err := postMessage(conf, createMessage()); err != nil {
		log.Errorf("Error: %s", err)
		os.Exit(1)
	}

	log.Donef("\nChatwork message successfully sent! 🍺\n")
	os.Exit(0)
}
