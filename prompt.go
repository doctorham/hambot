package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// HamPrompt uploads a new hamagram prompt.
type HamPrompt struct {
	rePrompt  *regexp.Regexp
	uploading bool
	history   map[string]([]time.Time)
}

// NewHamPrompt creates a new HamPrompt.
func NewHamPrompt() (*HamPrompt, error) {
	this := HamPrompt{}

	var err error
	this.rePrompt, err = regexp.Compile(`^prompt\s+(.+)$`)
	if err != nil {
		return nil, err
	}

	this.history = make(map[string]([]time.Time))

	return &this, nil
}

type hamagramConfig struct {
	Prompt string `json:"prompt"`
}

// HandleMessage handles a message.
func (p *HamPrompt) HandleMessage(msg Message) bool {
	const minPromptLength = 2
	const maxPromptLength = 64

	matches := p.rePrompt.FindStringSubmatch(msg.DirectText)
	if matches == nil {
		return false
	}

	if p.uploading {
		msg.Reply("Sorry, I'm currently uploading a prompt. :ham:")
		return true
	}

	unfilteredPrompt := matches[1]
	prompt := p.filterPrompt(unfilteredPrompt, false)
	promptWithSpaces := p.filterPrompt(unfilteredPrompt, true)

	if len(prompt) < minPromptLength {
		msg.Reply("Sorry, that prompt is too short. :ham:")
		return true
	}
	if len(prompt) > maxPromptLength {
		msg.Reply("Sorry, that prompt is too long. :ham:")
		return true
	}

	if p.tooManyChanges(msg.User) {
		msg.Reply("Sorry, you've changed the prompt too many times recently. :ham:")
		return true
	}

	data, err := p.generateConfig(prompt)
	if err != nil {
		msg.Reply("Sorry, something went wrong. :ham:")
		fmt.Println(err)
		return true
	}

	fmt.Printf("Prompt from @%v: %v\n", msg.Session.User(msg.User).Name, string(data))

	p.uploading = true
	go p.upload(msg, data,
		func() {
			if hamBase, _ := msg.Session.HamBase(); msg.Channel != hamBase {
				msg.Reply("I uploaded the prompt. :ham:")
			}
			p.history[msg.User] = append(p.history[msg.User], time.Now())

			msg.Session.Announce(
				fmt.Sprintf(
					"New prompt submitted by <@%v>:\n:sparkles:*%v*:sparkles:\n%v\nHam a nice day. :ham:",
					msg.User, promptWithSpaces, Settings.HamagramsURL))
		},
		func(err error) {
			fmt.Printf("Upload failed: %v\n", err)
			msg.Reply("Sorry, I couldn't upload the prompt. :ham:")
		},
		func() {
			p.uploading = false
		})

	return true
}

func (*HamPrompt) filterPrompt(
	prompt string,
	preserveSpaces bool,
) string {
	// convert to uppercase and remove non-alphabetic characters
	return strings.Map(
		func(r rune) rune {
			if (r >= 'A' && r <= 'Z') || (preserveSpaces && unicode.IsSpace(r)) {
				return r
			} else if r >= 'a' && r <= 'z' {
				return 'A' + (r - 'a')
			} else {
				return -1
			}
		}, prompt)
}

func (p *HamPrompt) tooManyChanges(user string) bool {
	const maxHistoryLength = 3
	const historyHours = 4.

	// remove history entries that are too old
	now := time.Now()
	var history []time.Time
	for _, entry := range p.history[user] {
		if now.Sub(entry).Hours() < historyHours {
			history = append(history, entry)
		}
	}
	p.history[user] = history

	return len(history) >= maxHistoryLength
}

func (p *HamPrompt) generateConfig(prompt string) (data []byte, err error) {
	data, err = json.Marshal(&hamagramConfig{
		Prompt: prompt,
	})
	if err != nil {
		return
	}

	data = append([]byte("config="), data...)
	data = append(data, byte(';'))
	return
}

func (p *HamPrompt) upload(
	msg Message,
	data []byte,
	then func(),
	catch func(error),
	finally func(),
) {
	defer func() {
		msg.Session.Callbacks <- finally
	}()

	onError := func(err error) {
		msg.Session.Callbacks <- func() {
			catch(err)
		}
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(Settings.AwsRegion),
		Credentials: credentials.NewStaticCredentials(
			Settings.AwsAccessKey, Settings.AwsSecretAccessKey, ""),
	})
	if err != nil {
		onError(err)
		return
	}

	uploader := s3manager.NewUploader(awsSession)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(Settings.AwsBucket),
		Key:         aws.String("config.js"),
		ContentType: aws.String("application/javascript"),
		ACL:         aws.String("public-read"),
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		onError(err)
		return
	}

	msg.Session.Callbacks <- then
}
