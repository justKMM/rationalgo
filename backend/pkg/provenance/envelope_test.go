package provenance

import (
	"testing"
	"time"
)

func TestEnvelopeRoundTrip(t *testing.T) {
	in := &Envelope{
		Version:      1,
		AgentID:      "drone-ops-01",
		SessionID:    "sess-frankfurt-001",
		TaskHash:     "abc123",
		DecisionHash: "def456",
		Vendor:       "goplausible-weather",
		AmountEURQ:   0.001,
		Intent:       "Should drone deliveries operate in Frankfurt in the next 2 hours?",
		Expected:     "91% forecast accuracy",
		Confidence:   0.87,
		CommittedAt:  time.Now().Unix(),
	}

	note, err := Encode(in)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	if !IsProvenance(note) {
		t.Fatal("expected IsProvenance true")
	}

	out, err := Decode(note)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if out.Vendor != in.Vendor || out.DecisionHash != in.DecisionHash {
		t.Fatalf("round-trip mismatch: %+v vs %+v", out, in)
	}
}

func TestOutcomeRoundTrip(t *testing.T) {
	in := &OutcomeEnvelope{
		Version:     1,
		OriginalTx:  "TXID123",
		Actual:      "12% precip",
		Score:       0.93,
		GroundTruth: "OpenMeteo historical",
		ComputedAt:  time.Now().Unix(),
	}

	note, err := EncodeOutcome(in)
	if err != nil {
		t.Fatalf("EncodeOutcome: %v", err)
	}
	if !IsOutcome(note) {
		t.Fatal("expected IsOutcome true")
	}

	out, err := DecodeOutcome(note)
	if err != nil {
		t.Fatalf("DecodeOutcome: %v", err)
	}
	if out.OriginalTx != in.OriginalTx || out.Score != in.Score {
		t.Fatalf("round-trip mismatch: %+v vs %+v", out, in)
	}
}

func TestDecodeMalformed(t *testing.T) {
	if _, err := Decode("not-a-note"); err == nil {
		t.Fatal("expected error for non-RAv1 note")
	}
	if _, err := Decode("RAv1:!!!"); err == nil {
		t.Fatal("expected error for invalid base64")
	}
	if _, err := DecodeOutcome("RAv1out:!!!"); err == nil {
		t.Fatal("expected error for invalid outcome base64")
	}
}
