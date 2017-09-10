package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nlopes/slack"
	"github.com/shibukawa/configdir"
)

// Settings contains the values specified in hambot's configuration file.
var Settings struct {
	SlackToken         string `json:"slackToken"`         // Slack API token
	AwsRegion          string `json:"awsRegion"`          // AWS region (for prompt upload)
	AwsBucket          string `json:"awsBucket"`          // AWS bucket name (for prompt upload)
	AwsAccessKey       string `json:"awsAccessKey"`       // AWS access key (for prompt upload)
	AwsSecretAccessKey string `json:"awsSecretAccessKey"` // AWS secret access key (for prompt upload)
	HamBase            string `json:"hamBase"`            // Channel used for announcements
	HamagramsURL       string `json:"hamagramsUrl"`       // URL announced by prompt uploader
}

func main() {
	var err error
	if err = loadConfig(); err != nil {
		panic(err)
	}

	client := slack.New(Settings.SlackToken)
	rtm := client.NewRTM()
	go rtm.ManageConnection()

	onNonFatalError := func(e error) {
		fmt.Printf("Error: %v\n", e)
	}
	onFatalError := func(e error) {
		fmt.Printf("Fatal error: %v\n", e)
	}

	fmt.Println("Hambot activated")

	var session Session
	var dispatcher *Dispatcher

EventLoop:
	for {
		select {
		case callback := <-session.Callbacks:
			callback()

		case event, ok := <-rtm.IncomingEvents:
			if !ok {
				break EventLoop
			}
			switch e := event.Data.(type) {
			case *slack.ConnectedEvent:
				session.Start(client, e.Info, rtm)
				if dispatcher, err = NewDispatcher(&session); err != nil {
					panic(err)
				}

				dispatcher.AddHandler(NewHamEcho())

				if hamPrompt, err := NewHamPrompt(); err == nil {
					dispatcher.AddHandler(hamPrompt)
				} else {
					panic(err)
				}

			case *slack.MessageEvent:
				dispatcher.Dispatch(e)

			// non-fatal errors
			case *slack.UnmarshallingErrorEvent:
				onNonFatalError(e)
			case *slack.MessageTooLongEvent:
				onNonFatalError(e)
			case *slack.OutgoingErrorEvent:
				onNonFatalError(e)
			case *slack.IncomingEventError:
				onNonFatalError(e)
			case *slack.AckErrorEvent:
				onNonFatalError(e)

			// fatal errors
			case *slack.ConnectionErrorEvent:
				onFatalError(e)
				break EventLoop
			case *slack.InvalidAuthEvent:
				onFatalError(errors.New("InvalidAuthEvent"))
				break EventLoop
			}
		}
	}
}

func loadConfig() (err error) {
	const appName = "hambot"
	const configName = "config.json"

	var configPath string

	// load from argument if given
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	// try executablePath/configName
	if configPath == "" {
		if exePath, err := os.Executable(); err == nil {
			configPath = filepath.Join(filepath.Dir(exePath), configName)
			if _, err = os.Stat(configPath); err != nil {
				configPath = ""
			}
		}
	}

	// try standardConfigPath/appName/configName
	if configPath == "" {
		configDirs := configdir.New("", appName).QueryFolders(configdir.Existing)
		if len(configDirs) == 0 {
			return errors.New("Missing configuration directory")
		}
		configPath = filepath.Join(configDirs[0].Path, configName)
	}

	fmt.Printf("Loading configuration from %v\n", configPath)

	var configData []byte
	configData, err = ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	err = json.Unmarshal(configData, &Settings)
	return
}
