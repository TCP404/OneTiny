package cmd

type releaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type releaseAsset struct {
	URL         string `json:"url"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

type tagList struct {
	TagName string `json:"name"`
}
