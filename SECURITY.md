# Security Policy

Sky Concurrent Crawler is an engineering-beta service. The default URL validator rejects non-HTTP(S), URL userinfo, and resolved loopback/private/link-local/unspecified/multicast destinations. Redirect targets are revalidated, response bodies are bounded, and the runtime container is non-root.

These controls do not eliminate all SSRF or DNS-rebinding risk. Production deployment should add authenticated access, strict outbound network policy, rate limiting, centralized audit logging, DNS controls, and allowlisting appropriate to the environment.

Do not report secrets or exploit payloads in public issues. Use the repository owner's private GitHub security/contact channel when available.
