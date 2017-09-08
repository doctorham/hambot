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

type HamPrompt struct {
	rePrompt  *regexp.Regexp
	uploading bool
	history   map[string]([]time.Time)
}

func NewHamPrompt() (*HamPrompt, error) {
	this := HamPrompt{}

	var err error
	this.rePrompt, err = regexp.Compile("^prompt\\s+(.+)$")
	if err != nil {
		return nil, err
	}

	this.history = make(map[string]([]time.Time))

	return &this, nil
}

type hamagramConfig struct {
	Prompt string `json:"prompt"`
}

func (p *HamPrompt) HandleMessage(message Message) bool {
	const minPromptLength = 2
	const maxPromptLength = 64

	matches := p.rePrompt.FindStringSubmatch(message.DirectText)
	if matches == nil {
		return false
	}

	if p.uploading {
		message.Reply("Sorry, I'm currently uploading a prompt. ham")
		return true
	}

	unfilteredPrompt := matches[1]
	prompt := p.filterPrompt(unfilteredPrompt, false)
	promptWithSpaces := p.filterPrompt(unfilteredPrompt, true)

	if len(prompt) < minPromptLength {
		message.Reply("Sorry, that prompt is too short. ham")
		return true
	}
	if len(prompt) > maxPromptLength {
		message.Reply("Sorry, that prompt is too long. ham")
		return true
	}

	if p.tooManyChanges(message.User) {
		message.Reply("Sorry, you've changed the prompt too many times recently. ham")
		return true
	}

	data, err := p.generateConfig(prompt)
	if err != nil {
		message.Reply("Sorry, something went wrong. ham")
		fmt.Println(err)
		return true
	}

	fmt.Printf("Prompt from @%v: %v\n", message.Session.GetUser(message.User).Name, string(data))

	p.uploading = true
	go p.upload(message, data,
		func() {
			message.Reply("I uploaded the prompt. ham")
			p.history[message.User] = append(p.history[message.User], time.Now())

			message.Session.Announce(
				fmt.Sprintf(
					"New prompt submitted by <@%v>:\n%v\nhttp://the.ham.doctor\nHam a nice day.",
					message.User, promptWithSpaces))
		},
		func(err error) {
			fmt.Printf("Upload failed: %v\n", err)
			message.Reply("Sorry, I couldn't upload the prompt. ham")
		},
		func() {
			p.uploading = false
		})

	return true
}

func (*HamPrompt) filterPrompt(prompt string, preserveSpaces bool) string {
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

func (p *HamPrompt) generateConfig(prompt string) ([]byte, error) {
	var err error
	var data []byte
	data, err = json.Marshal(&hamagramConfig{
		Prompt: prompt,
	})
	if err != nil {
		return nil, err
	}

	data = append([]byte("config="), data...)
	data = append(data, byte(';'))

	return data, nil
}

func (p *HamPrompt) upload(message Message, data []byte, then func(), catch func(error), finally func()) {
	defer func() {
		message.Session.Callbacks <- finally
	}()

	onError := func(err error) {
		message.Session.Callbacks <- func() {
			catch(err)
		}
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(gConfig.AwsRegion),
		Credentials: credentials.NewStaticCredentials(
			gConfig.AwsAccessKey, gConfig.AwsSecretAccessKey, ""),
	})
	if err != nil {
		onError(err)
		return
	}

	uploader := s3manager.NewUploader(awsSession)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(gConfig.AwsBucket),
		Key:         aws.String("config.js"),
		ContentType: aws.String("application/javascript"),
		ACL:         aws.String("public-read"),
		Body:        bytes.NewReader(data),
	})
	if err != nil {
		onError(err)
		return
	}

	message.Session.Callbacks <- then
}
