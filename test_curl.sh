#!/bin/bash

# URL of the API endpoint
url="http://3.115.106.100:8080/queryComment"

# Function to send POST request
send_post_request() {
    response=$(curl -s -w "%{http_code}" -o /dev/null -X POST -H "Content-Type: application/json" -d '{"postID":"147312379047182396"}' 'http://3.115.106.100:8080/queryComment')
    timestamp=$(date +"%Y-%m-%d %H:%M:%S")
    if [ "$response" -eq 500 ]; then
    echo "$timestamp - 500 Internal Server Error"
    else
    echo "$timestamp - Request sent successfully with status code: $response"
    fi
}

# Infinite loop to run the function every 15 seconds
while true; do
  send_post_request
  sleep 60
done
