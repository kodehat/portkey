package build

// Variables that can be injected during build.
var (

	// BuildTime time when the application was build.
	BuildTime string = "unknown"

	// CommitHash Git commit hash of the built application.
	CommitHash string

	// Version version of the built application.
	Version string = "dev"

	// GoVersion Go version the application was build with.
	GoVersion string = "unknown"
)

// BuildDetails struct that contains information about the built application.
type BuildDetails struct {

	// BuildTime time when the application was build.
	BuildTime string

	// CommitHash Git commit hash of the built application.
	CommitHash string

	// Version version of the built application.
	Version string

	// GoVersion Go version the application was build with.
	GoVersion string
}

// B contains the build details of the application.
var B BuildDetails = newBuildDetails()

// newBuildDetails loads the build details struct B during startup.
func newBuildDetails() BuildDetails {
	return BuildDetails{
		BuildTime:  BuildTime,
		CommitHash: CommitHash,
		Version:    Version,
		GoVersion:  GoVersion,
	}
}
