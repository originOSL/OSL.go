// name: zip
// description: ZIP compression and archive utilities
// author: roturbot
// requires: archive/zip as gozip, archive/tar, compress/gzip, io, os, path/filepath, strings, encoding/json, fmt

type Zip struct{}

func (Zip) compress(sourcePath any, outputPath any) bool {
	source := OSLtoString(sourcePath)
	output := OSLtoString(outputPath)

	zipFile, err := os.Create(output)
	if err != nil {
		return false
	}

	zipWriter := gozip.NewWriter(zipFile)
	if zipWriter == nil {
		zipFile.Close()
		return false
	}

	success := zip.addFileToZip(zipWriter, source, "")
	zipWriter.Close()
	zipFile.Close()

	return success
}

func (Zip) addFileToZip(zipWriter *gozip.Writer, basePath string, baseInZip string) bool {
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return false
	}

	baseInZip = filepath.Join(baseInZip, filepath.Base(basePath))

	if fileInfo.IsDir() {
		zipWriter.Create(baseInZip + "/")

		entries, err := os.ReadDir(basePath)
		if err != nil {
			return false
		}

		for _, entry := range entries {
			entryPath := filepath.Join(basePath, entry.Name())
			zip.addFileToZip(zipWriter, entryPath, baseInZip+"/")
		}
		return true
	}

	if fileInfo.Mode().IsRegular() {
		file, err := os.Open(basePath)
		if err != nil {
			return false
		}
		defer file.Close()

		header, err := gozip.FileInfoHeader(fileInfo)
		if err != nil {
			return false
		}
		header.Name = baseInZip
		header.Method = gozip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return false
		}

		_, err = OSLio.Copy(writer, file)
		if err != nil {
			return false
		}
	}

	return true
}

func (Zip) decompress(zipPath any, outputPath any) bool {
	zipStr := OSLtoString(zipPath)
	output := OSLtoString(outputPath)

	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return false
	}
	defer zipFile.Close()

	for _, file := range zipFile.File {
		filePath := filepath.Join(output, file.Name)

		if file.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			continue
		}

		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			continue
		}

		_, err = OSLio.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()

		if err != nil {
			continue
		}
	}

	return true
}

func (Zip) list(zipPath any) any {
	zipStr := OSLtoString(zipPath)
	items := []any{}

	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return items
	}
	defer zipFile.Close()

	for _, file := range zipFile.File {
		info := map[string]any{
			"name":         file.Name,
			"size":         file.UncompressedSize64,
			"compressed":   file.CompressedSize64,
			"isDirectory":  file.FileInfo().IsDir(),
			"modified":     file.Modified.String(),
			"method":       file.Method,
		}
		items = append(items, info)
	}

	return items
}

func (Zip) tar(sourcePath any, outputPath any) bool {
	source := OSLtoString(sourcePath)
	output := OSLtoString(outputPath)

	tarFile, err := os.Create(output)
	if err != nil {
		return false
	}
	defer tarFile.Close()

	tarWriter := tar.NewWriter(tarFile)
	defer tarWriter.Close()

	return zip.addFileToTar(tarWriter, source, "")
}

func (Zip) addFileToTar(tarWriter *tar.Writer, basePath string, baseInTar string) bool {
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return false
	}

	baseInTar = filepath.Join(baseInTar, filepath.Base(basePath))

	if fileInfo.IsDir() {
		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return false
		}
		header.Name = baseInTar + "/"

		if err := tarWriter.WriteHeader(header); err != nil {
			return false
		}

		entries, err := os.ReadDir(basePath)
		if err != nil {
			return false
		}

		for _, entry := range entries {
			entryPath := filepath.Join(basePath, entry.Name())
			zip.addFileToTar(tarWriter, entryPath, baseInTar+"/")
		}
		return true
	}

	if fileInfo.Mode().IsRegular() {
		file, err := os.Open(basePath)
		if err != nil {
			return false
		}
		defer file.Close()

		header, err := tar.FileInfoHeader(fileInfo, "")
		if err != nil {
			return false
		}
		header.Name = baseInTar

		if err := tarWriter.WriteHeader(header); err != nil {
			return false
		}

		OSLio.Copy(tarWriter, file)
	}

	return true
}

func (Zip) untar(tarPath any, outputPath any) bool {
	tarStr := OSLtoString(tarPath)
	output := OSLtoString(outputPath)

	tarFile, err := os.Open(tarStr)
	if err != nil {
		return false
	}
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)

	for {
		header, err := tarReader.Next()
		if err != nil {
			break
		}

		filePath := filepath.Join(output, header.Name)

		if header.Typeflag == tar.TypeDir {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			continue
		}

		dstFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(header.Mode))
		if err != nil {
			continue
		}
		defer dstFile.Close()

		OSLio.Copy(dstFile, tarReader)
	}

	return true
}

func (Zip) gzip(sourcePath any, outputPath any) bool {
	source := OSLtoString(sourcePath)
	output := OSLtoString(outputPath)

	srcFile, err := os.Open(source)
	if err != nil {
		return false
	}
	defer srcFile.Close()

	dstFile, err := os.Create(output)
	if err != nil {
		return false
	}
	defer dstFile.Close()

	gzipWriter := gzip.NewWriter(dstFile)
	defer gzipWriter.Close()

	OSLio.Copy(gzipWriter, srcFile)
	return true
}

func (Zip) gunzip(sourcePath any, outputPath any) bool {
	source := OSLtoString(sourcePath)
	output := OSLtoString(outputPath)

	srcFile, err := os.Open(source)
	if err != nil {
		return false
	}
	defer srcFile.Close()

	dstFile, err := os.Create(output)
	if err != nil {
		return false
	}
	defer dstFile.Close()

	gzipReader, err := gzip.NewReader(srcFile)
	if err != nil {
		return false
	}
	defer gzipReader.Close()

	OSLio.Copy(dstFile, gzipReader)
	return true
}

