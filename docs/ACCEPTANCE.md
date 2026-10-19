# Acceptance evidence

Release candidate checks are recorded in the repository history and release notes. Local Windows x64 tests use synthetic accounts and NTFS folders, never owner data. No hosted CI or GitHub Actions. This is a preview release.

Shared foundation tests cover immutable identity, OIDC code/PKCE/state/nonce/issuer/audience/expiry/replay, CSRF, local/managed mode separation, session/logout, migration rollback, NTFS/reparse rejection, consistent encrypted backup, wrong-password/tamper rejection, object/database integrity and restored-session invalidation.

Application Go tests and TypeScript/frontend contract tests are run by scripts/build.ps1. Browser end-to-end tests require a running disposable native app, APP_E2E_URL and APP_E2E_PASSWORD. Tests must not be pointed at a production installation. Test scripts create and modify synthetic records. Screenshots use synthetic records.

Not certified: another clean physical machine, actual reboot/Windows service transition, independent penetration audit, every assistive technology, every malformed third-party PDF or mail/chat search. Native no-runtime packaging is verified on this Windows host; no claim that a second clean machine was tested.
