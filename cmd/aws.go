package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/charmbracelet/log"
)

func GetTokenForRegion(region string) (string, error) {
	accessKey, isPresent := os.LookupEnv("AWS_ACCESS_KEY_ID")
	if !isPresent {
		log.Fatalf("Unable to find AWS_ACCESS_KEY_ID envvar!!")
	}
	secretKey, isPresent := os.LookupEnv("AWS_SECRET_ACCESS_KEY")
	if !isPresent {
		log.Fatalf("Unable to find AWS_SECRET_ACCESS_KEY envvar!!")
	}

	log.Debug("Authenticating with AWS")

	accessKeySanitize := strings.ReplaceAll(strings.TrimSpace(accessKey), "\n", "")
	secretKeySanitize := strings.ReplaceAll(strings.TrimSpace(secretKey), "\n", "")
	log.Debugf("aws-credentials : %s:%s", accessKeySanitize, secretKeySanitize)
	creds := credentials.NewStaticCredentialsProvider(accessKeySanitize, secretKeySanitize, "")
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(region))

	// cfg, err := func(skip bool) (aws.Config, error) {
	// 	log.Debug("Some var ", skip)
	// 	if skip {
	// 		creds := credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")
	// 		log.Debugf("aws-credentials : %s:%s", accessKey, secretKey)
	// 		return config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(creds), config.WithRegion(region))
	// 	} else {
	// 		log.Info("looking for default AWS credentials...")
	// 		return config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	// 	}
	// }(true)

	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	result, err := ecrClient.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Fatalf("Failed to get ECR authorization token: %v", err)
	}

	decodedToken, err := base64.StdEncoding.DecodeString(*result.AuthorizationData[0].AuthorizationToken)
	if err != nil {
		log.Errorf("Error decoding base64 token: %s", err)
		return "", err
	}

	// Split the token to get the actual token part
	tokenParts := strings.Split(string(decodedToken), ":")
	if len(tokenParts) != 2 {
		return "", fmt.Errorf("invalid token format: expected AWS:token, got %s", string(decodedToken))
	}

	return tokenParts[1], nil
}
