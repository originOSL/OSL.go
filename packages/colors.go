// name: color
// description: Color utilities
// author: Mist
// requires: image/color

type OSLCS struct{}

func (OSLCS) RGBA(r, g, b, a any) color.RGBA {
	return color.RGBA{
		R: uint8(OSLcastInt(r)),
		G: uint8(OSLcastInt(g)),
		B: uint8(OSLcastInt(b)),
		A: uint8(OSLcastInt(a)),
	}
}

func (OSLCS) RGB(r, g, b any) color.RGBA {
	return color.RGBA{
		R: uint8(OSLcastInt(r)),
		G: uint8(OSLcastInt(g)),
		B: uint8(OSLcastInt(b)),
		A: 255,
	}
}

func (OSLCS) Gray(v any) color.Gray {
	return color.Gray{Y: uint8(OSLcastInt(v))}
}

func (OSLCS) NRGBA(r, g, b, a any) color.NRGBA {
	return color.NRGBA{
		R: uint8(OSLcastInt(r)),
		G: uint8(OSLcastInt(g)),
		B: uint8(OSLcastInt(b)),
		A: uint8(OSLcastInt(a)),
	}
}

func (OSLCS) Hex(hex any) color.RGBA {
	h := OSLtoString(hex)
	var r, g, b, a uint8 = 0, 0, 0, 255
	if len(h) == 7 && h[0] == '#' {
		fmt.Sscanf(h, "#%02x%02x%02x", &r, &g, &b)
	} else if len(h) == 9 && h[0] == '#' {
		fmt.Sscanf(h, "#%02x%02x%02x%02x", &r, &g, &b, &a)
	}
	return color.RGBA{R: r, G: g, B: b, A: a}
}

var colors = OSLCS{}
