// name: img
// description: Memory-efficient dynamic image utilities for OSL
// author: Mist
// requires: bytes as OSL_bytes, image as OSL_image, image/png, image/jpeg, golang.org/x/image/draw as OSL_draw, github.com/nfnt/resize as OSL_img_resize, sync, sync/atomic, fmt, os

const (
	imgMaxDim = 4096
)

type IMG struct{}

type imgEntry struct {
	im *OSL_image.RGBA
}

var (
	OSL_img_Store = make(map[string]*imgEntry)
	OSL_img_Mu    sync.RWMutex
	OSL_img_ID    uint64
)

func OSL_allocRGBA(w, h int) *OSL_image.RGBA {
	return OSL_image.NewRGBA(OSL_image.Rect(0, 0, w, h))
}

func OSL_toRGBA(src OSL_image.Image) *OSL_image.RGBA {
	if src == nil {
		return nil
	}

	if im, ok := src.(*OSL_image.RGBA); ok {
		return im
	}

	b := src.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	OSL_draw.Draw(dst, dst.Bounds(), src, b.Min, OSL_draw.Src)
	return dst
}

func OSL_storeImage(src OSL_image.Image) string {
	im := OSL_toRGBA(src)
	if im == nil {
		return ""
	}

	OSL_img_Mu.Lock()
	defer OSL_img_Mu.Unlock()

	id := fmt.Sprintf("img_%d", atomic.AddUint64(&OSL_img_ID, 1))
	OSL_img_Store[id] = &imgEntry{im: im}
	return id
}

func OSL_getImage(id string) *OSL_image.RGBA {
	OSL_img_Mu.RLock()
	e := OSL_img_Store[id]
	OSL_img_Mu.RUnlock()

	if e == nil {
		return nil
	}
	return e.im
}

func (IMG) DecodeBytes(data []byte) string {
	im, _, err := OSL_image.Decode(OSL_bytes.NewReader(data))
	if err != nil {
		return ""
	}
	return OSL_storeImage(im)
}

func (IMG) DecodeFile(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return img.DecodeBytes(b)
}

func (IMG) EncodePNG(id any) []byte {
	im := OSL_getImage(OSLcastString(id))
	if im == nil {
		return nil
	}

	var buf OSL_bytes.Buffer
	_ = png.Encode(&buf, im)
	return buf.Bytes()
}

func (IMG) EncodeJPEG(id any, quality int) []byte {
	im := OSL_getImage(OSLcastString(id))
	if im == nil {
		return nil
	}

	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	var buf OSL_bytes.Buffer
	_ = jpeg.Encode(&buf, im, &jpeg.Options{Quality: quality})
	return buf.Bytes()
}

func (IMG) Size(id any) map[string]any {
	im := OSL_getImage(OSLcastString(id))
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
	im := OSL_getImage(OSLcastString(id))
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

func (IMG) Resize(id any, width, height int) string {
	src := OSL_getImage(OSLcastString(id))
	if src == nil {
		return ""
	}

	if width <= 0 || height <= 0 || width > imgMaxDim || height > imgMaxDim {
		return ""
	}

	m := OSL_img_resize.Resize(uint(width), uint(height), src, OSL_img_resize.Lanczos3)
	return OSL_storeImage(m)
}

func OSL_img_rotate90(src OSL_image.Image, angle int) string {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	if angle == 0 {
		return OSL_storeImage(src)
	}

	var dst *OSL_image.RGBA
	if angle == 180 {
		dst = OSL_allocRGBA(sw, sh)
	} else {
		dst = OSL_allocRGBA(sh, sw)
	}

	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			c := src.At(sb.Min.X+x, sb.Min.Y+y)

			switch angle {
			case 90:
				dst.Set(sh-1-y, x, c)
			case 180:
				dst.Set(sw-1-x, sh-1-y, c)
			case 270:
				dst.Set(y, sw-1-x, c)
			}
		}
	}

	return OSL_storeImage(dst)
}

func OSL_img_rotateAny(src OSL_image.Image, angle float64) string {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	rad := angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)

	// compute new bounds
	nw := int(math.Abs(float64(sw)*cos) + math.Abs(float64(sh)*sin))
	nh := int(math.Abs(float64(sw)*sin) + math.Abs(float64(sh)*cos))

	dst := OSL_allocRGBA(nw, nh)

	cx := float64(sw) / 2
	cy := float64(sh) / 2
	ncx := float64(nw) / 2
	ncy := float64(nh) / 2

	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			tx := float64(x) - ncx
			ty := float64(y) - ncy

			sx := tx*cos + ty*sin + cx
			sy := -tx*sin + ty*cos + cy

			if sx < 0 || sy < 0 || sx >= float64(sw) || sy >= float64(sh) {
				continue
			}

			dst.Set(x, y, src.At(
				sb.Min.X+int(sx),
				sb.Min.Y+int(sy),
			))
		}
	}

	return OSL_storeImage(dst)
}

func (IMG) Rotate(id any, angle any) string {
	src := OSL_getImage(OSLcastString(id))
	if src == nil {
		return ""
	}

	a := OSLround(OSLcastNumber(angle)) % 360
	if a < 0 {
		a += 360
	}

	if a%90 == 0 {
		return OSL_img_rotate90(src, a)
	}

	return OSL_img_rotateAny(src, float64(a))
}

func (IMG) SavePNG(id any, path any) bool {
	data := img.EncodePNG(id)
	if data == nil {
		return false
	}
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) SaveJPEG(id any, path any, quality int) bool {
	data := img.EncodeJPEG(id, quality)
	if data == nil {
		return false
	}
	return os.WriteFile(OSLcastString(path), data, 0644) == nil
}

func (IMG) Crop(id any, x, y, w, h int) string {
	src := OSL_getImage(OSLcastString(id))
	if src == nil || w <= 0 || h <= 0 {
		return ""
	}

	r := OSL_image.Rect(x, y, x+w, y+h).Intersect(src.Bounds())
	if r.Empty() {
		return ""
	}

	dst := OSL_allocRGBA(r.Dx(), r.Dy())
	OSL_draw.Draw(dst, dst.Bounds(), src, r.Min, OSL_draw.Src)
	return OSL_storeImage(dst)
}

func (IMG) Clone(id any) string {
	src := OSL_getImage(OSLcastString(id))
	if src == nil {
		return ""
	}

	b := src.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	copy(dst.Pix, src.Pix)
	return OSL_storeImage(dst)
}

func (IMG) Fill(id any, r, g, b, a uint8) bool {
	im := OSL_getImage(OSLcastString(id))
	if im == nil {
		return false
	}

	pix := im.Pix
	for i := 0; i < len(pix); i += 4 {
		pix[i+0] = r
		pix[i+1] = g
		pix[i+2] = b
		pix[i+3] = a
	}
	return true
}

func (IMG) Free(id any) {
	idStr := OSLcastString(id)
	if idStr == "" {
		return
	}

	OSL_img_Mu.Lock()
	delete(OSL_img_Store, idStr)
	OSL_img_Mu.Unlock()
}

func (IMG) FreeAll() {
	OSL_img_Mu.Lock()
	OSL_img_Store = make(map[string]*imgEntry)
	OSL_img_Mu.Unlock()
}

func (IMG) Stats() map[string]any {
	OSL_img_Mu.RLock()
	n := len(OSL_img_Store)
	OSL_img_Mu.RUnlock()

	return map[string]any{
		"count": n,
	}
}

var img = IMG{}
