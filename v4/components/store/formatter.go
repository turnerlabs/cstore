package store

import (
	"fmt"
	"strings"
)

func formatFileToken(baseToken, contextID, fileKey string) string {

	if strings.Index(contextID, "-") > -1 {
		contextID = contextID[0:strings.Index(contextID, "-")]
	}

	if len(fileKey) > 0 {
		return fmt.Sprintf("%s-%s-%s", baseToken, contextID, fileKey)
	}

	return fmt.Sprintf("%s-%s", baseToken, contextID)
}

func formatSecretToken(context, path, secret string) string {
	return fmt.Sprintf("%s/%s/%s", context, path, strings.ToLower(secret))
}
