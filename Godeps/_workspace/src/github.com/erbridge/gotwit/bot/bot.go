package bot

import (
	"errors"

	"github.com/ChimeraCoder/anaconda"
	"github.com/erbridge/gotwit/callback"
	"github.com/erbridge/gotwit/twitter"
)

type (
	Bot struct {
		screenName     string
		consumerConfig twitter.ConsumerConfig
		client         twitter.Client
		callbacks      map[callback.Type]map[int]callback.Callback
		nextCallbackId int
	}
)

var (
	activeBot *Bot
)

func New(screenName string, consumerConfig twitter.ConsumerConfig, accessConfig twitter.AccessConfig) Bot {
	b := Bot{
		screenName:     screenName,
		consumerConfig: consumerConfig,
		callbacks:      make(map[callback.Type]map[int]callback.Callback),
	}
	b.client = twitter.NewClient(accessConfig, b.handle)
	return b
}

func (b *Bot) ScreenName() string {
	return b.screenName
}

func (b *Bot) Start() error {
	if activeBot == nil {
		activeBot = b
	} else if activeBot == b {
		return errors.New("This bot is already started")
	} else {
		return errors.New("Only one bot can be active at once")
	}

	twitter.BindConsumer(b.consumerConfig)
	b.client.Start()

	return nil
}

func (b *Bot) Stop() error {
	if activeBot == b {
		activeBot = nil
	} else {
		return errors.New("This bot hasn't been started")
	}

	twitter.UnbindConsumer()
	b.client.Stop()

	return nil
}

func (b *Bot) RegisterCallback(t callback.Type, cb callback.Callback) (id int) {
	if _, ok := b.callbacks[t]; !ok {
		b.callbacks[t] = make(map[int]callback.Callback)
	}

	id = b.nextCallbackId
	b.callbacks[t][id] = cb
	b.nextCallbackId++
	return
}

func (b *Bot) UnregisterCallback(t callback.Type, id int) {
	delete(b.callbacks[t], id)
}

func (b *Bot) triggerCallback(t callback.Type, tweet anaconda.Tweet) {
	for _, cb := range b.callbacks[t] {
		cb(tweet)
	}
}

func (b *Bot) handle(tweet anaconda.Tweet) {
	if tweet.RetweetedStatus != nil {
		b.triggerCallback(callback.Retweet, tweet)
		return
	}

	if tweet.InReplyToScreenName == b.screenName {
		b.triggerCallback(callback.Reply, tweet)
		return
	}

	for _, entity := range tweet.Entities.User_mentions {
		if entity.Screen_name == b.screenName {
			b.triggerCallback(callback.Mention, tweet)
			return
		}
	}

	b.triggerCallback(callback.Post, tweet)
}

func (b *Bot) Post(message string, nsfw bool) error {
	return b.client.Post(message, nsfw)
}

func (b *Bot) Reply(tweet anaconda.Tweet, message string, nsfw bool) error {
	return b.client.Reply(tweet, message, nsfw)
}