func (Zip) compressString(data any) string {
	dataStr := OSLtoString(data)
	
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	gzipWriter.Write([]byte(dataStr))
	gzipWriter.Close()
	
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return base64Str
}

func (Zip) decompressString(data any) string {
	dataStr := OSLtoString(data)
	
	decoded, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return ""
	}
	
	buf := bytes.NewBuffer(decoded)
	gzipReader, err := gzip.NewReader(buf)
	if err != nil {
		return ""
	}
	defer gzipReader.Close()
	
	decompressed, err := OSLio.ReadAll(gzipReader)
	if err != nil {
		return ""
	}
	
	return string(decompressed)
}

func (Zip) fileInfo(zipPath any, filePath any) any {
	zipStr := OSLtoString(zipPath)
	fileStr := OSLtoString(filePath)
	
	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return nil
	}
	defer zipFile.Close()
	
	for _, file := range zipFile.File {
		if file.Name == fileStr {
			info := map[string]any{
				"name":         file.Name,
				"size":         file.UncompressedSize64,
				"compressed":   file.CompressedSize64,
				"isDirectory":  file.FileInfo().IsDir(),
				"modified":     file.Modified.String(),
				"method":       file.Method,
				"mode":         file.Mode(),
				"comment":      file.Comment,
			}
			return info
		}
	}
	
	return nil
}

func (Zip) extractFile(zipPath any, filePath any, outputPath any) bool {
	zipStr := OSLtoString(zipPath)
	fileStr := OSLtoString(filePath)
	output := OSLtoString(outputPath)
	
	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return false
	}
	defer zipFile.Close()
	
	for _, file := range zipFile.File {
		if file.Name == fileStr {
			dstFile, err := os.Create(output)
			if err != nil {
				return false
			}
			defer dstFile.Close()
			
			srcFile, err := file.Open()
			if err != nil {
				return false
			}
			defer srcFile.Close()
			
			OSLio.Copy(dstFile, srcFile)
			return true
		}
	}
	
	return false
}

func (Zip) addFile(zipPath any, filePath any) bool {
	return false
}

func (Zip) removeFile(zipPath any, filePath any) bool {
	
	zipStr := OSLtoString(zipPath)
	fileStr := OSLtoString(filePath)
	
	tempOutput := zipStr + ".tmp"
	
	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return false
	}
	defer zipFile.Close()
	
	newZipFile, err := os.Create(tempOutput)
	if err != nil {
		return false
	}
	defer newZipFile.Close()
	
	newZipWriter := gozip.NewWriter(newZipFile)
	defer newZipWriter.Close()
	
	for _, file := range zipFile.File {
		if file.Name == fileStr {
			continue
		}
		
		header := file.FileHeader
		newWriter, err := newZipWriter.CreateHeader(&header)
		if err != nil {
			continue
		}
		
		srcFile, err := file.Open()
		if err != nil {
			continue
		}
		
		OSLio.Copy(newWriter, srcFile)
		srcFile.Close()
	}
	
	os.Rename(tempOutput, zipStr)
	return true
}

func (Zip) copyFile(src, dst string) bool {
	srcFile, err := os.Open(src)
	if err != nil {
		return false
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return false
	}
	defer dstFile.Close()
	
	OSLio.Copy(dstFile, srcFile)
	return true
}

func (Zip) addFileToZipAtPath(zipWriter *gozip.Writer, basePath string, baseInZip string) bool {
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return false
	}
	
	baseInZip = filepath.Join(baseInZip, filepath.Base(basePath))
	
	if fileInfo.IsDir() {
		zipWriter.Create(baseInZip + "/")
		
		entries, err := os.ReadDir(basePath)
		if err != nil {
			return false
		}
		
		for _, entry := range entries {
			entryPath := filepath.Join(basePath, entry.Name())
			zip.addFileToZipAtPath(zipWriter, entryPath, baseInZip+"/")
		}
		return true
	}
	
	if fileInfo.Mode().IsRegular() {
		file, err := os.Open(basePath)
		if err != nil {
			return false
		}
		defer file.Close()
		
		header, err := gozip.FileInfoHeader(fileInfo)
		if err != nil {
			return false
		}
		header.Name = baseInZip
		header.Method = gozip.Deflate
		
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return false
		}
		
		OSLio.Copy(writer, file)
	}
	
	return true
}

func (Zip) create(path string) bool {
	file, err := os.Create(path)
	if err != nil {
		return false
	}
	defer file.Close()
	return true
}

func (Zip) statistics(zipPath any) any {
	zipStr := OSLtoString(zipPath)
	
	zipFile, err := gozip.OpenReader(zipStr)
	if err != nil {
		return nil
	}
	defer zipFile.Close()
	
	fileCount := 0
	dirCount := 0
	totalSize := int64(0)
	totalCompressed := int64(0)
	
	for _, file := range zipFile.File {
		if file.FileInfo().IsDir() {
			dirCount++
		} else {
			fileCount++
		}
		totalSize += int64(file.UncompressedSize64)
		totalCompressed += int64(file.CompressedSize64)
	}
	
	ratio := float64(0.0)
	if totalSize > 0 {
		ratio = float64(totalCompressed) / float64(totalSize)
	}
	
	return map[string]any{
		"files":         fileCount,
		"directories":   dirCount,
		"totalSize":     totalSize,
		"compressed":    totalCompressed,
		"uncompressed":  totalSize,
		"ratio":         ratio,
		"savings":       (1.0 - (float64(totalCompressed) / float64(totalSize))) * 100,
	}
}

var zip = Zip{}
