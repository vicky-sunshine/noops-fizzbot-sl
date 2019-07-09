package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var baseURL = "https://api.noopschallenge.com"
var savefile = "./save/save.yaml"
var dbdirpath = "./stages"

type Save struct {
	LastStageHash string `yaml:"last_stage_hash"`
}

func GetLastStage(savefile string) (*Save, error) {
	f, err := os.Open(savefile)
	if err != nil {
		return nil, err
	}

	var s Save
	err = yaml.NewDecoder(f).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type Stage struct {
	Hash          string           `yaml:"hash"`
	Question      QuestionResponse `yaml:"question"`
	Answer        string           `yaml:"answer"`
	NextStageHash string           `yaml:"next_stage_hash"`
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "users", Description: "Store the username and age"},
		{Text: "articles", Description: "Store the article text posted by user"},
		{Text: "comments", Description: "Store the text commented to articles"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func main() {
	// fmt.Println("Please select table.")
	// t := prompt.Input("> ", completer)
	// fmt.Println("You selected " + t)

	// savefile := flag.String("save", "", "savefile path")
	// dbpath := flag.String("save", "", "Q&A store db path")
	// if *savefile == "" {
	// 	log.Fatal("savefile path is empty")
	// }
	// if *dbpath == "" {
	// 	log.Fatal("Q&A store db path is empty")
	// }

	var rootCmd = &cobra.Command{Use: "fbsl"}

	var cmdLastQuestion = &cobra.Command{
		Use: "last",
		// Short: "Read hacker news",
		Run: func(cmd *cobra.Command, args []string) {
			s, err := GetLastStage(savefile)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Last Stage: %v\n\n", s.LastStageHash)

			qr, err := GetQuestion(s.LastStageHash)
			if err != nil {
				panic(err)
			}
			fmt.Printf("Message:\n%v\n", qr.Message)
			fmt.Printf("Rules:\n")
			for _, r := range qr.Rules {
				fmt.Printf("%v", r)
			}
			fmt.Printf("\nNumbers: %v\n", qr.Numbers)
			fmt.Printf("\nExampleResponse: %v\n", qr.ExampleResponse)
		},
	}

	var cmdGetQuestion = &cobra.Command{
		Use:   "get [hash of question]",
		Short: "get question",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			qr, err := GetQuestion(args[0])
			if err != nil {
				panic(err)
			}
			fmt.Printf("Message:\n%v\n", qr.Message)
			fmt.Printf("Rules:\n")
			for _, r := range qr.Rules {
				fmt.Printf("%v", r)
			}
			fmt.Printf("\nNumbers: %v\n", qr.Numbers)
			fmt.Printf("\nExampleResponse: %v\n", qr.ExampleResponse)
		},
	}

	var cmdPostAnswer = &cobra.Command{
		Use:   "ans [hash of question] [answer]",
		Short: "answer question",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			ar, err := PostAnswer(args[0], args[1])
			if err != nil {
				panic(err)
			}
			fmt.Printf("\nResult: %v\n", ar.Result)
			fmt.Printf("Message:\n%v\n", ar.Message)
			fmt.Printf("\nNextQuestion: %v\n", ar.NextQuestion)
		},
	}

	rootCmd.AddCommand(cmdLastQuestion, cmdGetQuestion, cmdPostAnswer)
	rootCmd.Execute()

	// _, err := GetQuestion("sQ4B9Ei5X0CFK7whpHhXImrXrp7tKEyT-n5La-Yi65A")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// _, err = PostAnswer(
	// 	"sQ4B9Ei5X0CFK7whpHhXImrXrp7tKEyT-n5La-Yi65A",
	// 	"1 2 Fizz 4 Buzz Fizz 7 8 Fizz Buzz 11 Fizz 13 14 FizzBuzz")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// artifacts for fuzzbuzz

type Rule struct {
	Number   int    `json:"number"`
	Response string `json:"response"`
}

func (rl Rule) String() string {
	return fmt.Sprintf("Number: %v\tResponse: %v\n", rl.Number, rl.Response)
}

type Anwser struct {
	Answer string `json:"answer"`
}

type QuestionResponse struct {
	Message         string `json:"message" yaml:"message"`
	Rules           []Rule `json:"rules" yaml:"rules"`
	Numbers         []int  `json:"numbers" yaml:"numbers"`
	ExampleResponse Anwser `json:"exampleResponse,omitempty" yaml:"example_response"`
}

type AnwserResponse struct {
	NextQuestion string `json:"nextQuestion,omitempty"`
	Message      string `json:"message"`
	Result       string `json:"result"`
}

func GetQuestion(hash string) (*QuestionResponse, error) {
	// prepare get request
	targetURL := fmt.Sprintf("%v%v%v", baseURL, "/fizzbot/questions/", hash)
	rsp, err := http.Get(targetURL)
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Hash Not Found: %v", hash)
	}

	// parse question
	var qr QuestionResponse
	err = json.NewDecoder(rsp.Body).Decode(&qr)
	if err != nil {
		return nil, err
	}

	// fmt.Println(rsp.StatusCode)
	// fmt.Println(qr.Message)
	// fmt.Println(qr.Rules)
	// fmt.Println(qr.Numbers)
	// fmt.Println(qr.ExampleResponse)
	return &qr, nil
}

func PostAnswer(hash string, answer string) (*AnwserResponse, error) {
	// prepare post request
	ans := Anwser{Answer: answer}
	b, err := json.Marshal(ans)
	if err != nil {
		return nil, err
	}

	targetURL := fmt.Sprintf("%v%v%v", baseURL, "/fizzbot/questions/", hash)
	rsp, err := http.Post(targetURL, "application/json", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	if rsp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Hash Not Found: %v", hash)
	}

	var ar AnwserResponse
	err = json.NewDecoder(rsp.Body).Decode(&ar)
	if err != nil {
		return nil, err
	}

	// store answer
	if rsp.StatusCode == http.StatusOK {

		qr, err := GetQuestion(hash)
		if err != nil {
			return nil, err
		}
		nextHash := strings.Split(ar.NextQuestion, "/")[3]

		stg := Stage{}
		stg.Hash = hash
		stg.Answer = answer
		stg.Question = *qr
		stg.NextStageHash = nextHash

		stgpath := fmt.Sprintf("%v/%v.yaml", dbdirpath, hash)
		stgf, err := os.Create(stgpath)
		if err != nil {
			return nil, err
		}
		err = yaml.NewEncoder(stgf).Encode(stg)
		if err != nil {
			return nil, err
		}

		save := Save{}
		save.LastStageHash = nextHash
		sf, err := os.Create(savefile)
		if err != nil {
			return nil, err
		}
		err = yaml.NewEncoder(sf).Encode(save)
		if err != nil {
			return nil, err
		}
	}
	return &ar, nil
}
