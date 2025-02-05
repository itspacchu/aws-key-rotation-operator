package main

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
)

func main() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-west-2"))
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	ecrClient := ecr.NewFromConfig(cfg)
	result, err := ecrClient.GetAuthorizationToken(context.TODO(), &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		log.Fatalf("Failed to get ECR authorization token: %v", err)
	}

	fmt.Println("Authorization Token:", *result.AuthorizationData[0].AuthorizationToken)
}
