package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"shared/globals"
	"shared/repository"
	"shared/shell"
	"strings"
)

type FileInfo struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	IsDirectory bool   `json:"is_directory"`
	Size        int64  `json:"size"`
	Modified    int64  `json:"modified"`
}

// ValidatePathAccess checks if a path is allowed to be accessed (read or written)
func ValidatePathAccess(targetPath string, writeAccess bool) (string, error) {
	if targetPath == "" {
		return "", fmt.Errorf("path is empty")
	}

	// Clean path and get absolute representation
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		absPath = filepath.Clean(targetPath)
	}

	// Allowed write directories
	allowedWriteDirs := []string{
		"/var/www",
		"/home",
		"/etc/nginx",
		"/etc/caddy",
		"/opt/deployberry",
		"/var/lib/deployberry",
		"/var/log/deployberry",
		"/etc/deployberry",
	}

	// Add dev mode paths
	if globals.IsDevelopment() {
		allowedWriteDirs = append(allowedWriteDirs, ".dev")
		if cwd, err := os.Getwd(); err == nil {
			allowedWriteDirs = append(allowedWriteDirs, cwd)
		}
	}

	// If write access is requested, check if it's inside allowed write directories
	if writeAccess {
		allowed := false
		for _, dir := range allowedWriteDirs {
			absDir, err := filepath.Abs(dir)
			if err != nil {
				absDir = filepath.Clean(dir)
			}
			rel, err := filepath.Rel(absDir, absPath)
			if err == nil && !strings.HasPrefix(rel, "..") {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("permission denied: path %s is not in allowed write directories", targetPath)
		}
		return absPath, nil
	}

	// Read access validation: allow reading except for highly sensitive system config files
	sensitivePatterns := []string{
		"/etc/shadow",
		"/etc/gshadow",
		"id_rsa",
		"id_dsa",
		"id_ecdsa",
		"id_ed25519",
	}
	for _, pattern := range sensitivePatterns {
		if strings.Contains(absPath, pattern) {
			return "", fmt.Errorf("permission denied: access to sensitive file %s is blocked", targetPath)
		}
	}

	return absPath, nil
}

func ListFiles(path string) ([]FileInfo, error) {
	if path == "" {
		path = "/"
	}
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	var fileList []FileInfo
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}
		fileList = append(fileList, FileInfo{
			Name:        f.Name(),
			Path:        filepath.Join(path, f.Name()),
			IsDirectory: f.IsDir(),
			Size:        info.Size(),
			Modified:    info.ModTime().Unix(),
		})
	}
	return fileList, nil
}

func DeleteFile(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	return os.RemoveAll(path)
}

func CreateDirectory(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	if err := os.Mkdir(path, 0755); err != nil {
		return err
	}
	EnsureCorrectFileOwnership(path)
	return nil
}

func copyFile(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}

	return os.Remove(src)
}

func copyDir(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return os.RemoveAll(src)
}

func copyAndDelete(src, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func RenameFile(path, newName, newPath, newOwner, newGroup string) error {
	if path == "" || newName == "" || newPath == "" {
		return fmt.Errorf("path, new name, and new path are required")
	}
	targetPath := filepath.Join(newPath, newName)
	if err := os.Rename(path, targetPath); err != nil {
		// Fallback to copy-and-delete if rename fails (e.g. cross-device link)
		if err = copyAndDelete(path, targetPath); err != nil {
			return err
		}
	}
	if newOwner != "" || newGroup != "" {
		owner := newOwner
		if owner == "" {
			owner = "root"
		}
		group := newGroup
		if group == "" {
			group = "root"
		}
		shell.ExecuteCommand(fmt.Sprintf("chown %s:%s %s", owner, group, shell.EscapeShellArg(targetPath)))
	} else {
		EnsureCorrectFileOwnership(targetPath)
	}
	return nil
}

func EnsureCorrectFileOwnership(path string) {
	db := repository.GetDB()
	var apps []repository.Application
	if err := db.Find(&apps).Error; err == nil {
		for _, app := range apps {
			if app.Path != "" && (strings.HasPrefix(path, app.Path) || path == app.Path) {
				shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(path)))
				return
			}
		}
	}
	if strings.HasPrefix(path, "/home/panel_apps") {
		shell.ExecuteCommand(fmt.Sprintf("chown -R panel_apps:www-data %s", shell.EscapeShellArg(path)))
	}
}
