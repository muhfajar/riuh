package api

import (
	"github.com/RadhiFadlillah/go-sastrawi"
	"github.com/gin-gonic/gin"
	"github.com/jbrukh/bayesian"
	"github.com/muhfajar/queue"
	twt "github.com/n0madic/twitter-scraper"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

const (
	Good bayesian.Class = "Good"
	Bad  bayesian.Class = "Bad"
)

type dataset struct {
	Tweet     string   `json:"tweet"`
	Tokenizer []string `json:"tokenizer"`
	Cluster   string   `json:"cluster"`
	MinScore  float64  `json:"min_score"`
	MaxScore  float64  `json:"max_score"`
	Sentiment string   `json:"sentiment"`
	Likely    int      `json:"likely"`
	Strict    bool     `json:"strict"`
	Timestamp int64    `json:"timestamp"`
}

type data struct {
	Data            []dataset `json:"data"`
	TweetCount      int       `json:"tweet_count"`
	BlankTokenCount int       `json:"blank_token_count"`
}

func Routes() *gin.Engine {
	r := gin.Default()

	r.GET("/api", func(c *gin.Context) {
		d := &data{}
		d.worker()
		c.JSON(http.StatusOK, d)
	})

	return r
}

func (d *data) worker() {
	trends, err := twt.GetTrends()
	if err != nil {
		panic(err)
	}

	q := queue.NewQueue(&queue.Worker{
		Thread: 10,
		Alloc:  len(trends),
		Set: queue.Callback{
			TaskDone: func(result interface{}) {
				trend := result.(string)
				d.sentiment(trend)
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

func (d *data) sentiment(trend string) {
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

	query := trend
	query += "-filter:retweets"

	tweets := twt.SearchTweets(query, 5000)

	q := queue.NewQueue(&queue.Worker{
		Thread: 250,
		Alloc:  len(tweets),
		Set: queue.Callback{
			TaskDone: func(result interface{}) {
				d.TweetCount++
				t := result.(*twt.Result)
				if t.Error != nil {
					panic(t.Error)
				}

				words := tokenizer.Tokenize(t.Text)
				scores, likely, strict := classifier.LogScores(
					words,
				)

				if len(words) == 0 {
					d.BlankTokenCount++
				}

				sentiment := func(min, max float64) string {
					if min < max {
						return "negative"
					} else if min > max {
						return "positive"
					} else {
						return "neutral"
					}
				}

				d.Data = append(d.Data, dataset{
					Tweet:     t.HTML,
					Tokenizer: words,
					Cluster:   trend,
					MinScore:  scores[0],
					MaxScore:  scores[1],
					Sentiment: sentiment(scores[0], scores[1]),
					Likely:    likely,
					Strict:    strict,
					Timestamp: t.Timestamp,
				})
			},
			QueueDone: func() {
				log.Print(trend)
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
