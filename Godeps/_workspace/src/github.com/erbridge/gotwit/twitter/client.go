package twitter

import (
	"net/url"
	"strconv"

	"github.com/ChimeraCoder/anaconda"
)

type (
	Client struct {
		api      *anaconda.TwitterApi
		stream   anaconda.Stream
		callback func(t anaconda.Tweet)
	}
)

func NewClient(accessConfig AccessConfig, callback func(t anaconda.Tweet)) Client {
	return Client{
		api:      anaconda.NewTwitterApi(accessConfig.token, accessConfig.tokenSecret),
		callback: callback,
	}
}

func (c *Client) handleStream() {
	for t := range c.stream.C {
		if t, ok := t.(anaconda.Tweet); ok {
			c.callback(t)
		}
	}
}

func (c *Client) Start() (err error) {
	if ok, err := c.api.VerifyCredentials(); !ok {
		return err
	}

	v := url.Values{
		"replies": {"all"},
	}

	if c.stream, err = c.api.UserStream(v); err != nil {
		return err
	}

	c.handleStream()

	return nil
}

func (c *Client) Stop() (err error) {
	if err = c.stream.Close(); err != nil {
		return
	}

	c.api.Close()

	return
}

func (c *Client) post(message string, v url.Values) error {
	_, err := c.api.PostTweet(message, v)
	return err
}

func (c *Client) Post(message string, nsfw bool) error {
	v := url.Values{
		"possibly_sensitive": {strconv.FormatBool(nsfw)},
	}
	return c.post(message, v)
}

func (c *Client) Reply(tweet anaconda.Tweet, message string, nsfw bool) error {
	v := url.Values{
		"possibly_sensitive":    {strconv.FormatBool(nsfw)},
		"in_reply_to_status_id": {tweet.IdStr},
	}
	return c.post(message, v)
}
