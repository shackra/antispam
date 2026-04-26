package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

func runEmacs(eval string) (string, error) {
	quizEl, exists := os.LookupEnv("EMACS_QUIZ_EL")
	if !exists {
		// estaremos en desarrollo?
		quizEl = "./lisp/emacs-quiz.el"
	}
	cmd := exec.Command("emacs", "-q", "--batch", "-l", quizEl, "--eval", eval)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return strings.TrimSpace(out.String()), err
}

func obtenerRetoBinding() Challenge {
	out, err := runEmacs(`(princ (format "%S" (zr-obtener-binding-azar)))`)
	if err != nil {
		log.Fatalf("hubo un error al intentar obtener el reto, error: %v", err)
	}

	out = strings.Trim(out, "()\"")

	parts := strings.SplitN(out, " . ", 2)
	key := strings.Trim(parts[0], "\"")
	cmd := strings.Trim(parts[1], "\"")

	return Challenge{
		Question: fmt.Sprintf("En Emacs ¿A qué comando corresponde el atajo `%s`?", key),
		Answer:   cmd,
		Key:      key,
	}
}
