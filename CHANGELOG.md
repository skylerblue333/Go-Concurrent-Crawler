# Changelog

## Unreleased

- Replaced simulated recursive crawling with a bounded concurrent HTTP fetch service.
- Added public-destination validation, redirect revalidation, response-size limits, timeouts, and deterministic unit tests.
- Added health/readiness endpoints, race/vet/govulncheck CI gates, non-root distroless packaging, and truthful integration/security documentation.
