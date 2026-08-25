# Sky Concurrent Crawler

A bounded concurrent HTTP fetch service in Go for the SKYCOIN4444 engineering portfolio.

## Status

**Engineering beta.** The service accepts a bounded batch of public HTTP(S) URLs, fetches them concurrently, extracts page titles, and returns per-target status/error information. It includes SSRF-oriented destination checks, redirect validation, request/response size limits, server/client timeouts, race-tested concurrency, vulnerability scanning, and a non-root container.

It is not a general search-engine crawler: recursive link discovery, robots.txt policy, persistence, distributed scheduling, authentication, tenant isolation, proxy pools, and production deployment are not implemented.

## API

- `GET /health`
- `GET /ready`
- `POST /api/v1/crawl`

Example:

```json
{"urls":["https://example.com","https://example.org"],"workers":4}
```

Requests are limited to 1–32 URLs and 1–16 workers. Each fetched response is capped at 1 MiB. Redirect chains are capped at five hops and every redirect is revalidated.

## Security boundary

The default validator rejects malformed/non-HTTP URLs, URL userinfo, loopback, private, link-local, unspecified, and multicast destinations after DNS resolution. This reduces common SSRF paths but is not a substitute for egress firewalling or DNS-rebinding-resistant infrastructure controls.

## Verification

Requires Go 1.25.x:

```bash
gofmt -w .
go vet ./...
go test -race -count=1 ./...
go build .
```

CI also runs `govulncheck`, builds the container, verifies its configured user is non-root, and smoke-tests `/health`.

## Container

```bash
docker build -t sky-crawler .
docker run --rm -p 8080:8080 sky-crawler
```

The runtime image is distroless and uses UID/GID 65532.

## SKYCOIN4444 integration

Keep this independently deployable and consume it through the HTTP contract for bounded metadata-fetch jobs. Before using it for untrusted production workloads, add authenticated access, egress policy, durable job state, rate limiting, observability aggregation, and explicit robots/content-policy handling.

## License

See `LICENSE`.
