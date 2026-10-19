# My PDF 0.1.0-preview.1

First native Windows x64 preview from oneworkspacestudio. One main executable, embedded branded web UI, protected NTFS private data, verified MyAccounts OIDC and standalone local mode where useful. No Docker, WSL, Node.js or Python required at runtime. My PDF uses a bounded child worker of its own EXE.

Download the ZIP for the app, operations/recovery guides, compatibility contract, full third-party license texts and dependency inventory. SHA256SUMS.txt verifies the release files. One Workspace preview.37 installs these new defaults automatically unless explicitly excluded. Independent updates use workspace-update.json; My Search's Notes connector needs Notes preview.5 or newer (use Update all).

Local Windows Go tests/vet, frontend/AUI contracts, native Chrome desktop/mobile flows, consistent encrypted restore and managed suite tests passed. Managed checks include one sign-in, Drive/PDF copy round trip, actual Monitor discovery, real federated search, moved/deleted records, HTML injection safety, second-user sharing/revocation and logout. See docs/ACCEPTANCE.md for exact evidence and limits.

Unsigned preview. Another clean physical host, actual reboot/service acceptance and independent penetration audit remain unverified. PDF raster exports lose text/forms/accessibility/signature validity; OCR and Office conversion are not included. Mail bodies/Chat messages are not searched; Vault secrets are excluded. Go source scans have no reachable/imported-package findings but retain module-only findings. Existing Notes retains the documented moderate sprintf-js dependency advisory. No claim of a clean binary vulnerability scan or production certification.

Built and tested locally; GitHub Actions were not used.
