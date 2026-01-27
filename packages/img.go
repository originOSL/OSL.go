// name: img
// description: Memory-efficient image utilities following Go idioms
// author: Mist
// requires: bytes as OSL_bytes, image as OSL_image, image/png, image/jpeg, golang.org/x/image/draw as OSL_draw, github.com/nfnt/resize as OSL_img_resize, os, io, runtime, github.com/rwcarlsen/goexif/exif as OSL_exif

type IMG struct{}

type OSL_img_Image struct {
	im     *OSL_image.RGBA
	closed bool
}

/* -------------------- helpers -------------------- */

func OSL_allocRGBA(w, h int) *OSL_image.RGBA {
	return OSL_image.NewRGBA(OSL_image.Rect(0, 0, w, h))
}

func OSL_toRGBA(src OSL_image.Image) *OSL_image.RGBA {
	if src == nil {
		return nil
	}

	if im, ok := src.(*OSL_image.RGBA); ok {
		dst := OSL_allocRGBA(im.Bounds().Dx(), im.Bounds().Dy())
		copy(dst.Pix, im.Pix)
		return dst
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
	return &OSL_img_Image{im: im}
}

/* -------------------- lifecycle -------------------- */

func (i *OSL_img_Image) Close() {
	if i == nil || i.closed {
		return
	}
	i.closed = true
	i.im = nil
}

func (i *OSL_img_Image) RGBA() *OSL_image.RGBA {
	if i == nil || i.closed {
		return nil
	}
	return i.im
}

/* -------------------- metadata -------------------- */

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

func (i *OSL_img_Image) Size() map[string]any {
	if i == nil || i.closed || i.im == nil {
		return map[string]any{}
	}
	b := i.im.Bounds()
	return map[string]any{"w": b.Dx(), "h": b.Dy()}
}

/* -------------------- decode helpers -------------------- */

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

func (IMG) OpenSize(path string) (int, int) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0
	}
	defer f.Close()

	cfg, _, err := OSL_image.DecodeConfig(f)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

func (IMG) DecodeSize(r io.Reader) (int, int) {
	cfg, _, err := OSL_image.DecodeConfig(r)
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

/* -------------------- encode -------------------- */

func (IMG) EncodePNG(w io.Writer, i *OSL_img_Image) bool {
	return i != nil && !i.closed && png.Encode(w, i.im) == nil
}

func (IMG) EncodeJPEG(w io.Writer, i *OSL_img_Image, q int) bool {
	if i == nil || i.closed {
		return false
	}
	if q < 1 {
		q = 1
	} else if q > 100 {
		q = 100
	}
	return jpeg.Encode(w, i.im, &jpeg.Options{Quality: q}) == nil
}

func (IMG) EncodePNGBytes(i *OSL_img_Image) []byte {
	if i == nil || i.closed {
		return nil
	}
	var buf OSL_bytes.Buffer
	if png.Encode(&buf, i.im) != nil {
		return nil
	}
	return buf.Bytes()
}

func (IMG) EncodeJPEGBytes(i *OSL_img_Image, q int) []byte {
	if i == nil || i.closed {
		return nil
	}
	if q < 1 {
		q = 1
	} else if q > 100 {
		q = 100
	}
	var buf OSL_bytes.Buffer
	if jpeg.Encode(&buf, i.im, &jpeg.Options{Quality: q}) != nil {
		return nil
	}
	return buf.Bytes()
}

/* -------------------- creation -------------------- */

func (IMG) New(w, h int) *OSL_img_Image {
	if w <= 0 || h <= 0 {
		return nil
	}
	return OSL_newImage(OSL_allocRGBA(w, h))
}

func (IMG) Clone(i *OSL_img_Image) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}
	b := i.im.Bounds()
	dst := OSL_allocRGBA(b.Dx(), b.Dy())
	copy(dst.Pix, i.im.Pix)
	return OSL_newImage(dst)
}

/* -------------------- resize helpers -------------------- */

func (IMG) Resize(i *OSL_img_Image, w, h int) *OSL_img_Image {
	if i == nil || i.closed || (w == 0 && h == 0) || w < 0 || h < 0 {
		return nil
	}
	r := OSL_img_resize.Resize(uint(w), uint(h), i.im, OSL_img_resize.Lanczos3)
	return OSL_newImage(OSL_toRGBA(r))
}

func (IMG) ResizeFast(i *OSL_img_Image, w, h int) *OSL_img_Image {
	if i == nil || i.closed || (w == 0 && h == 0) || w < 0 || h < 0 {
		return nil
	}
	r := OSL_img_resize.Resize(uint(w), uint(h), i.im, OSL_img_resize.Bilinear)
	return OSL_newImage(OSL_toRGBA(r))
}

