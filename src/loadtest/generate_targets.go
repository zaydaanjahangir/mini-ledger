package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
)

type vegetaTarget struct {
	Method string              `json:"method"`
	URL    string              `json:"url"`
	Header map[string][]string `json:"header,omitempty"`
	Body   []byte              `json:"body,omitempty"`
}

type accountPayload struct {
	Name string `json:"name"`
}

type txEntry struct {
	Account string `json:"account"`
	Amount  int64  `json:"amount"`
}

type txPayload struct {
	Description string    `json:"description"`
	Entries     []txEntry `json:"entries"`
}

func main() {
	var (
		baseURL     = flag.String("base-url", "http://localhost:8080", "API base URL")
		accountCnt  = flag.Int("accounts", 200, "number of accounts to generate")
		txCnt       = flag.Int("transactions", 10000, "number of transaction targets to generate")
		outputDir   = flag.String("out", ".", "output directory")
		randomSeed  = flag.Int64("seed", 42, "seed for deterministic transaction generation")
	)
	flag.Parse()

	if *accountCnt < 2 {
		log.Fatal("accounts must be at least 2")
	}
	if *txCnt < 1 {
		log.Fatal("transactions must be at least 1")
	}

	if err := os.MkdirAll(*outputDir, 0o755); err != nil {
		log.Fatalf("creating output directory: %v", err)
	}

	accountFile := filepath.Join(*outputDir, "accounts-targets.json")
	txFile := filepath.Join(*outputDir, "transactions-targets.json")

	accounts := buildAccountsTargets(*baseURL, *accountCnt)
	if err := writeNDJSON(accountFile, accounts); err != nil {
		log.Fatalf("writing %s: %v", accountFile, err)
	}

	transactions := buildTransactionTargets(*baseURL, *accountCnt, *txCnt, *randomSeed)
	if err := writeNDJSON(txFile, transactions); err != nil {
		log.Fatalf("writing %s: %v", txFile, err)
	}

	fmt.Printf("generated %d account targets in %s\n", len(accounts), accountFile)
	fmt.Printf("generated %d transaction targets in %s\n", len(transactions), txFile)
}

func buildAccountsTargets(baseURL string, accountCnt int) []vegetaTarget {
	targets := make([]vegetaTarget, 0, accountCnt)
	for i := 1; i <= accountCnt; i++ {
		payload := accountPayload{Name: fmt.Sprintf("account-%03d", i)}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("marshal account payload: %v", err)
		}
		targets = append(targets, vegetaTarget{
			Method: "POST",
			URL:    fmt.Sprintf("%s/accounts", baseURL),
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: body,
		})
	}
	return targets
}

func buildTransactionTargets(baseURL string, accountCnt, txCnt int, seed int64) []vegetaTarget {
	rnd := rand.New(rand.NewSource(seed))
	targets := make([]vegetaTarget, 0, txCnt)

	for i := 1; i <= txCnt; i++ {
		from := rnd.Intn(accountCnt) + 1
		to := rnd.Intn(accountCnt-1) + 1
		if to >= from {
			to++
		}
		amount := int64(rnd.Intn(9900) + 100)

		payload := txPayload{
			Description: fmt.Sprintf("load-tx-%06d", i),
			Entries: []txEntry{
				{Account: fmt.Sprintf("account-%03d", from), Amount: -amount},
				{Account: fmt.Sprintf("account-%03d", to), Amount: amount},
			},
		}
		body, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("marshal transaction payload: %v", err)
		}
		targets = append(targets, vegetaTarget{
			Method: "POST",
			URL:    fmt.Sprintf("%s/transactions", baseURL),
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
			Body: body,
		})
	}
	return targets
}

func writeNDJSON(path string, targets []vegetaTarget) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	for _, t := range targets {
		if err := enc.Encode(t); err != nil {
			return err
		}
	}
	return nil
}
