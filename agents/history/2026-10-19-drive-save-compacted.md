# Drive save repair

User reports My PDF files are not stored in Drive. Confirmed installed owner
manager preview.38 and MyPDF.exe are running; existing preview.2 intentionally
kept uploads/processing in its private library and required a separate save-copy.
Opening a document also did not select it, and picker filenames followed the
open original rather than a selected generated result. Those UX gaps explain
the observed missing ordinary Drive files; no owner data was used as a fixture.

Implemented managed default-on automatic save for new PDF uploads, image-to-PDF,
browser generated copies and native PDF results. Copies go through the actual
requesting user's Drive API into the first authorised scope's My PDF folder.
No server admin token/direct-filesystem bypass. Existing imports/legacy library
contents are not silently exported; per-document Save to Drive and Retry actions
are visible. Standalone remains local; ZIP/PNG/text downloads and editable mark
sidecars are not claimed as automatic PDF copies. Originals remain unchanged.

Schema2 adds durable per-document queue/receipts, five-minute leases serialising
tabs, target recorded before upload, expiry/retry, account ownership/CSRF/bounded
input. New startup skips migration1 when already applied and validates schema2.
Downgrading to the old executable requires a compatible app-data restore.
Receipts survive app/browser restart. Interrupted-response retries compare exact
bytes via SHA256; exclusive creation and numbered names preserve collisions.
App shows pending/saving/saved/error and actual destination. AUI regenerated.

Reconnect testing found an initial redirect race: the old page's finally/load
could claim a second save while navigating to Drive. Fixed by carrying connection
intent, deferring active leases before navigation and suppressing queue/reload
on the departing page. Batch uploads finish private staging before queue starts.

Native build, all Go tests/vet and eleven frontend checks passed. Ten synthetic
managed browser checks passed with actual Drive download byte comparison:
single MyAccounts login/native Drive connection, automatic upload, duplicates,
native rotation, browser numbering, custom-folder receipt, failed-transfer retry,
lost-response reconciliation without duplicates, old/local save action and
reload receipts. Separate backend test verifies receipt persistence across host
restart, lease expiry and cross-account denial. A test initially reopened App
on the same ServeMux (duplicate routes); corrected to close/reopen the Host.
No universal feature/security/clean-host certification implied.

Source/build/evidence on I:, protected isolated fixture only under named
C:/Projects/.cache/workspace-pdf-drive-20261019. Evidence/build tooling is
I:/Projects/tooling/pdf-drive-save. Latest UI notice refinement makes completed
saves replace the uploading notice. Rebuilt native binary and reran eleven
frontend checks; final native managed smoke passed automatic Drive reconnect,
completion notice and downloaded-byte equality on this exact build.

Changes currently uncommitted/unpublished and owner installation unchanged.
Asked whether to publish coordinated My PDF + Workspace releases and install;
await the human answer before external writes. A local EXE swap alone would
leave the updater pinned to preview.2 and cannot safely ship the schema upgrade.
Before final handoff stop the fixture and remove only its verified named C root,
preserving owner Apps/Private/Files. Append actual release/deployment/cleanup
results only after performed. Keep Actions disabled and all builds local.

## Final local handoff

Stopped the tracked synthetic manager and verified no task process or reparse
point references before removing the exact named C fixture (28,482,627 bytes).
Programs, source and evidence remain on I:; its deleted private data must not be
treated as a restartable installation. Owner installation was not modified.
No commit, push or release. Coordinated publication/deployment awaits the user's
answer to the explicit release/install question. Build and test outputs are in
the external tooling directory, with no owner credentials or data in source.

## Authorized publication follow-up

User explicitly requests push and release. Prepare My PDF v0.1.0-preview.3 and
coordinated One Workspace v0.3.0-preview.40, with local native builds, exact
component hashes and verified release assets. No GitHub Actions. This follow-up
authorizes publication; the owner live installation remains outside fixture
testing and is not claimed upgraded by creating releases.
