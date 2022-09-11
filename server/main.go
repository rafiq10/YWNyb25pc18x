package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	f, err := os.OpenFile("./logs/log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	l := zerolog.New(f).With().Timestamp().Logger()

	defer func() {
		if r := recover(); r != nil {
			l.Error().Msgf("Recovered in f", r)
		}
	}()

	s := &http.Server{
		Addr:         ":8080",
		Handler:      http.FileServer(http.Dir("./doc")),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	gracefulShutdown(s, &l)
}

func gracefulShutdown(s *http.Server, l *zerolog.Logger) {
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			l.Fatal().Msg(err.Error())
		}
	}()

	sigChan := make(chan os.Signal, 5)
	signal.Notify(sigChan, os.Interrupt)
	signal.Notify(sigChan, syscall.SIGTERM)
	sig := <-sigChan
	l.Fatal().Msgf("Received terminate shutdown", sig)

	tc, cncl := context.WithTimeout(context.Background(), 30*time.Second)
	cncl()
	err := s.Shutdown(tc)
	if err != nil {
		log.Println(err)
	}
}
