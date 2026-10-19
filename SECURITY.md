# Security

Use a supported Windows host, a dedicated NTFS data folder, HTTPS at the managed Workspace origin and the pinned MyAccounts issuer. Do not expose the app loopback listener directly. Managed mode does not accept standalone passwords. Data access uses immutable issuer/subject keys; an administrator role does not grant access to another user's application records through ordinary APIs. Infrastructure controls and full encrypted backups are owner-only.

App backups use Argon2id and authenticated chunked AES-GCM; restore into a new folder, with sessions cleared and object checksums/database integrity checked. Protect the password separately. Windows Job limits for PDF workers constrain resources and descendants; they are not a separate operating-system security sandbox. Browser PDF copies can invalidate digital signatures. Do not use drawing marks as permanent redaction.

Do not submit private files, passwords, tokens or personal message contents in public issues. Use a synthetic reproducer. These are unsigned preview builds; clean-host/reboot acceptance is recorded separately in docs/ACCEPTANCE.md.
