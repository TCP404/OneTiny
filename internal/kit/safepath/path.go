package safepath

import (
	"os"
	"path/filepath"
	"strings"
)

func ResolveWithinRoot(rootPath string, elems ...string) (string, bool) {
	cleanRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", false
	}

	parts := append([]string{cleanRoot}, elems...)
	target, err := filepath.Abs(filepath.Join(parts...))
	if err != nil {
		return "", false
	}

	rel, err := filepath.Rel(cleanRoot, target)
	if err != nil {
		return "", false
	}
	if !isContainedRel(rel) {
		return "", false
	}
	return target, true
}

func ResolveExistingWithinRoot(rootPath string, elems ...string) (string, bool) {
	target, ok := ResolveWithinRoot(rootPath, elems...)
	if !ok {
		return "", false
	}

	evalRoot, ok := evalRootPath(rootPath)
	if !ok {
		return "", false
	}
	evalTarget, err := filepath.EvalSymlinks(target)
	if err != nil {
		return "", false
	}
	evalTarget, err = filepath.Abs(evalTarget)
	if err != nil {
		return "", false
	}
	if !isWithinRoot(evalRoot, evalTarget) {
		return "", false
	}
	return target, true
}

func ResolveCreateWithinRoot(rootPath string, elems ...string) (string, bool) {
	target, ok := ResolveWithinRoot(rootPath, elems...)
	if !ok {
		return "", false
	}

	evalRoot, ok := evalRootPath(rootPath)
	if !ok {
		return "", false
	}
	parent, err := filepath.Abs(filepath.Dir(target))
	if err != nil {
		return "", false
	}
	evalParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", false
	}
	evalParent, err = filepath.Abs(evalParent)
	if err != nil {
		return "", false
	}
	if !isWithinRoot(evalRoot, evalParent) {
		return "", false
	}

	info, err := os.Lstat(target)
	if err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", false
		}
		return target, true
	}
	if !os.IsNotExist(err) {
		return "", false
	}
	return target, true
}

func evalRootPath(rootPath string) (string, bool) {
	cleanRoot, err := filepath.Abs(rootPath)
	if err != nil {
		return "", false
	}
	evalRoot, err := filepath.EvalSymlinks(cleanRoot)
	if err != nil {
		return "", false
	}
	evalRoot, err = filepath.Abs(evalRoot)
	if err != nil {
		return "", false
	}
	return evalRoot, true
}

func isWithinRoot(rootPath, targetPath string) bool {
	rel, err := filepath.Rel(rootPath, targetPath)
	if err != nil {
		return false
	}
	return isContainedRel(rel)
}

func isContainedRel(rel string) bool {
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
