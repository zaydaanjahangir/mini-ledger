# Load Test Files

This directory contains generated Vegeta targets for stress testing:

- `accounts-targets.json`: creates `account-001` through `account-200`
- `transactions-targets.json`: 10,000 valid double-entry transactions between those accounts

Re-generate targets (optional):

```bash
cd src/loadtest
go run ./generate_targets.go -accounts 200 -transactions 10000 -out .
```

Run with Vegeta JSON target format:

```bash
vegeta attack -format=json -targets=src/loadtest/accounts-targets.json -rate=20 -duration=10s | tee accounts-results.bin | vegeta report
vegeta attack -format=json -targets=src/loadtest/transactions-targets.json -rate=300 -duration=2m | tee tx-results.bin | vegeta report
```
#### Benchmark Script
```bash
rates=(5 10 20 30 40 50 75 100 150 200 250 300)
printf "%-6s %-9s %-9s %-10s %-10s %-10s\n" "rate" "success%" "fail%" "thpt(rps)" "p95(ms)" "p99(ms)"

for r in "${rates[@]}"; do
  vegeta attack -format=json -targets=transactions-targets.json -rate="${r}" -duration=30s \
    | tee "tx-${r}.bin" >/dev/null

  vegeta report -type=json < "tx-${r}.bin" \
    | python3 -c 'import json,sys; d=json.load(sys.stdin); print("{:.0f} {:.2f} {:.2f} {:.2f} {:.2f} {:.2f}".format(d["rate"], d["success"]*100, (1-d["success"])*100, d["throughput"], d["latencies"]["95th"]/1e6, d["latencies"]["99th"]/1e6))' \
    | awk '{printf "%-6s %-9s %-9s %-10s %-10s %-10s\n",$1,$2,$3,$4,$5,$6}'
done
```

## Initial Results

| Rate (rps) | Success % | Fail % | Throughput (rps) | P95 (ms) | P99 (ms) |
|---:|---:|---:|---:|---:|---:|
| 5   | 100.00 | 0.00  | 5.03   | 6.16   | 6.96   |
| 10  | 100.00 | 0.00  | 10.03  | 5.95   | 6.87   |
| 20  | 100.00 | 0.00  | 20.03  | 7.70   | 12.51  |
| 30  | 100.00 | 0.00  | 30.03  | 8.05   | 9.14   |
| 40  | 100.00 | 0.00  | 40.02  | 8.22   | 9.65   |
| 50  | 100.00 | 0.00  | 50.03  | 7.35   | 9.10   |
| 75  | 100.00 | 0.00  | 75.02  | 5.37   | 7.06   |
| 100 | 99.70  | 0.30  | 99.69  | 6.93   | 10.56  |
| 150 | 97.36  | 2.64  | 146.01 | 14.20  | 25.15  |
| 200 | 94.98  | 5.02  | 189.71 | 24.36  | 41.27  |
| 250 | 83.91  | 16.09 | 209.46 | 61.07  | 83.70  |
| 300 | 62.91  | 37.09 | 188.10 | 118.63 | 146.20 |
