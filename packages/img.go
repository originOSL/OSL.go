// name: img
// description: Dynamic image utilities for OSL
// author: Mist
// requires: bytes as OSL_bytes, golang.org/x/image/draw as OSL_draw, image as OSL_image, image/png, image/jpeg, sync, sync/atomic

type IMG struct{}

var (
	OSL_img_Store = map[string]OSL_image.Image{}
	OSL_img_ID    uint64
	OSL_img_Mu    sync.RWMutex
)

const (
	OSL_img_Max    = 1024
	OSL_img_MaxDim = 4096
)

func OSL_img_store(im OSL_image.Image) string {
	if im == nil {
		return ""
	}

	OSL_img_Mu.RLock()
	if len(OSL_img_Store) >= OSL_img_Max {
		OSL_img_Mu.RUnlock()
		return ""
	}
	OSL_img_Mu.RUnlock()

	id := fmt.Sprintf("img_%d", atomic.AddUint64(&OSL_img_ID, 1))

	OSL_img_Mu.Lock()
	OSL_img_Store[id] = im
	OSL_img_Mu.Unlock()
	return id
}

func OSL_img_get(id string) OSL_image.Image {
	if id == "" {
		return nil
	}

	OSL_img_Mu.RLock()
	im := OSL_img_Store[id]
	OSL_img_Mu.RUnlock()
	return im
}

func (IMG) GetImage(id string) OSL_image.Image {
	return OSL_img_get(id)
}

func (IMG) UseImage(imgage OSL_image.Image) string {
	return OSL_img_store(imgage)
}

func (IMG) DecodeBytes(data []byte) string {
	im, _, err := OSL_image.Decode(OSL_bytes.NewReader(data))
	if err != nil {
		return ""
	}

	return OSL_img_store(im)
}

func (IMG) DecodeFile(path any) string {
	b, err := os.ReadFile(OSLcastString(path))
	if err != nil {
		return ""
	}
	return img.DecodeBytes(b)
}

func (IMG) EncodePNG(id any) []byte {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return nil
	}

	var buf OSL_bytes.Buffer
	_ = png.Encode(&buf, im)
	return buf.Bytes()
}

func (IMG) EncodeJPEG(id any, quality any) []byte {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return nil
	}

	v := int(OSLcastNumber(quality))

	var buf OSL_bytes.Buffer
	_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: v})
	return buf.Bytes()
}

func (IMG) Size(id any) map[string]any {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return nil
	}

	b := im.Bounds()
	return map[string]any{
		"w": b.Dx(),
		"h": b.Dy(),
	}
}

func (IMG) Bounds(id any) map[string]any {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return nil
	}

	b := im.Bounds()
	return map[string]any{
		"minX": b.Min.X,
		"minY": b.Min.Y,
		"maxX": b.Max.X,
		"maxY": b.Max.Y,
	}
}

func (IMG) SavePNG(id any, path any) bool {
	data := img.EncodePNG(OSLcastString(id))
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) SaveJPEG(id any, path any, quality any) bool {
	data := img.EncodeJPEG(OSLcastString(id), OSLcastNumber(quality))
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) Resize(id any, width any, height any) string {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return ""
	}

	w := int(OSLcastNumber(width))
	h := int(OSLcastNumber(height))

	if w <= 0 || h <= 0 || w > OSL_img_MaxDim || h > OSL_img_MaxDim {
		return ""
	}

	dst := OSL_image.NewRGBA(OSL_image.Rect(0, 0, w, h))
	OSL_draw.CatmullRom.Scale(dst, dst.Bounds(), im, im.Bounds(), OSL_draw.Over, nil)

	return OSL_img_store(dst)
}

func (IMG) Free(id any) {
	OSL_img_Mu.Lock()
	delete(OSL_img_Store, OSLcastString(id))
	OSL_img_Mu.Unlock()
}

func (IMG) FreeAll() {
	OSL_img_Mu.Lock()
	OSL_img_Store = map[string]OSL_image.Image{}
	OSL_img_Mu.Unlock()
}

func (IMG) Stats() map[string]any {
	OSL_img_Mu.RLock()
	n := len(OSL_img_Store)
	OSL_img_Mu.RUnlock()

	return map[string]any{
		"count": n,
		"max":   OSL_img_Max,
	}
}

var img = IMG{}
