# API and integration

All paths below are relative to the app origin/subpath. GET /healthz is public health only. GET /api/bootstrap returns non-secret configuration. GET /api/me returns the signed-in account and session CSRF token. Writes require an exact Origin header plus X-CSRF-Token from that session; use same-origin credentials and never log tokens. Owner-only POST /api/backups streams the encrypted archive. Login/OIDC/session endpoints are provided by the public version-pinned One Workspace apphost package.

GET/POST /api/documents; GET/DELETE /api/documents/{id}; GET/PUT /api/documents/{id}/annotations; GET/POST /api/jobs; DELETE /api/jobs/{id}; GET/DELETE /api/artifacts/{id}. Upload is raw application/pdf plus X-File-Name; processing accepts operation/documents/pages/rotation/password. Job polling exposes queue position/state/message/result, not fake progress percentages. Passwords are memory-only; interrupted jobs cancel on restart.

Upload header `X-Save-To-Drive: true` and job field `saveToDrive: true` durably
queue ordinary PDF copies. Document listings include `drive` state/source/path.
`POST /api/documents/{id}/drive-save` accepts actions retry, claim, claimCopy,
target, saved, failed, defer and blocked. Claims return a private five-minute
lease; mutations require it, document ownership, Origin and session CSRF. Leases
serialise tabs and can be reclaimed after expiry. Target paths are recorded before
transferring. Interrupted-response retries compare actual destination bytes by
SHA256 before making another copy. Receipts never confer Drive access: the
browser verifies its OIDC immutable subject with Drive before file writes.
Actual Drive permissions, quotas and exclusive-create operations apply there.
Only requested PDFs enter this queue; legacy library contents are not silently
exported. Processing split archives does not queue a PDF save.

User resources are authorised on every request. Missing/inaccessible user records do not expose private titles. Limits and filters use bounded input; unsupported operations fail explicitly. No API grants direct database access to another app. Managed MyAccounts credentials are server-side only. Search and Drive integrations verify that the source's current immutable subject matches the requesting app, and recheck identity after search. Source permissions apply again when opening a link.
