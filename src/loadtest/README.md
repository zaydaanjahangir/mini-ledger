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

# Removed MaybeProduceDigest in /POST Transaction

| Rate (rps) | Success % | Fail % | Throughput (rps) | P95 (ms) | P99 (ms) |
|---:|---:|---:|---:|---:|---:|
| 75   | 100.00 | 0.00 | 75.03   | 1.65     | 2.48     |
| 100  | 100.00 | 0.00 | 100.03  | 1.37     | 2.04     |
| 150  | 99.93  | 0.07 | 149.93  | 1.50     | 2.01     |
| 200  | 99.73  | 0.27 | 199.49  | 1.57     | 2.32     |
| 250  | 99.88  | 0.12 | 249.73  | 1.34     | 2.00     |
| 300  | 99.76  | 0.24 | 299.29  | 1.42     | 2.23     |
| 400  | 99.52  | 0.48 | 398.06  | 1.33     | 2.21     |
| 550  | 99.48  | 0.52 | 547.19  | 1.25     | 1.77     |
| 750  | 99.34  | 0.66 | 745.09  | 1.02     | 1.55     |
| 1000 | 99.50  | 0.50 | 995.05  | 0.74     | 1.03     |
| 1500 | 97.93  | 2.07 | 1468.92 | 0.33     | 3.66     |
| 2000 | 98.33  | 1.67 | 1966.59 | 0.30     | 2.25     |
| 3000 | 94.71  | 5.29 | 2841.41 | 0.42     | 32.81    |
| 3710 | 8.21   | 91.79 | 158.95  | 25046.60 | 30000.44 |
| 5000 | 69.18  | 30.82 | 3459.03 | 56.22    | 164.76   |
