// name: img
// description: Dynamic image utilities for OSL
// author: Mist
// requires: bytes as OSL_bytes, image as OSL_image, image/png, image/jpeg, sync, time, os, fmt

type IMG struct{}

var (
	imgStore = map[string]OSL_image.Image{}
	imgMu    sync.Mutex
)

func OSL_img_store(im OSL_image.Image) string {
	id := fmt.Sprintf("img_%d", time.Now().UnixNano())
	imgMu.Lock()
	imgStore[id] = im
	imgMu.Unlock()
	return id
}

func OSL_img_get(id string) OSL_image.Image {
	imgMu.Lock()
	im := imgStore[id]
	imgMu.Unlock()

	return im
}

func (IMG) GetImage(id string) OSL_image.Image {
	return OSL_img_get(id)
}

func (IMG) DecodeBytes(data []byte) string {
	b, ok := data
	if !ok {
		return ""
	}

	im, _, err := OSL_image.Decode(OSL_bytes.NewReader(b))
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
		return []byte{}
	}

	var buf OSL_bytes.Buffer
	_ = png.Encode(&buf, im)
	return buf.Bytes()
}

func (IMG) EncodeJPEG(id any, quality any) []byte {
	im := OSL_img_get(id)
	if im == nil {
		return []byte{}
	}

	q := 80
	v := OSLcastNumber(quality)

	var buf OSL_bytes.Buffer
	_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: q})
	return buf.Bytes()
}

func (IMG) Size(id any) map[string]any {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return map[string]any{"w": 0, "h": 0}
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
		return map[string]any{}
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

var img = IMG{}
