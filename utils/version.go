package utils

import (
	"lisk/globals"
	"lisk/httpClient"
	"lisk/logger"
	"lisk/models"
)

func CheckVersion() error {
	client, err := httpClient.NewHttpClient("")
	if err != nil {
		return err
	}

	var version models.VersionInfo
	if err := client.SendJSONRequest(globals.LinkRepo, "GET", nil, &version); err != nil {
		return err
	}

	if globals.SoftVersion != version.TagName {
		logger.GlobalLogger.Warnf("Your version (%s) is outdated. Latest version is %s \n", globals.SoftVersion, version.TagName)
		logger.GlobalLogger.Warnf("Read about the changes here: %s", version.HTMLURL)
	}

	return nil
}
