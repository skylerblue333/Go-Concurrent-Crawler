# Go Concurrent Crawler

Small Go concurrency demonstration/service boundary for bounded crawling workflows.

## Implemented

- Concurrent crawl orchestration with goroutines
- `sync.Map` visited-set protection
- Worker-slot semaphore
- Context cancellation and timeout handling
- Buffered result collection
- HTTP endpoint at `POST /api/v1/crawl`

## Important limitation

The current crawler uses **simulated fetching and link discovery** (`time.Sleep` and generated `/a`/`/b` links). It does not currently fetch arbitrary websites or parse real HTML.

Therefore this repository is **not yet a production web crawler**. It is a concurrency foundation that can be connected to a real HTTP client/parser when that capability is required by the canonical ecosystem.

## Ecosystem role

Potential canonical boundary: **Supporting Services / Data Ingestion**. Prefer integrating this concurrency pattern into the canonical data-ingestion pipeline rather than creating an unnecessary standalone production microservice.

## Validation

The repository contains Go tests and CI configuration according to the existing repository audit. Passing status must be established from actual workflow/test evidence; this README makes no unsupported production-readiness claim.

## License

See the repository license and existing source files for applicable terms.
