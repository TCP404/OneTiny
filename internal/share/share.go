package share

import (
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tcp404/OneTiny/internal/kit/safepath"
	"github.com/tcp404/OneTiny/internal/server/routepath"
	"github.com/tcp404/eutil"
)

var BufferLimit = 512 * eutil.KB

type Entry struct {
	Size          string
	IsDir         bool
	DeviceAbsPath string
	URLRelPath    string
	Name          string
}

type Breadcrumb struct {
	Name    string
	URL     string
	Current bool
}

func ParseWildcardFilename(filename string) string {
	if filename == "" || filename == routepath.Root {
		return routepath.Root
	}
	return strings.TrimPrefix(filename, "/")
}

func CleanRelPath(filePath string) string {
	if filePath == routepath.Root {
		return routepath.Root
	}
	cleaned := filepath.Clean(filePath)
	if cleaned == "." {
		return routepath.Root
	}
	return cleaned
}

func IsDir(rootPath, filePath string) bool {
	if filePath == routepath.Root {
		return true
	}
	target, ok := safepath.ResolveExistingWithinRoot(rootPath, filePath)
	if !ok {
		return false
	}
	fInfo, err := os.Stat(target)
	if err != nil || fInfo == nil {
		return false
	}
	return fInfo.IsDir()
}

func IsOverLevel(rootPath string, maxLevel uint8, relPath string, isFile bool, isDownload bool) bool {
	cleanRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return true
	}
	target, ok := safepath.ResolveWithinRoot(rootPath, relPath)
	if !ok {
		return true
	}
	if PathExists(target) {
		target, ok = safepath.ResolveExistingWithinRoot(rootPath, relPath)
		if !ok {
			return true
		}
	}
	rel, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return true
	}
	parts := strings.Split(rel, string(filepath.Separator))
	level := len(parts)
	if parts[0] == "." {
		level = 0
	}
	if isFile || isDownload {
		level--
	}
	return level > int(maxLevel)
}

func PathExists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

func List(rootPath, absPath string) ([]Entry, error) {
	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		relPath = strings.TrimPrefix(absPath, rootPath)
	}
	relPath = filepath.ToSlash(relPath)
	entries := make([]Entry, len(dirEntries))

	for i, f := range dirEntries {
		info, _ := f.Info()
		size := info.Size()
		deviceAbsPath := filepath.Join(absPath, f.Name())
		urlRelPath := path.Join(routepath.FileGroupPrefix, relPath, f.Name())
		isDir := f.Type().IsDir()

		if isDir {
			size = 0
			deviceAbsPath += string(filepath.Separator)
			urlRelPath += string(filepath.Separator)
		}
		entries[i] = Entry{
			DeviceAbsPath: deviceAbsPath,
			URLRelPath:    urlRelPath,
			Name:          f.Name(),
			Size:          eutil.SizeFmt(size),
			IsDir:         isDir,
		}
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})
	return entries, nil
}

func BuildBreadcrumbs(rel string) []Breadcrumb {
	cleanRel := strings.Trim(filepath.ToSlash(rel), "/")
	if cleanRel == "" || rel == routepath.Root {
		return []Breadcrumb{{
			Name:    "根目录",
			URL:     routepath.FileGroupPrefix + "/?action=view",
			Current: true,
		}}
	}

	parts := strings.Split(cleanRel, "/")
	crumbs := make([]Breadcrumb, 0, len(parts)+1)
	crumbs = append(crumbs, Breadcrumb{
		Name: "根目录",
		URL:  routepath.FileGroupPrefix + "/?action=view",
	})

	for i, part := range parts {
		joined := path.Join(parts[:i+1]...)
		crumbs = append(crumbs, Breadcrumb{
			Name:    part,
			URL:     path.Join(routepath.FileGroupPrefix, joined) + "/?action=view",
			Current: i == len(parts)-1,
		})
	}
	return crumbs
}

func SafeUploadFilename(filename string) (string, bool) {
	base := path.Base(strings.ReplaceAll(filename, "\\", "/"))
	base = filepath.Base(base)
	if base == "" || base == "." || base == ".." {
		return "", false
	}
	return base, true
}

func ContentLen(absPath string) int64 {
	var contentLen int64 = -1
	info, err := os.Stat(absPath)
	if err == nil {
		contentLen = info.Size()
	}
	return contentLen
}
