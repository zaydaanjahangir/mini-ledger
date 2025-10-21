package ledger

import (
	"mini-ledger/src/models"
	"database/sql"
	"encoding/json"
	"encoding/hex"
	"crypto/sha256"
	"mini-ledger/src/utils"
	"fmt"
)

func HashTransaction(transaction models.Transaction) string{
	data, _ := json.Marshal(transaction)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func MaybeProduceDigest(db *sql.DB, batchSize int) {
	rows, _ := db.Query("SELECT hash FROM transaction_hashes ORDER BY id DESC LIMIT ?", batchSize)
    defer rows.Close()

	var hashes []string
	for rows.Next() {
		var h string
		rows.Scan(&h)
		hashes = append(hashes, h)
	}

	if len(hashes) < batchSize {
		return
	}

	tree := utils.NewMerkleTree(hashes)
    root := tree.GetRootHash()

	var prev string
    db.QueryRow("SELECT root_hash FROM digests ORDER BY id DESC LIMIT 1").Scan(&prev)

	_, err := db.Exec("INSERT INTO digests(root_hash, prev_root) VALUES(?, ?)", root, prev)
    if err != nil {
        fmt.Println("Error inserting digest:", err)
    }
}

func VerifyDigests(db *sql.DB) bool {
    rows, err := db.Query("SELECT hash FROM transaction_hashes ORDER BY id ASC")
	if err != nil {
    	fmt.Println("Error reading hashes:", err)
    return false
	}
    defer rows.Close()

    var hashes []string
    for rows.Next() {
        var h string
        rows.Scan(&h)
        hashes = append(hashes, h)
    }

    tree := utils.NewMerkleTree(hashes)
    computedRoot := tree.GetRootHash()

    var storedRoot string
    db.QueryRow("SELECT root_hash FROM digests ORDER BY id DESC LIMIT 1").Scan(&storedRoot)

    return computedRoot == storedRoot
}


