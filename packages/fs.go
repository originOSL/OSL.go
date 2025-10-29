// name: fs
// description: File system utilities
// author: Mist
// requires: os, path/filepath

type FS struct{}

func (FS) ReadFile(path any) string {
	data, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return ""
	}
	return string(data)
}

func (FS) ReadFileBytes(path any) []byte {
	data, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return []byte{}
	}
	return data
}

func (FS) WriteFile(path any, data any) bool {
	err := os.WriteFile(OSLcastString(path), []byte(OSLcastString(data)), 0644)
	return err == nil
}

func (FS) Rename(oldPath any, newPath any) bool {
	err := os.Rename(OSLcastString(oldPath), OSLcastString(newPath))
	return err == nil
}

func (FS) Exists(path any) bool {
	_, err := os.Stat(OSLcastString(path))
	return err == nil
}

func (FS) Remove(path any) bool {
	pathStr := OSLcastString(path)
	if pathStr == "" {
		return false
	}
	if !FS.Exists(pathStr) {
		return false
	}

	if FS.IsDir(pathStr) {
		return os.RemoveAll(pathStr) == nil
	}

	return os.Remove(pathStr) == nil
}

func (FS) Mkdir(path any) bool {
	err := os.Mkdir(OSLcastString(path), 0755)
	return err == nil
}

func (FS) MkdirAll(path any) bool {
	err := os.MkdirAll(OSLcastString(path), 0755)
	return err == nil
}

func (FS) CopyDir(srcPath any, dstPath any) bool {
	src := OSLcastString(srcPath)
	dst := OSLcastString(dstPath)

	entries, err := os.ReadDir(src)
	if err != nil {
		return false
	}

	if err := os.MkdirAll(dst, 0755); err != nil {
		return false
	}

	for _, entry := range entries {
		srcFile := filepath.Join(src, entry.Name())
		dstFile := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			ok := (FS{}).CopyDir(srcFile, dstFile)
			if !ok {
				return false
			}
			continue
		}

		in, err := os.Open(srcFile)
		if err != nil {
			return false
		}

		out, err := os.Create(dstFile)
		if err != nil {
			in.Close()
			return false
		}

		if _, err := io.Copy(out, in); err != nil {
			in.Close()
			out.Close()
			return false
		}

		in.Close()
		out.Close()

		if info, err := os.Stat(srcFile); err == nil {
			_ = os.Chmod(dstFile, info.Mode())
		}
	}

	return true
}

func (FS) ReadDir(path any) []any {
	files, err := os.ReadDir(OSLcastString(path))
	if err != nil {
		return []any{}
	}
	names := make([]any, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	return names
}

func (FS) ReadDirAll(path any) []map[string]any {
	dir := OSLcastString(path)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return []map[string]any{}
	}

	filesOut := make([]map[string]any, len(entries))
	for i, f := range entries {
		filesOut[i] = map[string]any{
			"name":  f.Name(),
			"ext":   filepath.Ext(f.Name()),
			"path":  filepath.Join(dir, f.Name()),
			"isDir": f.IsDir(),
			"type":  f.Type(),
		}
	}

	return filesOut
}

func (FS) WalkDir(path any, fn func(path string, file map[string]any, control map[string]any)) {
	dir := OSLcastString(path)
	filepath.WalkDir(dir, func(p string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		fileData := map[string]any{
			"name":    entry.Name(),
			"ext":     filepath.Ext(entry.Name()),
			"path":    p,
			"isDir":   entry.IsDir(),
			"size":    info.Size(),
			"mode":    info.Mode(),
			"modTime": info.ModTime(),
			"sys":     info.Sys(),
			"type":    entry.Type(),
		}

		control := map[string]any{
			"skip": false,
		}
		fn(p, fileData, control)
		if control["skip"] == true {
			return filepath.SkipDir
		}
		return nil
	})
}

func (FS) IsDir(path any) bool {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return false
	}
	return info.IsDir()
}

func (FS) Getwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (FS) Chdir(path any) bool {
	err := os.Chdir(OSLcastString(path))
	return err == nil
}

func (FS) JoinPath(path ...any) string {
	stringPath := make([]string, len(path))
	for i, p := range path {
		stringPath[i] = OSLcastString(p)
	}
	return filepath.Join(stringPath...)
}

func (FS) GetBase(path any) string {
	return filepath.Base(OSLcastString(path))
}

func (FS) GetDir(path any) string {
	return filepath.Dir(OSLcastString(path))
}

func (FS) GetExt(path any) string {
	return filepath.Ext(OSLcastString(path))
}

func (FS) GetParts(path any) []any {
	stringPath := OSLcastString(path)
	return []any{filepath.Base(stringPath), filepath.Dir(stringPath), filepath.Ext(stringPath)}
}

func (FS) GetSize(path any) float64 {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return 0
	}
	return float64(info.Size())
}

func (FS) GetModTime(path any) float64 {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return 0.0
	}
	return float64(info.ModTime().UnixMilli())
}

func (FS) GetStat(path any) map[string]any {
	info, err := os.Stat(OSLcastString(path))
	if err != nil {
		return map[string]any{"success": false}
	}
	return map[string]any{
		"success": true,
		"name":    filepath.Base(info.Name()),
		"ext":     filepath.Ext(info.Name()),
		"path":    info.Name(),
		"isDir":   info.IsDir(),
		"size":    info.Size(),
		"mode":    info.Mode(),
		"modTime": info.ModTime().UnixMicro(),
		"sys":     info.Sys(),
	}
}

// Global instance
var fs = FS{}
