package service

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"sync"

	"github.com/nkrus/gitlab-flagman/config"
	"github.com/nkrus/gitlab-flagman/internal/client"
)

const maxConcurrency = 5

type FeatureFlagService struct {
	GitLabClient *client.GitLabClient
}

func (ffs *FeatureFlagService) SyncFeatureFlags(flags []config.FeatureFlag) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Printf("Total flags in config: %d", len(flags))

	existingFlags, err := ffs.GitLabClient.GetAllFeatureFlags(ctx)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing feature flags: %w", err)
	}

	log.Printf("Total flags found remotely: %d", len(existingFlags))
	remoteFlagMap := make(map[string]config.FeatureFlag)
	for _, ef := range existingFlags {
		remoteFlagMap[ef.Name] = ef
	}

	desiredFlagMap := make(map[string]config.FeatureFlag)
	for _, df := range flags {
		desiredFlagMap[df.Name] = df
	}

	var flagsToAdd []config.FeatureFlag
	var flagsToUpdate []config.FeatureFlag
	var flagsToDelete []string

	for _, flag := range flags {
		if remoteFlag, exists := remoteFlagMap[flag.Name]; exists {
			if !flagsEqual(remoteFlag, flag) {
				flagsToUpdate = append(flagsToUpdate, flag)
			}
		} else {
			flagsToAdd = append(flagsToAdd, flag)
		}
	}

	for _, existingFlag := range existingFlags {
		if _, exists := desiredFlagMap[existingFlag.Name]; !exists {
			flagsToDelete = append(flagsToDelete, existingFlag.Name)
		}
	}

	log.Printf("Flags to delete: %d", len(flagsToDelete))
	log.Printf("Flags to add: %d", len(flagsToAdd))
	log.Printf("Flags to update: %d", len(flagsToUpdate))

	log.Println("Synchronization process started")

	if err := processFlagsConcurrently(ctx, flagsToDelete, ffs.deleteFlag, maxConcurrency); err != nil {
		return fmt.Errorf("failed to delete feature flags: %w", err)
	}

	if err := processFlagsConcurrently(ctx, flagsToAdd, ffs.addFlag, maxConcurrency); err != nil {
		return fmt.Errorf("failed to add feature flags: %w", err)
	}

	if err := processFlagsConcurrently(ctx, flagsToUpdate, ffs.updateFlag, maxConcurrency); err != nil {
		return fmt.Errorf("failed to update feature flags: %w", err)
	}
	log.Printf("Synced %d flags successfully", len(flagsToDelete)+len(flagsToAdd)+len(flagsToUpdate))

	return nil
}

func flagsEqual(a, b config.FeatureFlag) bool {
	if a.Name != b.Name || a.Description != b.Description || a.Active != b.Active {
		return false
	}

	if len(a.Strategies) != len(b.Strategies) {
		return false
	}

	for i := range a.Strategies {
		if a.Strategies[i].Name != b.Strategies[i].Name {
			return false
		}

		if !mapsEqual(a.Strategies[i].Parameters, b.Strategies[i].Parameters) {
			return false
		}

		if len(a.Strategies[i].Scopes) != len(b.Strategies[i].Scopes) {
			return false
		}
		aScopes := make(map[string]bool, len(a.Strategies[i].Scopes))
		for _, scope := range a.Strategies[i].Scopes {
			aScopes[scope.Environment] = true
		}
		for _, scope := range b.Strategies[i].Scopes {
			if !aScopes[scope.Environment] {
				return false
			}
		}
	}

	return true
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}
		if !reflect.DeepEqual(valueA, valueB) {
			return false
		}
	}

	return true
}

func processFlagsConcurrently[T any](
	ctx context.Context,
	items []T,
	action func(context.Context, T) error,
	concurrency int,
) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(items))
	sem := make(chan struct{}, concurrency)

	for _, item := range items {
		wg.Add(1)
		sem <- struct{}{}
		go func(item T) {
			defer wg.Done()
			defer func() { <-sem }()

			if err := action(ctx, item); err != nil {
				errChan <- err
			}
		}(item)
	}

	wg.Wait()
	close(errChan)

	if len(errChan) > 0 {
		return <-errChan
	}
	return nil
}

func (ffs *FeatureFlagService) addFlag(ctx context.Context, flag config.FeatureFlag) error {
	return ffs.GitLabClient.CreateFeatureFlag(ctx, flag)
}

func (ffs *FeatureFlagService) deleteFlag(ctx context.Context, flagName string) error {
	return ffs.GitLabClient.DeleteFeatureFlag(ctx, flagName)
}

func (ffs *FeatureFlagService) updateFlag(ctx context.Context, flag config.FeatureFlag) error {
	if err := ffs.GitLabClient.DeleteFeatureFlag(ctx, flag.Name); err != nil {
		return err
	}
	return ffs.GitLabClient.CreateFeatureFlag(ctx, flag)
}
