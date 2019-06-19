package main

import "math/rand"

func generateKek() string {
	phrases := []string{
		"Кек",
		"кпек",
		"пук",
		"КЕКУС",
		"КЕК",
	}
	phrase := phrases[rand.Intn(len(phrases))]
	return phrase
}
