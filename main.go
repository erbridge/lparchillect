package main

import (
	"os"

	"github.com/ChimeraCoder/anaconda"
	"github.com/erbridge/gotwit/bot"
	"github.com/erbridge/gotwit/callback"
	"github.com/erbridge/gotwit/twitter"
)

var (
	validParams = []map[string]struct{}{
		{
			"triangle":    {},
			"voronoi":     {},
			"subdivision": {},
			"iso":         {},
			"rect":        {},
			"square":      {},
			"kite":        {},
			"angles":      {},
			"broken":      {},
			"minimal":     {},
		},
		{
			"flat":     {},
			"gradient": {},
			"soft":     {},
			"strong":   {},
			"burst":    {},
			"dial":     {},
			"flow":     {},
			"uni":      {},
			"perp":     {},
		},
		{
			"grain":   {},
			"nograin": {},
			"raster":  {},
		},
		{
			"fat":  {},
			"thin": {},
		},
		{
			"black": {},
			"white": {},
		},
		{
			"colorize":    {},
			"randomcolor": {},
			"spectra":     {},
			"gram":        {},
			"overdrive":   {},
		},
		{
			"flip":      {},
			"shuffle":   {},
			"eleven":    {},
			"gram":      {},
			"overdrive": {},
		},
		{
			"face": {},
		},
	}
)

var (
	// TODO: Need to read these from the the last valid post on start,
	//       rather than relying on state.
	params []string
)

func createSetParamsCallback(b *bot.Bot) func(anaconda.Tweet) {
	return func(t anaconda.Tweet) {
		sender := t.User.ScreenName
		botName := b.ScreenName()

		if sender == botName {
			return
		}

		reset := false
		// Do we need to prevent duplicates from each category?
		for _, ht := range t.Entities.Hashtags {
			for _, v := range validParams {
				if _, ok := v[ht.Text]; ok {
					if reset {
						params = []string{}
						reset = true
					}
					params = append(params, ht.Text)
					break
				}
			}
		}

		if reset {
			text := "@" + sender + " set new parameters:"

			for _, p := range params {
				text += " #" + p
			}

			b.Reply(t, text, false)
		}
	}
}

func createLowPolyTweetCallback(b *bot.Bot) func(anaconda.Tweet) {
	return func(t anaconda.Tweet) {
		if t.User.ScreenName != "archillect" || len(t.Entities.Media) == 0 {
			return
		}

		text := "@Lowpolybot"

		for _, p := range params {
			text += " #" + p
		}

		if t.PossiblySensitive {
			text += " #NSFW"
		}

		text += " " + t.Entities.Media[0].Url

		b.Post(text, t.PossiblySensitive)
	}
}

func createLowPolyRepostCallback(b *bot.Bot) func(anaconda.Tweet) {
	return func(t anaconda.Tweet) {
		if t.User.ScreenName != "Lowpolybot" || len(t.Entities.Media) == 0 {
			return
		}

		b.Post(t.Entities.Media[0].Url, t.PossiblySensitive)
	}
}

func main() {
	var (
		con twitter.ConsumerConfig
		acc twitter.AccessConfig
	)

	f := "secrets.json"
	if _, err := os.Stat(f); err == nil {
		con, acc, _ = twitter.LoadConfigFile(f)
	} else {
		con, acc, _ = twitter.LoadConfigEnv()
	}

	b := bot.New("lparchillect", con, acc)

	b.RegisterCallback(callback.Mention, createSetParamsCallback(&b))
	b.RegisterCallback(callback.Reply, createLowPolyRepostCallback(&b))
	b.RegisterCallback(callback.Post, createLowPolyTweetCallback(&b))

	b.Start()
	b.Stop()
}
