package main

import (
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
)

func main() {
	fmt.Println("init")

	err := sentry.Init(sentry.ClientOptions{})

	if err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}

	_, err = os.Open("INVALID")
	if err != nil {
		fmt.Println("Capturing...")
		eventID := sentry.CaptureException(err)
		fmt.Println(*eventID)
	}

	sentry.Flush(time.Second * 5)
}
