package ledger

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mini-ledger/src/models"
	"mini-ledger/src/utils"
	"time"
)

func HashTransaction(transaction models.Transaction) string {
	data, _ := json.Marshal(transaction)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func MaybeProduceDigest(db *sql.DB) {
	rows, err := db.Query("SELECT hash FROM transaction_hashes ORDER BY id DESC")
	if err != nil {
		fmt.Println("Error getting hashes:", err)
		return
	}
	if rows == nil {
		return
	}
	defer rows.Close()

	var hashes []string
	for rows.Next() {
		var h string
		rows.Scan(&h)
		hashes = append(hashes, h)
	}

	tree := utils.NewMerkleTree(hashes)
	root := tree.GetRootHash()

	var prev string
	db.QueryRow("SELECT root_hash FROM digests ORDER BY id DESC LIMIT 1").Scan(&prev)

	_, err = db.Exec("INSERT INTO digests(root_hash, prev_root) VALUES(?, ?)", root, prev)
	if err != nil {
		fmt.Println("Error inserting digest:", err)
	}
}

func VerifyDigests(db *sql.DB) bool {
	rows, err := db.Query("SELECT hash FROM transaction_hashes ORDER BY id DESC")
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

func StartWorker(db *sql.DB, batchSize int) {
	for {
		var cursor int64
		db.QueryRow("SELECT COALESCE(MAX(end_id), 0) FROM digests").Scan(&cursor)
		var pending int
		db.QueryRow("SELECT COUNT(*) FROM transaction_hashes WHERE id > ?", cursor).Scan(&pending)

		if pending < batchSize {
			time.Sleep(2 * time.Second)
			continue
		}

		q := "SELECT id, hash FROM transaction_hashes WHERE id > ? ORDER BY id ASC LIMIT ?"
		rows, err := db.Query(q, cursor, batchSize)
		if err != nil {
			fmt.Println("query batch error", err)
			time.Sleep(time.Second)
			continue
		}
		var ids []int64
		var hashes []string
		for rows.Next() {
			var id int64
			var hash string
			if err := rows.Scan(&id, &hash); err != nil {
				fmt.Println("row scan error", err)
				break
			}
			ids = append(ids, id)
			hashes = append(hashes, hash)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			fmt.Println("rows error:", err)
			time.Sleep(time.Second)
			continue
		}

		if len(hashes) == 0 {
			continue
		}

		startID := ids[0]
		endID := ids[len(ids)-1]
		tree := utils.NewMerkleTree(hashes)
		root := tree.GetRootHash()

		var prev sql.NullString
		db.QueryRow("SELECT root_hash FROM digests ORDER BY id DESC LIMIT 1").Scan(&prev)
		_, err = db.Exec("INSERT INTO digests(root_hash, prev_root, start_id, end_id) VALUES(?, ?, ?, ?)",
			root, prev.String, startID, endID)

		if err != nil {
			fmt.Println("Error inserting digest:", err)
			time.Sleep(time.Second)
			continue
		}

	}
}
