package main

import (
	"fmt"
	"time"

	"rationalgo/pkg/provenance"
)

func main() {
	env := &provenance.Envelope{
		Version:      1,
		AgentID:      "drone-ops-01",
		SessionID:    "example-session",
		TaskHash:     "sha256-of-task-intent",
		DecisionHash: "sha256-of-decision-record",
		Vendor:       "goplausible-weather",
		AmountEURQ:   0.001,
		Intent:       "Should drone deliveries operate in Frankfurt in the next 2 hours?",
		Expected:     "91% forecast accuracy",
		Confidence:   0.87,
		CommittedAt:  time.Now().Unix(),
	}

	note, err := provenance.Encode(env)
	if err != nil {
		panic(err)
	}

	fmt.Println("RAv1 provenance note (commit before x402 payment):")
	fmt.Println(note)
}
