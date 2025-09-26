// name: fs
// description: File system utilities
// author: Mist
// requires: os, path/filepath

type FS struct{}

func (FS) ReadFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

func (FS) WriteFile(path string, data string) bool {
	err := os.WriteFile(path, []byte(data), 0644)
	return err == nil
}

func (FS) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (FS) Remove(path string) bool {
	err := os.Remove(path)
	return err == nil
}

func (FS) RemoveAll(path string) bool {
	err := os.RemoveAll(path)
	return err == nil
}

func (FS) Mkdir(path string) bool {
	err := os.Mkdir(path, 0755)
	return err == nil
}

func (FS) MkdirAll(path string) bool {
	err := os.MkdirAll(path, 0755)
	return err == nil
}

func (FS) ReadDir(path string) []string {
	files, err := os.ReadDir(path)
	if err != nil {
		return []string{}
	}
	names := make([]string, len(files))
	for i, f := range files {
		names[i] = f.Name()
	}
	return names
}

func (FS) Getwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return dir
}

func (FS) Chdir(path string) bool {
	err := os.Chdir(path)
	return err == nil
}

func (FS) JoinPath(path ...string) string {
	return filepath.Join(path...)
}

func (FS) getBase(path string) string {
	return filepath.Base(path)
}

func (FS) getDir(path string) string {
	return filepath.Dir(path)
}

func (FS) getExt(path string) string {
	return filepath.Ext(path)
}

func (FS) getParts(path string) []string {
	return []string{filepath.Base(path), filepath.Dir(path), filepath.Ext(path)}
}

// Global instance
var fs = FS{}
