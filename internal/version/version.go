package version

var Version string

const (
	VersionListURL   = "https://api.github.com/repos/tcp404/OneTiny/tags"
	VersionLatestURL = "https://api.github.com/repos/tcp404/OneTiny/releases/latest"
	VersionByTagURL  = "https://api.github.com/repos/tcp404/OneTiny/releases/tags/"
)

var ReleaseName = map[string]string{
	"linux":   "OneTiny",
	"windows": "OneTiny.exe",
	"darwin":  "OneTiny_mac",
}
