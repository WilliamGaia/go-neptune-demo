#!/bin/bash

# Variables
REGION="ap-northeast-1"
SERVICE="neptune-db"
ENDPOINT="https://pkm-demo-instance-real.cluster-cfhqexhrq16e.ap-northeast-1.neptune.amazonaws.com:8182/openCypher"
PAYLOAD='{"query":"MATCH (n) RETURN COUNT(n);"}'

# AWS SigV4 signing
SIGNED_REQUEST=$(aws-sigv4 sign --service $SERVICE --region $REGION --host $(echo $ENDPOINT | awk -F[/:] '{print $4}') --http-method POST --payload "$PAYLOAD" --canonical-uri /openCypher --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY)

# Execute the request with curl
curl -X POST "$ENDPOINT" \
    -H "Content-Type: application/json" \
    -H "Authorization: $(echo "$SIGNED_REQUEST" | jq -r .headers.Authorization)" \
    -H "x-amz-date: $(echo "$SIGNED_REQUEST" | jq -r .headers."x-amz-date")" \
    -H "x-amz-security-token: $(echo "$SIGNED_REQUEST" | jq -r .headers."x-amz-security-token")" \
    -d "$PAYLOAD"
    # -w "@performance-format.txt"
