package version

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/go-github/v53/github"
	log "github.com/sirupsen/logrus"
)

var (
	errUnableToRetrieveLatestVersion = errors.New("unable to fetch the latest version from GitHub")
)

// Version defines the version for the binary, this is actually set by GoReleaser.
var Version = "main"

// Template controls how the version is displayed
var Template = fmt.Sprintf("version %s\n", Version)

// CheckLatest asks GitHub
func CheckLatest() {
	if Version != "main" {
		latest, err := getLatestFromGitHub()
		if err != nil {
			fmt.Println(err)
			return
		}

		version := Version
		if latest != "" && (strings.TrimPrefix(latest, "v") != strings.TrimPrefix(version, "v")) {
			fmt.Printf("A newer version of auth-gateway is available, please update to %s\n", latest)
		} else {
			fmt.Println("You are on the latest version")
		}
	}
}

func getLatestFromGitHub() (string, error) {
	fmt.Print("checking latest version... ")
	c := github.NewClient(nil)
	repoRelease, resp, err := c.Repositories.GetLatestRelease(context.Background(), "cortexproject", "auth-gateway")
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Debugln("error while retrieving the latest version")
		return "", errUnableToRetrieveLatestVersion
	}

	if resp.StatusCode/100 != 2 {
		log.WithFields(log.Fields{"status": resp.StatusCode}).Debugln("non-2xx status code while contacting the GitHub API")
		return "", errUnableToRetrieveLatestVersion
	}

	return *repoRelease.TagName, nil
}
