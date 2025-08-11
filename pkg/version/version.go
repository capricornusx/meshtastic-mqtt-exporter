package version

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func GetVersion() string {
	return Version
}

func GetBuildInfo() (string, string, string) {
	return Version, GitCommit, BuildDate
}
