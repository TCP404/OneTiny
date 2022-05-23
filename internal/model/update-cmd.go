package model

type ReleaseInfo struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleaseAsset `json:"assets"`
}

type ReleaseAsset struct {
	URL         string `json:"url"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

type TagList struct {
	TagName string `json:"name"`
}
