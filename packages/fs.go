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
	err := os.Remove(OSLcastString(path))
	return err == nil
}

func (FS) RemoveAll(path any) bool {
	err := os.RemoveAll(OSLcastString(path))
	return err == nil
}

func (FS) Mkdir(path any) bool {
	err := os.Mkdir(OSLcastString(path), 0755)
	return err == nil
}

func (FS) MkdirAll(path any) bool {
	err := os.MkdirAll(OSLcastString(path), 0755)
	return err == nil
}

func (FS) ReadDir(path any) []string {
	files, err := os.ReadDir(OSLcastString(path))
	if err != nil {
		return []string{}
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	return names
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

func (FS) GetParts(path any) []string {
	stringPath := OSLcastString(path)
	return []string{filepath.Base(stringPath), filepath.Dir(stringPath), filepath.Ext(stringPath)}
}

// Global instance
var fs = FS{}
