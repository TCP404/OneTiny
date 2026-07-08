package file

import (
	"bytes"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/tcp404/OneTiny/internal/share"
	"github.com/tcp404/OneTiny/resource"
)

func TestListTemplateRendersPolishedFileBrowserControls(t *testing.T) {
	tpl, err := template.ParseFS(resource.FS, "template/*.tpl")
	if err != nil {
		t.Fatalf("ParseFS returned error: %v", err)
	}

	var body bytes.Buffer
	err = tpl.ExecuteTemplate(&body, "list.tpl", gin.H{
		"pathTitle": "/projects/design",
		"breadcrumbs": []share.Breadcrumb{
			{Name: "根目录", URL: "/file/?action=view"},
			{Name: "projects", URL: "/file/projects/?action=view"},
			{Name: "design", URL: "/file/projects/design/?action=view", Current: true},
		},
		"upload": true,
		"files": []share.Entry{
			{
				Name:       "screenshots",
				URLRelPath: "/file/projects/design/screenshots/",
				Size:       "0 B",
				IsDir:      true,
			},
			{
				Name:       "release-notes.md",
				URLRelPath: "/file/projects/design/release-notes.md",
				Size:       "18 KB",
				IsDir:      false,
			},
		},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate returned error: %v", err)
	}

	html := body.String()
	assertContains(t, html, `id="fileSearch"`)
	assertContains(t, html, `<img class="brand-logo" src="/logo.png" alt="OneTiny">`)
	assertContains(t, html, `data-shortcut="mod+k"`)
	assertContains(t, html, `data-total-count`)
	assertContains(t, html, `data-upload-submit disabled`)
	assertContains(t, html, `event.preventDefault()`)
	assertContains(t, html, `class="breadcrumb-link" href="/file/?action=view"`)
	assertContains(t, html, `class="breadcrumb-link" href="/file/projects/?action=view"`)
	assertContains(t, html, `class="breadcrumb-current" aria-current="page">design</span>`)
	assertContains(t, html, `data-view-toggle="list"`)
	assertContains(t, html, `data-view-toggle="grid"`)
	assertContains(t, html, `data-open-url="/file/projects/design/screenshots/?action=view"`)
	assertContains(t, html, `data-open-url="/file/projects/design/release-notes.md?action=view"`)
	assertContains(t, html, `href="/file/projects/design/screenshots/?action=dl"`)
	assertContains(t, html, `href="/file/projects/design/release-notes.md?action=dl"`)
	assertContains(t, html, `data-file-entry`)
	assertContains(t, html, `data-countable`)
	assertContains(t, html, `data-search-text="screenshots"`)
	assertContains(t, html, `event.metaKey || event.ctrlKey`)
	assertContains(t, html, `function setView(view)`)
	assertContains(t, html, `onetiny.fileView`)
	assertContains(t, html, `上传到当前目录`)
	assertNotContains(t, html, `data-focus-upload`)
	assertNotContains(t, html, `目录，点击进入`)
	assertNotContains(t, html, `文件，点击打开`)
}

func TestBuildBreadcrumbsRendersRootWithoutDuplicateSlash(t *testing.T) {
	tpl, err := template.ParseFS(resource.FS, "template/*.tpl")
	if err != nil {
		t.Fatalf("ParseFS returned error: %v", err)
	}

	var body bytes.Buffer
	err = tpl.ExecuteTemplate(&body, "list.tpl", gin.H{
		"pathTitle":   "/",
		"breadcrumbs": share.BuildBreadcrumbs("/"),
		"files":       []share.Entry{},
	})
	if err != nil {
		t.Fatalf("ExecuteTemplate returned error: %v", err)
	}

	html := body.String()
	assertContains(t, html, `class="breadcrumb-current" aria-current="page">根目录</span>`)
	assertNotContains(t, html, `当前目录</span>
            <span aria-hidden="true">/</span>
            <span class="path-value">/</span>`)
}

func TestBuildBreadcrumbsLinksAncestorDirectories(t *testing.T) {
	crumbs := share.BuildBreadcrumbs("projects/design")

	if len(crumbs) != 3 {
		t.Fatalf("len(crumbs) = %d, want 3: %+v", len(crumbs), crumbs)
	}
	if crumbs[0] != (share.Breadcrumb{Name: "根目录", URL: "/file/?action=view"}) {
		t.Fatalf("root crumb = %+v", crumbs[0])
	}
	if crumbs[1] != (share.Breadcrumb{Name: "projects", URL: "/file/projects/?action=view"}) {
		t.Fatalf("projects crumb = %+v", crumbs[1])
	}
	if crumbs[2] != (share.Breadcrumb{Name: "design", URL: "/file/projects/design/?action=view", Current: true}) {
		t.Fatalf("design crumb = %+v", crumbs[2])
	}
}

func TestGetFileInfosSortsDirectoriesBeforeFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "aaa-file.txt"), []byte("file"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Mkdir(filepath.Join(root, "zzz-dir"), 0o755); err != nil {
		t.Fatalf("mkdir dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "mmm-file.txt"), []byte("file"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	files, err := share.List(root, root)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	got := make([]string, len(files))
	for i, f := range files {
		got[i] = f.Name
	}

	want := []string{"zzz-dir", "aaa-file.txt", "mmm-file.txt"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("file order = %v, want %v", got, want)
	}
}

func assertContains(t *testing.T, value, want string) {
	t.Helper()
	if !strings.Contains(value, want) {
		t.Fatalf("rendered template does not contain %q\nbody:\n%s", want, value)
	}
}

func assertNotContains(t *testing.T, value, unwanted string) {
	t.Helper()
	if strings.Contains(value, unwanted) {
		t.Fatalf("rendered template contains unwanted %q\nbody:\n%s", unwanted, value)
	}
}
