# ADR002: My PDF

Decision: original Apache-2.0 app at oneworkspacestudio/my-pdf. PDF.js (Apache-2.0)
renders documents in the browser; pdf-lib (MIT) implements safe copy-oriented
browser composition; pdfcpu (Apache-2.0) provides native validation/assembly and
metadata operations in bounded child-worker mode of the same MyPDF.exe. Native
worker processes have time/memory/descendant limits. This is one main executable,
not a promise of one total process. qpdf (Apache-2.0) would add a native helper;
MuPDF is AGPL/commercial and is not selected; Stirling-PDF would add Java/runtime
and different licensing/packaging obligations. No upstream source is repackaged.
True redaction creates a brand-new raster PDF from painted page images, omitting
all original PDF objects, text, forms and metadata. This intentionally sacrifices
searchable text/accessibility on redacted output; confirm explicitly. A black
annotation is never called redaction. OCR and Office conversion are optional
unsupported capabilities until a separately managed engine passes acceptance.
Drive integration uses actual /drive/api/resources and download endpoints in the
requesting browser's existing user session. No admin token or direct filesystem
shortcut. Save a copy by default; atomic/version overwrite only with supported
Quantum contract and explicit confirmation. Signed originals remain unchanged;
processed copies cannot retain a claim of original signature validity.

Shared decision: native Go server, embedded AUI-generated React/TypeScript UI,
SQLite (modernc, BSD-3-Clause), protected per-app NTFS data folders and one EXE.
Reuse the existing One Workspace Apache-2.0 platform, OIDC, branding and supervisor
conventions through a public pkg/apphost package in the actual suite repository.
Canonicalise the suite Go module path to its existing organization repository;
there are no supported external consumers of its internal packages. Each original
app has a dedicated organization repository, rather than a misleading fork.
OIDC uses code+PKCE, state cookie, nonce, issuer/audience verification and immutable
issuer/subject keys. Session/CSRF secrets stay on the server. Local bootstrap is
loopback-only and explicitly disabled in managed mode. No GitHub Actions: user
instruction overrides the brief's Actions-template suggestion; run gates locally.
