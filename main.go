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

var gConfig struct {
	SlackToken string `json:"slackToken"`
}

func main() {
	var err error
	if err = loadConfig(); err != nil {
		panic(err)
	}

	client := slack.New(gConfig.SlackToken)
	rtm := client.NewRTM()
	go rtm.ManageConnection()

	done := false
	onNonFatalError := func(e error) {
		fmt.Printf("Error: %v\n", e)
	}
	onFatalError := func(e error) {
		fmt.Printf("Fatal error: %v\n", e)
		done = true
	}

	fmt.Println("Hambot activated")

	var session Session
	var dispatcher *Dispatcher

	for {
		select {
		case callback := <-session.Callbacks:
			callback()

		case event, ok := <-rtm.IncomingEvents:
			if !ok {
				done = true
				break
			}
			switch e := event.Data.(type) {
			case *slack.ConnectedEvent:
				session.Start(client, e.Info, rtm)
				if dispatcher, err = NewDispatcher(&session); err != nil {
					panic(err)
				}
				dispatcher.AddHandler(&HamEcho{})

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
			case *slack.InvalidAuthEvent:
				onFatalError(errors.New("InvalidAuthEvent"))
			}
		}

		if done {
			break
		}
	}
}

func loadConfig() error {
	var configPath string
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	} else {
		configDirs := configdir.New("", "hambot").QueryFolders(configdir.Existing)
		if len(configDirs) == 0 {
			return errors.New("Missing configuration directory")
		}
		configPath = filepath.Join(configDirs[0].Path, "config.json")
	}

	var configData []byte
	var err error

	configData, err = ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configData, &gConfig)
	if err != nil {
		return err
	}

	return nil
}
