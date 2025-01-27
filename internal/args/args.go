package args

import (
	"flag"
	"fmt"
	"log"
)

type Args struct {
	FlagsFile            string
	GitLabBase           string
	GitLabToken          string
	GitLabProjectID      string
	GitLabRequestTimeout int
}

var args Args

const (
	defaultFlagsFile  = "feature_flags.yaml"
	defaultGitLabBase = "https://gitlab.com/api/v4"
)

func RegisterFlags() {
	flag.StringVar(&args.FlagsFile, "flagsFile", defaultFlagsFile, "Путь к файлу с фичами")
	flag.StringVar(&args.GitLabBase, "gitLabBase", defaultGitLabBase, "Базовый URL GitLab API")
	flag.StringVar(&args.GitLabToken, "gitLabToken", "", "Токен доступа к GitLab")
	flag.StringVar(&args.GitLabProjectID, "gitLabProjectID", "", "ID проекта в GitLab")
	flag.IntVar(&args.GitLabRequestTimeout, "gitLabRequestTimeout", 10, "Таймаут ожидания ответа от Gitlab")
}
func init() {
	RegisterFlags()
}

func ParseArgs() (*Args, error) {
	flag.Parse()

	if !isFlagPassed("gitLabToken") {
		return nil, fmt.Errorf("-gitLabToken обязателен")
	}
	if !isFlagPassed("gitLabProjectID") {
		return nil, fmt.Errorf("-gitLabProjectID обязателен")
	}

	logArgs(&args)

	return &args, nil
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func logArgs(config *Args) {
	log.Printf(`
Using parameters:
-------------------- 
flagsFile: %q 
gitLabBase: %q 
gitLabProjectID: %q 
gitLabRequestTimeout: %ds
-------------------- `,
		config.FlagsFile, config.GitLabBase, config.GitLabProjectID, config.GitLabRequestTimeout)
}
