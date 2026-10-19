# Acceptance evidence

Release candidate checks are recorded in the repository history and release notes. Local Windows x64 tests use synthetic accounts and NTFS folders, never owner data. No hosted CI or GitHub Actions. This is a preview release.

Shared foundation tests cover immutable identity, OIDC code/PKCE/state/nonce/issuer/audience/expiry/replay, CSRF, local/managed mode separation, session/logout, migration rollback, NTFS/reparse rejection, consistent encrypted backup, wrong-password/tamper rejection, object/database integrity and restored-session invalidation.

Application Go tests and TypeScript/frontend contract tests are run by scripts/build.ps1. Browser end-to-end tests require a running disposable native app, APP_E2E_URL and APP_E2E_PASSWORD. Tests must not be pointed at a production installation. Test scripts create and modify synthetic records. Screenshots use synthetic records.

Not certified: another clean physical machine, actual reboot/Windows service transition, independent penetration audit, every assistive technology, every malformed third-party PDF or mail/chat search. Native no-runtime packaging is verified on this Windows host; no claim that a second clean machine was tested.

## Windows host results — 2026-10-19

Passed local TypeScript/Vite/AUI regeneration contracts, Go tests and vet, native Windows GUI EXE packaging, and standalone Chrome Playwright desktop/mobile flow. All four app encrypted backups restored to new protected NTFS folders with matching application state and rejected old sessions. PDF redaction output independently checked with pypdf: no original extractable text or annotations; original bytes preserved. Encrypted PDF password handling and a real detached-CMS self-signed fixture are tested without claiming trusted-signature validation.

Fresh native Workspace installation passed: one MyAccounts sign-in opened Tasks/PDF/Monitor/Search/Drive/Notes without another password; real Drive upload -> PDF picker import -> save-new-copy produced matching ordinary-file bytes; Monitor discovered live app checks and kept public status off; Search returned actual Tasks/Notes/Drive results without HTML execution and refreshed moved/deleted files and deleted tasks. A second MyAccounts user could not read a private task, could read a viewer-shared list but could not edit, and lost access on revocation. Non-owner Monitor API returned 403. Workspace logout removed all four app session cookies. The native updater applied a verified PDF package transaction, backed up and restarted all seven native processes successfully. Automated online/offline catalog regressions cover new default apps, opt-outs, new client contracts and custom-port preservation.

Go source vulnerability scans for Tasks and PDF found no reachable or imported-package vulnerabilities; module-only findings remain. A stripped-binary scan reports a deprecated OpenPGP module finding although source dependency traversal has no OpenPGP import path. This is recorded, not a clean binary-scan claim. New app runtime npm audit (including PDF dependencies) reported zero findings. Notes, an existing separate AGPL source dependency, retains the documented moderate sprintf-js advisory.

A second physical clean host, actual reboot/service transition and independent security audit remain unverified. Local test scripts and synthetic evidence are available; these limits are why the distribution is explicitly a preview.
