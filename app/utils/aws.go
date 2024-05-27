package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	v4 "github.com/aws/aws-sdk-go/aws/signer/v4"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func FetchAWSSignedToken(ctx context.Context) (neo4j.AuthToken, *time.Time, error) {
	ServiceName := os.Getenv("SERVICENAME")
	DummyUsername := os.Getenv("USERNAME")
	region := os.Getenv("REGION")
	hostAndPort := os.Getenv("HOSTPORT")
	// Create a new request
	req, err := http.NewRequest(http.MethodGet, "https://"+hostAndPort+"/opencypher", nil)
	if err != nil {
		return neo4j.AuthToken{}, nil, err
	}

	// Create a signer and sign the request
	signer := v4.NewSigner(credentials.NewEnvCredentials())
	_, err = signer.Sign(req, nil, ServiceName, region, time.Now())
	if err != nil {
		return neo4j.AuthToken{}, nil, err
	}

	// Extract the necessary headers for the auth token
	hdrs := []string{"Authorization", "X-Amz-Date", "X-Amz-Security-Token"}
	hdrMap := make(map[string]string)
	for _, h := range hdrs {
		hdrMap[h] = req.Header.Get(h)
	}

	hdrMap["Host"] = req.Host
	hdrMap["HttpMethod"] = req.Method

	// Create the auth token
	password, err := json.Marshal(hdrMap)
	if err != nil {
		return neo4j.AuthToken{}, nil, err
	}
	authToken := neo4j.BasicAuth(DummyUsername, string(password), "")

	// Set token expiration time to 60 minutes (Amazon credentials generally expire every hour)
	expiresIn := time.Now().Add(15 * time.Minute)
	// Add a buffer to refresh the token before it expires
	expiresIn = expiresIn.Add(-5 * time.Minute)
	fmt.Printf("Generated token expires at: %v\n", expiresIn)

	return authToken, &expiresIn, nil
}
