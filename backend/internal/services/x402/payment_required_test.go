package x402

import "testing"

func TestDecodePaymentRequired(t *testing.T) {
	// Sample from GoPlausible testnet /avm/weather (truncated accepts to algorand only).
	header := "eyJ4NDAyVmVyc2lvbiI6MiwiZXJyb3IiOiJQYXltZW50IHJlcXVpcmVkIiwicmVzb3VyY2UiOnsidXJsIjoiaHR0cHM6Ly9leGFtcGxlLng0MDIuZ29wbGF1c2libGUueHl6L2F2bS93ZWF0aGVyIiwiZGVzY3JpcHRpb24iOiJXZWF0aGVyIGRhdGEiLCJtaW1lVHlwZSI6ImFwcGxpY2F0aW9uL2pzb24ifSwiYWNjZXB0cyI6W3sic2NoZW1lIjoiZXhhY3QiLCJuZXR3b3JrIjoiYWxnb3JhbmQ6U0dPMUdLU3p5RTdJRVBJdFR4Q0J5dzl4OEZtbnJDRGV4aTkvY09VSk9pST0iLCJhbW91bnQiOiIxMDAwIiwiYXNzZXQiOiIxMDQ1ODk0MSIsInBheVRvIjoiTVBZNTRDTFBIMk9LRUdDNlM1TjJMREFGRE5PNUJWTlY1MzJOQlo1VkQ2R09ORDNTVFBOWFpZWE9GRSIsIm1heFRpbWVvdXRTZWNvbmRzIjozMDAsImV4dHJhIjp7Im5hbWUiOiJVU0RDIiwiZGVjaW1hbHMiOjYsImZlZVBheWVyIjoiWk1GSzJPSTdaQkQyVTI3SVNFUlpDNFM2TEtNNldNRkpQWlE0TVlOSkRaMlZOQk5NQkE2N1JBMjJBQSJ9fV19"

	req, err := decodePaymentRequired(header)
	if err != nil {
		t.Fatalf("decodePaymentRequired: %v", err)
	}
	if req.X402Version != 2 {
		t.Fatalf("version: got %d", req.X402Version)
	}

	accept, err := selectAlgorandAccept(req.Accepts, "algorand:SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI=")
	if err != nil {
		t.Fatalf("selectAlgorandAccept: %v", err)
	}
	if accept.Amount != "1000" {
		t.Fatalf("amount: got %q", accept.Amount)
	}
	if accept.Asset != "10458941" {
		t.Fatalf("asset: got %q", accept.Asset)
	}
	if feePayerFromExtra(accept.Extra) == "" {
		t.Fatal("expected feePayer in extra")
	}
}
