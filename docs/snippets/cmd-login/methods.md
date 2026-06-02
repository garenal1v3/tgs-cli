| Method | Description |
|--------|-------------|
| `desktop` | Import session from Telegram Desktop. No phone or code needed. |
| `code` | Phone number + SMS/Telegram verification code. Supports 2FA. Can be fully non-interactive with `--phone`, `--code`, and `--password` flags. |
| `qr` | Display QR code in terminal. Scan with Telegram on another device. Accounts with 2FA enabled cannot use QR login — use `code` with `--password` instead. |