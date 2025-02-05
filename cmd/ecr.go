package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/log"
)

type DockerConfig struct {
	Auths map[string]Auth `json:"auths"`
}

type Auth struct {
	Auth string `json:"auth"`
}

func parseDockerURI(uri string) (string, string, string, error) {
	parts := strings.Split(uri, ".dkr.ecr.")
	if len(parts) < 2 {
		log.Warnf("Invalid Docker URI : %s", uri)
		return "", "", "", fmt.Errorf("not a AWS URL")
	}

	accountID := parts[0]
	regionAndDomain := strings.Split(parts[1], ".amazonaws.com")
	if len(regionAndDomain) < 1 {
		log.Warnf("Invalid Docker URI : %s", uri)
		return "", "", "", fmt.Errorf("not a AWS URL")
	}

	region := regionAndDomain[0]
	imageNameAndTag := strings.Split(regionAndDomain[1], ":")
	if len(imageNameAndTag) < 2 {
		log.Warnf("Invalid Docker URI : %s", uri)
		return "", "", "", fmt.Errorf("not a AWS URL")
	}
	imageName := imageNameAndTag[0]

	return accountID, region, imageName, nil
}
