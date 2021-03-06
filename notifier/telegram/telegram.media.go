package telegram

import (
	"log"
	"time"
)

type MediaNotifier struct {
	notifier *Notifier

	// caption(could be empty): images, only work for Merge Images
	channel chan map[string][]string
}

func NewMediaNotifier(token, chatID string, disableNotification bool) *MediaNotifier {
	n := &MediaNotifier{
		&Notifier{
			token, chatID, "", disableNotification, false,
		},
		make(chan map[string][]string, 10),
	}

	go debounce(500*time.Millisecond, n.channel, func(args map[string][]string) {
		// url: caption
		var buffer []*textMediaPair

		clearAndSend := func() {
			if len(buffer) > 0 {
				err := n.notifier.notify(buffer)
				if err != nil {
					log.Println("error:", err)
				}

				buffer = nil
			}
		}

		for caption, images := range args {
			pairs := toPairs(caption, images)

			if len(buffer)+len(pairs) <= 10 {
				buffer = append(buffer, pairs...)
				continue
			} else {
				clearAndSend()
			}

			to := 0
			for to+10 < len(pairs) {
				buffer = append(buffer, pairs[to:to+10]...)
				clearAndSend()
				to += 10
			}
		}

		clearAndSend()
	})

	return n
}

func (n *MediaNotifier) Notify(id, text string, images []string) error {
	n.channel <- map[string][]string{
		text: images,
	}

	// fmt.Println("notify:", id, text, images)

	// TODO: handle error
	return nil
}

func debounce(interval time.Duration, input chan map[string][]string, callback func(args map[string][]string)) {
	timer := time.NewTimer(interval)
	var results map[string][]string

	for {
		select {
		case item := <-input:
			if results == nil {
				results = make(map[string][]string)
			}

			for k, v := range item {
				results[k] = append(results[k], v...)
			}

			timer.Reset(interval)
		case <-timer.C:
			if len(results) != 0 {
				callback(results)
				results = nil
			}
		}
	}
}
