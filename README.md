# black-history-api

## Troubleshooting
-  ⚠️ MongoDB Atlas Whitelist Issue
Your MongoDB Atlas cluster must allow Heroku’s dynamic IPs or 0.0.0.0/0 (open to all):

✅ Fix:
Go to MongoDB Atlas: https://cloud.mongodb.com/v2/6862a4b57aa4f563a2a5feb5#/metrics/replicaSet/6862a5bd9d250e3a989a52fd/explorer/BlackHistory/figures/find

Go to Network Access → IP Whitelist

Add:

Copy
Edit
0.0.0.0/0
⚠️ For public APIs, this is temporarily OK. Secure it later using VPC peering or authentication restrictions.