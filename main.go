package main

import (
	"encoding/json"
	"errors"
	"log"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/nlopes/slack"
	"github.com/shibukawa/configdir"
)

// Settings contains the values specified in hambot's configuration file.
// Some values are optional.
var Settings struct {
	SlackToken         string `json:"slackToken"`         // Slack API token
	AwsRegion          string `json:"awsRegion"`          // AWS region (for prompt upload)
	AwsBucket          string `json:"awsBucket"`          // AWS bucket name (for prompt upload)
	AwsAccessKey       string `json:"awsAccessKey"`       // AWS access key (for prompt upload)
	AwsSecretAccessKey string `json:"awsSecretAccessKey"` // AWS secret access key (for prompt upload)
	HamBase            string `json:"hamBase"`            // Channel used for announcements
	HamagramsURL       string `json:"hamagramsUrl"`       // URL announced by prompt uploader
	Announcer          string `json:"announcer"`          // User allowed to use the 'announce' command
}

func main() {
	var err error
	if err = loadConfig(); err != nil {
		panic(err)
	}

	client := slack.New(Settings.SlackToken)
	rtm := client.NewRTM()
	go rtm.ManageConnection()

	printNonFatalError := func(err error) {
		log.Printf("Error: %v\n", err)
	}
	printFatalError := func(err error) {
		log.Printf("Fatal error: %v\n", err)
	}

	log.Println("Hambot activated")

	var session Session
	var dispatcher *Dispatcher

	onConnect := func(event *slack.ConnectedEvent) {
		session.Start(client, event.Info, rtm)
		if dispatcher, err = NewDispatcher(&session); err != nil {
			panic(err)
		}

		dispatcher.AddHandler(NewHamEcho())

		if hamPrompt, err := NewHamPrompt(); err == nil {
			dispatcher.AddHandler(hamPrompt)
		} else {
			log.Printf("Error creating HamPrompt: %v\n", err)
		}

		if hamAnnounce, err := NewHamAnnounce(); err == nil {
			dispatcher.AddHandler(hamAnnounce)
		} else {
			log.Printf("Error creating HamAnnounce: %v\n", err)
		}
	}

EventLoop:
	for {
		select {
		case callback := <-session.Callbacks:
			callback()

		case baseEvent, ok := <-rtm.IncomingEvents:
			if !ok {
				break EventLoop
			}
			switch event := baseEvent.Data.(type) {
			case *slack.ConnectedEvent:
				onConnect(event)
			case *slack.MessageEvent:
				dispatcher.Dispatch(event)

			// non-fatal errors
			case *slack.UnmarshallingErrorEvent:
				if event == nil {
					log.Println("Error: UnmarshallingErrorEvent(nil)")
				} else {
					printNonFatalError(event)
				}
			case *slack.MessageTooLongEvent:
				if event == nil {
					log.Println("Error: MessageTooLongEvent(nil)")
				} else {
					printNonFatalError(event)
				}
			case *slack.OutgoingErrorEvent:
				if event == nil {
					log.Println("Error: OutgoingErrorEvent(nil)")
				} else {
					printNonFatalError(event)
				}
			case *slack.IncomingEventError:
				if event == nil {
					log.Println("Error: IncomingEventError(nil)")
				} else {
					printNonFatalError(event)
				}
			case *slack.AckErrorEvent:
				if event == nil {
					log.Println("Error: AckErrorEvent(nil)")
				} else {
					printNonFatalError(event)
				}
			case *slack.ConnectionErrorEvent:
				if event == nil {
					log.Println("Error: ConnectionErrorEvent(nil)")
				} else {
					printNonFatalError(event)
				}

			// fatal errors
			case *slack.InvalidAuthEvent:
				if event == nil {
					log.Println("Error: InvalidAuthEvent(nil)")
				} else {
					printFatalError(errors.New("InvalidAuthEvent"))
				}
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

	log.Printf("Loading configuration from %v\n", configPath)

	var configData []byte
	configData, err = ioutil.ReadFile(configPath)
	if err != nil {
		return
	}

	err = json.Unmarshal(configData, &Settings)
	return
}
