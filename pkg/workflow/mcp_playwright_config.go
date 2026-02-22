package workflow

import (
	"strings"

	"github.com/github/gh-aw/pkg/constants"
)

// PlaywrightDockerArgs represents the common Docker arguments for Playwright container
type PlaywrightDockerArgs struct {
	ImageVersion      string // Version for Docker image (mcr.microsoft.com/playwright:version)
	MCPPackageVersion string // Version for NPM package (@playwright/mcp@version)
}

func getPlaywrightDockerImageVersion(playwrightConfig *PlaywrightToolConfig) string {
	playwrightDockerImageVersion := string(constants.DefaultPlaywrightBrowserVersion) // Default Playwright browser Docker image version
	// Extract version setting from tool properties
	if playwrightConfig != nil && playwrightConfig.Version != "" {
		playwrightDockerImageVersion = playwrightConfig.Version
	}
	return playwrightDockerImageVersion
}

// getPlaywrightMCPPackageVersion extracts version setting for the @playwright/mcp NPM package
// This is separate from the Docker image version because they follow different versioning schemes
func getPlaywrightMCPPackageVersion(playwrightConfig *PlaywrightToolConfig) string {
	// Always use the default @playwright/mcp package version.
	return string(constants.DefaultPlaywrightMCPVersion)
}

// generatePlaywrightDockerArgs creates the common Docker arguments for Playwright MCP server
func generatePlaywrightDockerArgs(playwrightConfig *PlaywrightToolConfig) PlaywrightDockerArgs {
	return PlaywrightDockerArgs{
		ImageVersion:      getPlaywrightDockerImageVersion(playwrightConfig),
		MCPPackageVersion: getPlaywrightMCPPackageVersion(playwrightConfig),
	}
}

// extractExpressionsFromPlaywrightArgs extracts all GitHub Actions expressions from playwright arguments
// Returns a map of environment variable names to their original expressions
// Uses the same ExpressionExtractor as used for shell script security
func extractExpressionsFromPlaywrightArgs(customArgs []string) map[string]string {
	if len(customArgs) == 0 {
		return make(map[string]string)
	}

	// Join all arguments with a separator that won't appear in expressions
	combined := strings.Join(customArgs, "\n")

	// Use ExpressionExtractor to find all expressions
	extractor := NewExpressionExtractor()
	mappings, err := extractor.ExtractExpressions(combined)
	if err != nil {
		return make(map[string]string)
	}

	// Convert to map of env var name -> original expression
	result := make(map[string]string)
	for _, mapping := range mappings {
		result[mapping.EnvVar] = mapping.Original
	}

	return result
}

// replaceExpressionsInPlaywrightArgs replaces all GitHub Actions expressions with environment variable references
// This prevents any expressions from being exposed in GitHub Actions logs
func replaceExpressionsInPlaywrightArgs(args []string, expressions map[string]string) []string {
	if len(expressions) == 0 {
		return args
	}

	// Create a temporary extractor with the same mappings
	combined := strings.Join(args, "\n")
	extractor := NewExpressionExtractor()
	_, _ = extractor.ExtractExpressions(combined)

	// Replace expressions in the combined string
	replaced := extractor.ReplaceExpressionsWithEnvVars(combined)

	// Split back into individual arguments
	return strings.Split(replaced, "\n")
}
