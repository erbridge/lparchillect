package twitter

import (
	"github.com/ChimeraCoder/anaconda"
)

func BindConsumer(consumerConfig ConsumerConfig) {
	anaconda.SetConsumerKey(consumerConfig.key)
	anaconda.SetConsumerSecret(consumerConfig.secret)
}

func UnbindConsumer() {
	anaconda.SetConsumerKey("")
	anaconda.SetConsumerSecret("")
}