func (IMG) ResizeWidth(i *OSL_img_Image, w int) *OSL_img_Image {
	return img.Resize(i, w, 0)
}

func (IMG) ResizeHeight(i *OSL_img_Image, h int) *OSL_img_Image {
	return img.Resize(i, 0, h)
}

func (IMG) ResizeFit(i *OSL_img_Image, maxW, maxH int) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}
	sw, sh := i.Width(), i.Height()
	rw := float64(maxW) / float64(sw)
	rh := float64(maxH) / float64(sh)
	scale := math.Min(rw, rh)
	return img.Resize(i, int(float64(sw)*scale), int(float64(sh)*scale))
}

/* -------------------- draw / composite -------------------- */

func (IMG) Draw(dst, src *OSL_img_Image, x, y int) bool {
	if dst == nil || src == nil || dst.closed || src.closed {
		return false
	}
	r := OSL_image.Rect(x, y, x+src.Width(), y+src.Height())
	OSL_draw.Draw(dst.im, r, src.im, OSL_image.Point{}, OSL_draw.Src)
	return true
}

func (IMG) DrawOver(dst, src *OSL_img_Image, x, y int) bool {
	if dst == nil || src == nil || dst.closed || src.closed {
		return false
	}
	r := OSL_image.Rect(x, y, x+src.Width(), y+src.Height())
	OSL_draw.Draw(dst.im, r, src.im, OSL_image.Point{}, OSL_draw.Over)
	return true
}

/* -------------------- rotation -------------------- */

func (IMG) Rotate(i *OSL_img_Image, angle float64) *OSL_img_Image {
	if i == nil || i.closed {
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

func (IMG) rotate90(src *OSL_image.RGBA, a int) *OSL_img_Image {
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()

	if a == 0 {
		dst := OSL_allocRGBA(sw, sh)
		copy(dst.Pix, src.Pix)
		return OSL_newImage(dst)
	}

	var dst *OSL_image.RGBA
	if a == 180 {
		dst = OSL_allocRGBA(sw, sh)
	} else {
		dst = OSL_allocRGBA(sh, sw)
	}

	for y := 0; y < sh; y++ {
		for x := 0; x < sw; x++ {
			c := src.At(x, y)
			switch a {
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

	cx, cy := float64(sw)/2, float64(sh)/2
	ncx, ncy := float64(nw)/2, float64(nh)/2

	for y := 0; y < nh; y++ {
		for x := 0; x < nw; x++ {
			tx := float64(x) - ncx
			ty := float64(y) - ncy

			sx := tx*cos + ty*sin + cx
			sy := -tx*sin + ty*cos + cy

			ix := int(math.Round(sx))
			iy := int(math.Round(sy))

			if ix >= 0 && iy >= 0 && ix < sw && iy < sh {
				dst.Set(x, y, src.At(ix, iy))
			}
		}
	}
	return OSL_newImage(dst)
}

/* -------------------- exif orientation -------------------- */

func (IMG) NormalizeOrientation(i *OSL_img_Image, r io.Reader) *OSL_img_Image {
	if i == nil || i.closed {
		return nil
	}

	ex, err := OSL_exif.Decode(r)
	if err != nil {
		return i
	}

	tag, err := ex.Get(OSL_exif.Orientation)
	if err != nil {
		return i
	}

	o, _ := tag.Int(0)

	switch o {
	case 3:
		return img.Rotate(i, 180)
	case 6:
		return img.Rotate(i, 90)
	case 8:
		return img.Rotate(i, 270)
	default:
		return i
	}
}

/* -------------------- fill / color helpers -------------------- */

func (IMG) Fill(i *OSL_img_Image, r, g, b, a uint8) bool {
	if i == nil || i.closed {
		return false
	}
	p := i.im.Pix
	for j := 0; j < len(p); j += 4 {
		p[j+0] = r
		p[j+1] = g
		p[j+2] = b
		p[j+3] = a
	}
	return true
}

func RGB(r, g, b uint8) OSL_color.RGBA {
	return OSL_color.RGBA{R: r, G: g, B: b, A: 255}
}

func RGBA(r, g, b, a uint8) OSL_color.RGBA {
	return OSL_color.RGBA{R: r, G: g, B: b, A: a}
}

/* -------------------- saving helpers -------------------- */

func (IMG) SavePNG(i *OSL_img_Image, path string) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return png.Encode(f, i.im) == nil
}

func (IMG) SaveJPEG(i *OSL_img_Image, path string, quality int) bool {
	if i == nil || i.closed || i.im == nil {
		return false
	}

	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}

	f, err := os.Create(path)
	if err != nil {
		return false
	}
	defer f.Close()

	return jpeg.Encode(f, i.im, &jpeg.Options{Quality: quality}) == nil
}

var img = IMG{}