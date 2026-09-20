# Repository audit — 2026-09-20

## Changes

Constrain remote paths and rooted writes; replace only completed downloads; require output flag; make LocalStack tests opt-in; repair invalid test name.

## Validation

Offline go test -race ./... passes with fake S3 client and path/interruption regressions.

## Limits

Requires Go 1.25+. LocalStack suite unverified: older testcontainers native initialization crashed on this M5 environment. Run integration tests explicitly with -tags integration after resolving runtime compatibility.
