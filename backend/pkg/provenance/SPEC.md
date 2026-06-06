# RAv1 — Agentic x402 Provenance Standard

RAv1 defines how AI agents attach structured, verifiable metadata to Algorand
transactions around x402 (HTTP 402) payments.

## Note field formats

| Phase | Prefix | When |
|-------|--------|------|
| Pre-payment commit | `RAv1:` | Before `X-PAYMENT` is sent |
| Post-outcome commit | `RAv1out:` | After response is evaluated |

Value after prefix: **base64url** encoding of canonical JSON (UTF-8, `json.Marshal`).

Example:

```
RAv1:eyJ2IjoxLCJhZ2VudF9pZCI6ImRyb25lLW9wcy0wMSIs...
```

## Envelope fields (pre-payment)

| Field | Type | Semantics |
|-------|------|-----------|
| `v` | int | Schema version (currently `1`) |
| `agent_id` | string | Stable agent identifier |
| `session_id` | string | Groups related decisions in one run |
| `task_hash` | string | SHA-256 hex of raw task intent string |
| `decision_hash` | string | SHA-256 hex of DecisionRecord JSON |
| `vendor` | string | Vendor ID chosen for payment |
| `amount_eurq` | float | Quoted price in EURQ |
| `intent` | string | One-sentence human-readable task |
| `expected` | string | Predicted value (e.g. forecast accuracy) |
| `confidence` | float | Agent confidence 0.0–1.0 |
| `committed_at` | int64 | Unix seconds before payment |

## Outcome envelope fields (post-payment)

| Field | Type | Semantics |
|-------|------|-----------|
| `v` | int | Schema version |
| `original_tx` | string | Tx ID of the pre-payment RAv1 commit |
| `actual` | string | Observed outcome |
| `score` | float | Match score 0.0–1.0 |
| `ground_truth` | string | Verification source name |
| `computed_at` | int64 | Unix seconds when outcome was computed |

## Algorand transaction pattern

Both envelopes are written as **0-ALGO self-payments** with the encoded string
in the transaction `note` field (max 1000 bytes). No smart contract required.

## Querying the audit trail

Use the Algorand Indexer with a note prefix filter:

```
note-prefix=RAv1:
note-prefix=RAv1out:
```

Anyone with indexer access can reconstruct agent reasoning and outcomes without
the application database.

## Implementation in other languages

1. JSON-marshal the envelope struct (snake_case field names).
2. Base64url-encode the JSON bytes (no padding required; Go uses StdEncoding which may pad — interoperable decoders should accept padding).
3. Prefix with `RAv1:` or `RAv1out:`.
4. Submit as transaction note before/after payment.

Go reference: `rationalgo/pkg/provenance`.
