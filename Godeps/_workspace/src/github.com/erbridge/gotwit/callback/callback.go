package callback

import (
	"github.com/ChimeraCoder/anaconda"
)

type (
	Callback func(anaconda.Tweet)

	Type int
)

const (
	Retweet Type = iota
	Reply
	Mention
	Post
)
