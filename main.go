package main

import (
	"fizzbotplay/fizzbot"
	"flag"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

type SavePoint struct {
	LastStageHash string `yaml:"last_stage_hash"`
}

type Stage struct {
	Hash          string                   `yaml:"hash"`
	Question      fizzbot.QuestionResponse `yaml:"question"`
	Answer        string                   `yaml:"answer"`
	NextStageHash string                   `yaml:"next_stage_hash"`
}

func LoadSavePoint(savefile string) (*SavePoint, error) {
	f, err := os.Open(savefile)
	if err != nil {
		return nil, err
	}

	var s SavePoint
	err = yaml.NewDecoder(f).Decode(&s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func WriteSavePoint(savefile string, hash string) error {
	save := SavePoint{}
	save.LastStageHash = hash
	sf, err := os.Create(savefile)
	if err != nil {
		return err
	}
	err = yaml.NewEncoder(sf).Encode(save)
	if err != nil {
		return err
	}
	return nil
}

func SaveStage(hash string, answer string, ar fizzbot.AnwserResponse, dbdirpath string, savefile string) error {
	qr, err := fizzbot.GetQuestion(hash)
	if err != nil {
		return err
	}

	nextHash := ""
	if ar.NextQuestion != "" {
		nextHash = strings.Split(ar.NextQuestion, "/")[3]
	}

	stg := Stage{}
	stg.Hash = hash
	stg.Answer = answer
	stg.Question = *qr
	stg.NextStageHash = nextHash

	stgpath := fmt.Sprintf("%v/%v.yaml", dbdirpath, hash)
	stgf, err := os.Create(stgpath)
	if err != nil {
		return err
	}
	err = yaml.NewEncoder(stgf).Encode(stg)
	if err != nil {
		return err
	}
	err = WriteSavePoint(savefile, nextHash)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	savepointdir := flag.String("savepoint", "", "savepoint path")
	flag.Parse()
	if *savepointdir == "" {
		panic("savepoint path is empty")
	}

	dbdirpath := *savepointdir
	savefile := *savepointdir + "/save.yaml"
	firstHash := "1"
	firstAnswer := "go"

	// init
	if _, err := os.Stat(dbdirpath); os.IsNotExist(err) {
		os.Mkdir(dbdirpath, 0755)
	}
	err := WriteSavePoint(savefile, firstHash)
	if err != nil {
		panic(err)
	}

	count := 1
	hash := firstHash
	answer := firstAnswer

	for {
		// Get Save Point
		fmt.Printf("=====%v=====\n", count)
		s, err := LoadSavePoint(savefile)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Last Stage: %v\n\n", s.LastStageHash)

		// Get Question
		hash = s.LastStageHash
		qr, err := fizzbot.GetQuestion(s.LastStageHash)
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

		// Find Answer
		if count > 1 {
			answer = fizzbot.Solve(qr.Numbers, qr.Rules)
		}
		fmt.Printf("\nWhat we Answer: %v\n", answer)

		// Post Answer
		ar, err := fizzbot.PostAnswer(hash, answer)
		if err != nil {
			panic(err)
		}
		fmt.Printf("\nResult: %v\n", ar.Result)
		fmt.Printf("Message:\n%v\n", ar.Message)

		// Save Stage
		if ar.Result == "correct" {
			err = SaveStage(hash, answer, *ar, dbdirpath, savefile)
			if err != nil {
				panic(err)
			}
			fmt.Printf("\nNextQuestion: %v\n\n", ar.NextQuestion)
		} else if ar.Result == "interview complete" {
			fmt.Printf("\nGrade: %v", ar.Grade)
			fmt.Printf("\nElapsed Seconds: %v\n", ar.ElapsedSeconds)
			err = SaveStage(hash, answer, *ar, dbdirpath, savefile)
			if err != nil {
				panic(err)
			}
			break
		} else {
			s := fmt.Sprintf("Question:\n%#v\n\nAnswer:\n%#v\n\n", qr, ar)
			panic(s)
		}
		count++
	}
}
