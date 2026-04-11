package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadQuoteFileSupportsWrappedQuotePayload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "quote.json")
	data := `{
  "stage": "quote",
  "status": "ok",
  "quote": {
    "type": "lifi",
    "id": "quote-id",
    "action": {
      "fromAmount": "1000",
      "fromAddress": "0x1111111111111111111111111111111111111111"
    },
    "estimate": {
      "approvalAddress": "0x2222222222222222222222222222222222222222"
    }
  }
}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write quote file: %v", err)
	}

	quote, err := readQuoteFile(path)
	if err != nil {
		t.Fatalf("readQuoteFile returned error: %v", err)
	}
	if quote == nil || quote.ID != "quote-id" {
		t.Fatalf("unexpected quote: %#v", quote)
	}
	if quote.Estimate.ApprovalAddress != "0x2222222222222222222222222222222222222222" {
		t.Fatalf("unexpected approval address: %#v", quote)
	}
}

func TestCompletionSpecsIncludeLatestExecutionFlags(t *testing.T) {
	t.Parallel()

	specs := completionSpecs()
	depositFlags := strings.Join(specs["deposit"], " ")
	quoteFlags := strings.Join(specs["quote"], " ")
	approveFlags := strings.Join(specs["approve"], " ")

	for _, flag := range []string{"--approval-amount", "--gas-policy", "--wait-timeout", "--portfolio-timeout", "--simulate", "--skip-simulate"} {
		if !strings.Contains(depositFlags, flag) {
			t.Fatalf("deposit completion missing %s", flag)
		}
	}
	if !strings.Contains(quoteFlags, "--unsigned") {
		t.Fatalf("quote completion missing --unsigned")
	}
	if !strings.Contains(approveFlags, "--gas-policy") {
		t.Fatalf("approve completion missing --gas-policy")
	}
}
