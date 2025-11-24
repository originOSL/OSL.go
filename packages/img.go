// name: img
// description: Dynamic image utilities for OSL
// author: Mist
// requires: bytes as OSL_bytes, golang.org/x/image/draw as OSL_draw, image as OSL_image, image/png, image/jpeg

type IMG struct{}

var (
	OSL_img_Store = map[string]OSL_image.Image{}
	OSL_img_Mu    sync.Mutex
)

func OSL_img_store(im OSL_image.Image) string {
	id := fmt.Sprintf("img_%d", time.Now().UnixNano())
	OSL_img_Mu.Lock()
	OSL_img_Store[id] = im
	OSL_img_Mu.Unlock()
	return id
}

func OSL_img_get(id string) OSL_image.Image {
	OSL_img_Mu.Lock()
	im := imgStore[id]
	OSL_img_Mu.Unlock()

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
		return []byte{}
	}

	var buf OSL_bytes.Buffer
	_ = png.Encode(&buf, im)
	return buf.Bytes()
}

func (IMG) EncodeJPEG(id any, quality any) []byte {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return []byte{}
	}

	v := OSLcastNumber(quality)

	var buf OSL_bytes.Buffer
	_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: v})
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

func (IMG) Resize(id any, width any, height any) string {
	im := OSL_img_get(OSLcastString(id))
	if im == nil {
		return ""
	}

	w := OSLcastNumber(width)
	h := OSLcastNumber(height)

	if w <= 0 || h <= 0 {
		return ""
	}

	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	draw.CatmullRom.Scale(dst, dst.Bounds(), im, im.Bounds(), draw.Over, nil)

	return storeImage(dst)
}

var img = IMG{}
