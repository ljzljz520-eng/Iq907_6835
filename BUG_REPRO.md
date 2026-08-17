# BUG_REPRO

The following failures were observed while validating the initial project state.
Each section records what failed, how to reproduce it, and the complete command output.
They are preserved intentionally; only failing build gates are omitted from the generated Dockerfile.

## Failure 1: Go test (.)

- Observed problem: `Go test (.)` failed in the initial project state.
- Working directory: `.`
- Command: `cd /app && GOTOOLCHAIN=local GOPROXY=off GOSUMDB=off go test -count=1 ./...`
- Exit status: `1`

```text
?   	volunteertraining/cmd/trainingd	[no test files]
ok  	volunteertraining/integration	0.587s
?   	volunteertraining/internal/app	[no test files]
?   	volunteertraining/internal/audit	[no test files]
ok  	volunteertraining/internal/auth	0.128s
ok  	volunteertraining/internal/catalog	0.129s
ok  	volunteertraining/internal/domain	0.018s
ok  	volunteertraining/internal/httpapi	0.334s
ok  	volunteertraining/internal/reporting	0.256s
ok  	volunteertraining/internal/store	0.328s
--- FAIL: TestBatchStopsWhenContextCancelled (0.30s)
    batch_test.go:70: progress=3
FAIL
FAIL	volunteertraining/internal/training	0.364s
FAIL
```

## Architecture reproduction

### linux/amd64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/trainingd): exit `0`
### linux/arm64
- Go toolchain version: exit `0`
- Go build (.): exit `0`
- Go test (.): exit `1`
- Go run smoke (cmd/trainingd): exit `0`
