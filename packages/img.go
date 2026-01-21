// name: img
// description: Memory-efficient image utilities following Go idioms
// author: Mist
// requires: bytes as OSL_bytes, image as OSL_image, image/png, image/jpeg, golang.org/x/image/draw as OSL_draw, github.com/nfnt/resize as OSL_img_resize, os, io, runtime

const (
	imgMaxDim = 16384
)

type IMG struct{}

type OSL_img_Image struct {
	im     *OSL_image.RGBA
	closed bool
}

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

func OSL_newImage(im *OSL_image.RGBA) *OSL_img_Image {
	if im == nil {
		return nil
	}

	img := &OSL_img_Image{im: im}

	runtime.SetFinalizer(img, func(i *OSL_img_Image) {
		i.Close()
	})

	return img
}

func (i *OSL_img_Image) Close() {
	if i == nil || i.closed {
		return
	}

	i.closed = true
	i.im = nil
}

func (i *OSL_img_Image) Size() map[string]any {
	if i == nil || i.closed || i.im == nil {
		return map[string]any{}
	}
	b := i.im.Bounds()
	h := b.Dy()
	w := b.Dx()
	return map[string]any{
		"w": w,
		"h": h,
	}
}

func (i *OSL_img_Image) Width() int {
	if i == nil || i.closed || i.im == nil {
		return 0
	}
	return i.im.Bounds().Dx()
}

func (i *OSL_img_Image) Height() int {
	if i == nil || i.closed || i.im == nil {
		return 0
	}
	return i.im.Bounds().Dy()
}

func (i *OSL_img_Image) RGBA() *OSL_image.RGBA {
	if i == nil || i.closed {
		return nil
	}
	return i.im
}

func (IMG) Open(path string) *OSL_img_Image {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	im, _, err := OSL_image.Decode(f)
	if err != nil {
		return nil
	}

	return OSL_newImage(OSL_toRGBA(im))
}

func (IMG) Decode(r io.Reader) *OSL_img_Image {
	im, _, err := OSL_image.Decode(r)
	if err != nil {
		return nil
	}

	return OSL_newImage(OSL_toRGBA(im))
}

func (IMG) DecodeBytes(data []byte) *OSL_img_Image {
	return img.Decode(OSL_bytes.NewReader(data))
}

func (IMG) EncodePNG(w io.Writer, i *OSL_img_Image) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}
	return png.Encode(w, i.im) == nil
}

func (IMG) EncodeJPEG(w io.Writer, i *OSL_img_Image, quality int) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	return jpeg.Encode(w, i.im, &jpeg.Options{Quality: quality}) == nil
}

func (IMG) EncodePNGBytes(i *OSL_img_Image) []byte {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	var buf OSL_bytes.Buffer
	if png.Encode(&buf, i.im) != nil {
		return nil
	}
	return buf.Bytes()
}

func (IMG) EncodeJPEGBytes(i *OSL_img_Image, quality int) []byte {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	if quality < 1 {
		quality = 1
	}
	if quality > 100 {
		quality = 100
	}

	var buf OSL_bytes.Buffer
	if jpeg.Encode(&buf, i.im, &jpeg.Options{Quality: quality}) != nil {
		return nil
	}
	return buf.Bytes()
}

func (IMG) SavePNG(i *OSL_img_Image, path string) bool {
	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return img.EncodePNG(f, i)
}

func (IMG) SaveJPEG(i *OSL_img_Image, path string, quality int) bool {
	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return img.EncodeJPEG(f, i, quality)
}

func (IMG) Size(i *OSL_img_Image) map[string]any {
	if i == nil || i.closed || i.im == nil {
		return map[string]any{}
	}
	return i.Size()
}

func (IMG) Resize(i *OSL_img_Image, width, height int) *OSL_img_Image {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	if width <= 0 || height <= 0 || width > imgMaxDim || height > imgMaxDim {
		return nil
	}

	resized := OSL_img_resize.Resize(uint(width), uint(height), i.im, OSL_img_resize.Lanczos3)
	return OSL_newImage(OSL_toRGBA(resized))
}

func (IMG) ResizeFast(i *OSL_img_Image, width, height int) *OSL_img_Image {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	if width <= 0 || height <= 0 || width > imgMaxDim || height > imgMaxDim {
		return nil
	}

	resized := OSL_img_resize.Resize(uint(width), uint(height), i.im, OSL_img_resize.Bilinear)
	return OSL_newImage(OSL_toRGBA(resized))
}

func (IMG) Rotate(i *OSL_img_Image, angle float64) *OSL_img_Image {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	a := int(angle) % 360
	if a < 0 {
		a += 360
	}

	if a%90 == 0 {
		return img.rotate90(i.im, a)
	}

	return img.rotateAny(i.im, angle)
}

func (IMG) rotate90(src *OSL_image.RGBA, angle int) *OSL_img_Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	if angle == 0 {
		return OSL_newImage(src)
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

	return OSL_newImage(dst)
}

func (IMG) rotateAny(src *OSL_image.RGBA, angle float64) *OSL_img_Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	rad := angle * math.Pi / 180
	sin, cos := math.Sin(rad), math.Cos(rad)

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

	return OSL_newImage(dst)
}

func (IMG) Crop(i *OSL_img_Image, x, y, w, h int) *OSL_img_Image {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	if w <= 0 || h <= 0 {
		return nil
	}

	r := OSL_image.Rect(x, y, x+w, y+h).Intersect(i.im.Bounds())
	if r.Empty() {
		return nil
	}

	dst := OSL_allocRGBA(r.Dx(), r.Dy())
	OSL_draw.Draw(dst, dst.Bounds(), i.im, r.Min, OSL_draw.Src)
	return OSL_newImage(dst)
}

func (IMG) Clone(i *OSL_img_Image) *OSL_img_Image {
	if i == nil || i.closed || i.im == nil {
		return nil
	}

	b := i.im.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	copy(dst.Pix, i.im.Pix)
	return OSL_newImage(dst)
}

func (IMG) Fill(i *OSL_img_Image, r, g, b, a uint8) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	pix := i.im.Pix
	for j := 0; j < len(pix); j += 4 {
		pix[j+0] = r
		pix[j+1] = g
		pix[j+2] = b
		pix[j+3] = a
	}
	return true
}

func (IMG) New(width, height int) *OSL_img_Image {
	if width <= 0 || height <= 0 || width > imgMaxDim || height > imgMaxDim {
		return nil
	}

	return OSL_newImage(OSL_allocRGBA(width, height))
}

var img = IMG{}