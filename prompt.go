package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"
	"unicode"
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

	log.Printf("Prompt from @%v: %v\n", msg.Session.User(msg.User).Name, unfilteredPrompt)

	p.uploading = true

	config := HamConfig{
		Prompt: prompt,
	}

	go config.Upload(msg.Session,
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
			log.Printf("Upload failed: %v\n", err)
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
