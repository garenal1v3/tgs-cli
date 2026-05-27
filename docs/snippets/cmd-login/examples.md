```bash
# Import session from Telegram Desktop (default)
tgs login

# Custom tdata path
tgs login --desktop-dir ~/AppData/Roaming/Telegram\ Desktop/tdata

# Passcode-protected Telegram Desktop
tgs login --passcode mysecret

# Phone + verification code (interactive)
tgs login --type code

# Phone + code with phone pre-filled
tgs login --type code --phone +1234567890

# Non-interactive two-step login (for AI agents / automation)
# Step 1: send the verification code
tgs login --type code --phone +1234567890
# Step 2: complete login with the code
tgs login --type code --phone +1234567890 --code 12345

# Non-interactive login with 2FA
tgs login --type code --phone +1234567890 --code 12345 --password mysecret

# QR code login
tgs login --type qr

# Login to a specific profile
tgs login --profile work

# Login to a profile with a specific method
tgs login --type code --profile work --phone +0987654321
```
