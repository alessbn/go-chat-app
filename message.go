package main

import (
	"time"
)

// mesages represent a single message.
type message struct {
	Name      string
	Message   string
	When      time.Time
	AvatarURL string
}
