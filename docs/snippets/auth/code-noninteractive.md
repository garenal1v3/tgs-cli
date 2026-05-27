```bash
# Step 1: request the code
tgs login --type code --phone +1234567890
# → {"status":"code_sent","profile":"default"}

# Step 2: complete login with the received code
tgs login --type code --phone +1234567890 --code 12345
# → {"status":"logged_in","user":{...}}

# With 2FA
tgs login --type code --phone +1234567890 --code 12345 --password mysecret
```
