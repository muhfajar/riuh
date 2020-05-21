package handler

import (
	"encoding/json"
	"fmt"
	"github.com/RadhiFadlillah/go-sastrawi"
	"github.com/jbrukh/bayesian"
	"github.com/muhfajar/riuh/api/lexicon"
	"log"
	"net/http"
)

func Handler(w http.ResponseWriter, _ *http.Request) {
	tokenizer := sastrawi.NewTokenizer()
	classifier := bayesian.NewClassifierTfIdf(Good, Bad)
	classifier.Learn(lexicon.Positive(), Good)
	classifier.Learn(lexicon.Negative(), Bad)
	classifier.ConvertTermsFreqToTfIdf()

	d := &data{
		tokenizer:  tokenizer,
		classifier: classifier,
	}
	d.worker()

	resp, err := json.Marshal(d)
	if err != nil {
		log.Println(err)
	}

	_, _ = fmt.Fprintf(w, string(resp))
}
