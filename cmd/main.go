package main

import (
	"log"

	"gitlab.com/nkrus/gitlab-flagman/config"
	"gitlab.com/nkrus/gitlab-flagman/internal/args"
	"gitlab.com/nkrus/gitlab-flagman/internal/client"
	"gitlab.com/nkrus/gitlab-flagman/internal/service"
)

func main() {
	parsedArgs, err := args.ParseArgs()
	if err != nil {
		log.Fatalf("Error parsing arguments: %v", err)
	}

	featureFlags, err := config.ReadFlagsFromYAML(parsedArgs.FlagsFile)
	if err != nil {
		log.Fatalf("Error reading feature flags from file %q: %v", parsedArgs.FlagsFile, err)
	}

	gitLabClient := client.NewGitLabClient(
		parsedArgs.GitLabBase,
		parsedArgs.GitLabToken,
		parsedArgs.GitLabProjectID,
		parsedArgs.GitLabRequestTimeout,
	)

	featureFlagService := service.FeatureFlagService{GitLabClient: gitLabClient}
	if err := featureFlagService.SyncFeatureFlags(featureFlags); err != nil {
		log.Fatalf("Error syncing feature flags: %v", err)
	}
}
