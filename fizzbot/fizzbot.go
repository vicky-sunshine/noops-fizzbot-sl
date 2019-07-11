package fizzbot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var baseURL = "https://api.noopschallenge.com"

type Rule struct {
	Number   int    `json:"number"`
	Response string `json:"response"`
}

func (rl Rule) String() string {
	return fmt.Sprintf("Number: %v\tResponse: %v\n", rl.Number, rl.Response)
}

type QuestionResponse struct {
	Message         string `json:"message" yaml:"message"`
	Rules           []Rule `json:"rules" yaml:"rules"`
	Numbers         []int  `json:"numbers" yaml:"numbers"`
	ExampleResponse Anwser `json:"exampleResponse,omitempty" yaml:"example_response"`
}

type Anwser struct {
	Answer string `json:"answer"`
}

type AnwserResponse struct {
	NextQuestion   string `json:"nextQuestion,omitempty"`
	Message        string `json:"message"`
	Result         string `json:"result"`
	Grade          string `json:"grade,omitempty"`
	ElapsedSeconds int    `json:"elapsedSeconds,omitempty"`
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

	// parse answer result
	var ar AnwserResponse
	err = json.NewDecoder(rsp.Body).Decode(&ar)
	if err != nil {
		return nil, err
	}

	return &ar, nil
}

func FizzBuzz(num int, rules []Rule) string {
	result := ""

	for _, v := range rules {
		if num%v.Number == 0 {
			result += v.Response
		}
	}

	if result == "" {
		result = fmt.Sprintf("%d", num)

	}
	return result
}

func Solve(numbers []int, rules []Rule) string {
	if len(numbers) == 0 {
		// starter question
		return "go"
	}
	result := make([]string, len(numbers))
	for i, v := range numbers {
		result[i] = FizzBuzz(v, rules)
	}
	return strings.Join(result, " ")
}
