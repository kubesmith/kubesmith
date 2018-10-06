package buildinfo

var (
	// Version is the current version of Kubesmith, set by the go linker's -X flag at build time.
	Version string

	// GitSHA is the actual commit that is being built, set by the go linker's -X flag at build time.
	GitSHA string
)
