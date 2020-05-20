package main

import (
	"fmt"
	"github.com/RadhiFadlillah/go-sastrawi"
	"github.com/jbrukh/bayesian"
	"github.com/muhfajar/queue"
	twt "github.com/n0madic/twitter-scraper"
	"io/ioutil"
	"strings"
	"time"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

type dataset struct {
	tweet     string
	tokenizer []string
	minScore  float64
	maxScore  float64
	sentiment string
	likely    int
	strict    bool
	timestamp int64
}

var data []dataset

func main() {
	trends, err := twt.GetTrends()
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(&queue.Worker{
		Thread: 3,
		Alloc:  len(trends),
		Set: queue.Callback{
			TaskDone: func(result interface{}) {
				trend := result.(string)
				sentiment(trend)
			},
			QueueDone: func() {
				fmt.Println("done")
			},
		},
	})

	for _, trend := range trends {
		value := trend
		task := func() interface{} {
			return value
		}

		q.Append(task)
	}

	q.Start()
}

func sentiment(trend string) {
	tokenizer := sastrawi.NewTokenizer()
	classifier := bayesian.NewClassifierTfIdf(Good, Bad)

	idp, _ := ioutil.ReadFile("lexicon/id/positive.txt")
	enp, _ := ioutil.ReadFile("lexicon/en/positive.txt")
	goodStuff := strings.Split(string(idp), "\n")
	goodStuff = append(goodStuff, strings.Split(string(enp), "\n")...)

	idn, _ := ioutil.ReadFile("lexicon/id/negative.txt")
	enn, _ := ioutil.ReadFile("lexicon/en/negative.txt")
	badStuff := strings.Split(string(idn), "\n")
	badStuff = append(badStuff, strings.Split(string(enn), "\n")...)

	classifier.Learn(goodStuff, Good)
	classifier.Learn(badStuff, Bad)
	classifier.ConvertTermsFreqToTfIdf()

	tweets := twt.SearchTweets(trend, 50)

	q := queue.NewQueue(&queue.Worker{
		Thread: 3,
		Alloc:  len(tweets),
		Set: queue.Callback{
			TaskDone: func(result interface{}) {
				t := result.(*twt.Result)
				if t.Error != nil {
					panic(t.Error)
				}

				words := tokenizer.Tokenize(t.Text)
				scores, likely, strict := classifier.LogScores(
					words,
				)

				sentiment := func(min, max float64) string {
					if min < max {
						return "negative"
					} else if min > max {
						return "positive"
					} else {
						return "neutral"
					}
				}

				data = append(data, dataset{
					tweet:     t.HTML,
					tokenizer: words,
					minScore:  scores[0],
					maxScore:  scores[1],
					sentiment: sentiment(scores[0], scores[1]),
					likely:    likely,
					strict:    strict,
					timestamp: t.Timestamp,
				})

				fmt.Println(data)
				time.Sleep(10 * time.Second)
			},
			QueueDone: func() {
				fmt.Println("done")
			},
		},
	})

	for tweet := range tweets {
		t := tweet
		task := func() interface{} {
			return t
		}

		q.Append(task)
	}

	q.Start()
}
