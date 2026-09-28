# Code style

- Do not add Go tests (`_test.go`) unless explicitly asked.
- Keep comments sparse: only for genuinely non-obvious mechanics, not one per block.

# Headless Casdoor core

This fork intentionally contains Casdoor's backend and identity logic without
the bundled Web UI. Application-specific authentication interfaces live in
separate repositories and must integrate through the documented Casdoor APIs
and OIDC endpoints.

Do not add Golden, Silver, or another application's branding and page code here.
