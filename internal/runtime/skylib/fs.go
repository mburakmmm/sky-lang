package skylib

import (
	"io"
	"os"
	"path/filepath"
)

// Exists checks if path exists
func FSExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsFile checks if path is a file
func FSIsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsDir checks if path is a directory
func FSIsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// ReadText reads file as text
func FSReadText(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteText writes text to file
func FSWriteText(path, data string) error {
	return os.WriteFile(path, []byte(data), 0644)
}

// ReadBytes reads file as bytes
func FSReadBytes(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteBytes writes bytes to file
func FSWriteBytes(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}

// Mkdir creates a directory
func FSMkdir(path string, recursive bool) error {
	if recursive {
		return os.MkdirAll(path, 0755)
	}
	return os.Mkdir(path, 0755)
}

// Remove removes a file or directory
func FSRemove(path string, recursive bool) error {
	if recursive {
		return os.RemoveAll(path)
	}
	return os.Remove(path)
}

// Rename renames a file or directory
func FSRename(src, dst string) error {
	return os.Rename(src, dst)
}

// Copy copies a file
func FSCopy(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	return err
}

// ListDir lists directory contents
func FSListDir(path string) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	
	return names, nil
}

// Walk walks directory tree
func FSWalk(root string, fn func(string, bool) error) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		return fn(path, info.IsDir())
	})
}

// Stat returns file info
func FSStat(path string) (map[string]interface{}, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"name":    info.Name(),
		"size":    info.Size(),
		"mode":    info.Mode().String(),
		"is_dir":  info.IsDir(),
		"mod_time": info.ModTime().Unix(),
	}, nil
}

// Chmod changes file permissions
func FSChmod(path string, mode uint32) error {
	return os.Chmod(path, os.FileMode(mode))
}

// Chown changes file owner (Unix only)
func FSChown(path string, uid, gid int) error {
	return os.Chown(path, uid, gid)
}

// Symlink creates a symbolic link
func FSSymlink(target, link string) error {
	return os.Symlink(target, link)
}

// Readlink reads a symbolic link
func FSReadlink(path string) (string, error) {
	return os.Readlink(path)
}

// Abs returns absolute path
func FSAbs(path string) (string, error) {
	return filepath.Abs(path)
}

// RealPath resolves symbolic links
func FSRealPath(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

// Glob finds files matching pattern
func FSGlob(pattern string) ([]string, error) {
	return filepath.Glob(pattern)
}

