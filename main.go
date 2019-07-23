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
	client := &http.Client{}
	apiURL := fmt.Sprintf("https://api.chatwork.com/v2/rooms/%s/messages?body=%s", string(conf.RoomID), msg)
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

func createMessage() string {
	if !success {
		return "%5Binfo%5D%5Btitle%5DNotification%20from%20Bitrise%5B%2Ftitle%5D(dance)%20Build%20Passed%20%E2%9C%85%20-%20%24%7BBITRISE_APP_TITLE%7D%20(%24%7BBITRISE_GIT_BRANCH%7D)%20%5C%5CnCommit%20message%3A%20%24%7BBITRISE_GIT_MESSAGE%7D%20%5C%5CnBuild%20logs%3A%20%24%7BBITRISE_BUILD_URL%7D%5B%2Finfo%5D"
	}
	return "%5Binfo%5D%5Btitle%5DNotification%20from%20Bitrise%5B%2Ftitle%5D%3B(%20Build%20Error%20%F0%9F%9A%AB%20-%20%24%7BBITRISE_APP_TITLE%7D%20(%24%7BBITRISE_GIT_BRANCH%7D)%20%5C%5CnCommit%20message%3A%20%24%7BBITRISE_GIT_MESSAGE%7D%20%5C%5CnBuild%20logs%3A%20%24%7BBITRISE_BUILD_URL%7D%5B%2Finfo%5D"
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
