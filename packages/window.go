// name: window
// description: PixelGL window wrapper
// author: Mist
// requires: github.com/faiface/pixel, github.com/faiface/pixel/pixelgl, github.com/faiface/pixel/imdraw, image, image/color

type OSLwinRender struct {
	win        *pixelgl.Window
	color      color.Color
	rectSprite *pixel.Sprite
	currentX   float64
	currentY   float64
	thickness  float64
	direction  float64
	window     *OSLWindow
}

type OSLWindow struct {
	win  *pixelgl.Window
	cfg  pixelgl.WindowConfig
	loop func(w *OSLWindow)
}

type OSLwinpkg struct {
	configWidth  float64
	configHeight float64
	width        float64
	height       float64
	renderer     *OSLwinRender
	window       *OSLWindow
}

var mouse_x float64 = 0.0
var mouse_y float64 = 0.0
var mouse_down = false
var direction float64 = 0.0
var x_position float64 = 0.0
var y_position float64 = 0.0
var clicked = false
var window_width float64 = 0.0
var window_height float64 = 0.0

var OSLdrawctx *OSLwinRender

var window = &OSLwinpkg{
	width:  800,
	height: 600,
}

func (pkg *OSLwinpkg) Create(setup func(w *OSLWindow)) {
	pixelgl.Run(func() {
		width := pkg.configWidth
		height := pkg.configHeight
		if width <= 0 {
			width = 800
		}
		if height <= 0 {
			height = 600
		}

		pkg.width = width
		pkg.height = height

		cfg := pixelgl.WindowConfig{
			Title:     "Untitled",
			Bounds:    pixel.R(0, 0, width, height),
			VSync:     true,
			Resizable: true,
		}

		w := &OSLWindow{cfg: cfg, loop: nil}

		if setup != nil {
			setup(w)
		}

		win, err := pixelgl.NewWindow(w.cfg)
		if err != nil {
			panic(err)
		}
		w.win = win

		r := &OSLwinRender{
			win:       win,
			currentX:  0,
			currentY:  0,
			thickness: 1,
			direction: 0,
			window:    w,
		}

		pkg.renderer = r

		rectImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
		rectImg.Set(0, 0, color.White)
		rectPic := pixel.PictureDataFromImage(rectImg)
		r.rectSprite = pixel.NewSprite(rectPic, rectPic.Bounds())

		OSLdrawctx = r

		for !w.win.Closed() {
			w.win.Clear(color.RGBA{0, 0, 0, 255})

			mousePos := w.win.MousePosition()

			winBounds := w.win.Bounds()
			winWidth := winBounds.Max.X
			winHeight := winBounds.Max.Y

			pkg.width = winWidth
			pkg.height = winHeight
			window_width = winWidth
			window_height = winHeight

			mouse_x = mousePos.X - (winWidth / 2)
			mouse_y = mousePos.Y - (winHeight / 2)

			if w.loop != nil {
				w.loop(w)
			}

			w.win.Update()
		}
	})
}

func (w *OSLWindow) Run(loop func(w *OSLWindow)) {
	w.loop = loop
}

func (w *OSLWindow) width() float64 {
	return w.win.Bounds().Max.X
}

func (w *OSLWindow) height() float64 {
	return w.win.Bounds().Max.Y
}

func (w *OSLWindow) left() float64 {
	return w.win.Bounds().Max.X / -2
}

func (w *OSLWindow) right() float64 {
	return w.win.Bounds().Max.X / 2
}

func (w *OSLWindow) top() float64 {
	return w.win.Bounds().Max.Y / 2
}

func (w *OSLWindow) bottom() float64 {
	return w.win.Bounds().Max.Y / -2
}

func (w *OSLWindow) SetTitle(title string) {
	w.cfg.Title = title
	if w.win != nil {
		w.win.SetTitle(title)
	}
}

func (w *OSLWindow) resize(width, height any) {
	w.cfg.Bounds = pixel.R(0, 0, OSLcastNumber(width), OSLcastNumber(height))
	if w.win != nil {
		w.win.SetBounds(w.cfg.Bounds)
	}
}

func (w *OSLWindow) setResizable(resizable bool) {
	return
}

func (w *OSLWindow) Goto(x, y any) {
	if w.win == nil {
		return
	}
	x_position = OSLcastNumber(x)
	y_position = OSLcastNumber(y)
	w.win.SetPos(pixel.V(x_position, y_position))
}

func (w *OSLWindow) Clear(col color.Color) {
	if w.win == nil {
		return
	}
	w.win.Clear(col)
}

func (w *OSLWindow) Update() {
	if w.win != nil {
		w.win.Update()
	}
}

func (w *OSLWindow) close() {
	if w.win != nil {
		w.win.SetClosed(true)
	}
}

func (r *OSLwinRender) Hex(hex any) color.RGBA {
	h := OSLtoString(hex)
	var rr, gg, bb, aa uint8 = 0, 0, 0, 255
	if h[0] == '#' {
		h = h[1:]
	}
	if len(h) == 6 {
		fmt.Sscanf(h, "%02x%02x%02x", &rr, &gg, &bb)
	} else if len(h) == 8 {
		fmt.Sscanf(h, "%02x%02x%02x%02x", &rr, &gg, &bb, &aa)
	} else if len(h) == 3 {
		fmt.Sscanf(h, "%1x%1x%1x", &rr, &gg, &bb)
		rr *= 17
		gg *= 17
		bb *= 17
	} else if len(h) == 4 {
		fmt.Sscanf(h, "%1x%1x%1x%1x", &rr, &gg, &bb, &aa)
		rr *= 17
		gg *= 17
		bb *= 17
		aa *= 17
	}
	return color.RGBA{R: rr, G: gg, B: bb, A: aa}
}

// Render methods
func (r *OSLwinRender) Color(col any) {
	r.color = r.Hex(col)
}

func (r *OSLwinRender) Goto(x, y float64) {
	r.currentX = x
	r.currentY = y
}

func (w *OSLwinRender) Loc(a, b, c, d any) {
	numA := OSLcastNumber(a)
	numB := OSLcastNumber(b)
	numC := OSLcastNumber(c)
	numD := OSLcastNumber(d)

	w.Goto((window.width/-numA)+numC, (window.height/numB)+numD)
}

func (r *OSLwinRender) LineTo(endX, endY float64) {
	if r.win == nil {
		return
	}

	winBounds := r.win.Bounds()
	winWidth := winBounds.Max.X
	winHeight := winBounds.Max.Y

	startPixelX := (winWidth / 2) + r.currentX
	startPixelY := (winHeight / 2) + r.currentY
	endPixelX := (winWidth / 2) + endX
	endPixelY := (winHeight / 2) + endY

	col := r.color
	if col == nil {
		col = color.White
	}

	thickness := r.thickness
	if thickness <= 0 {
		thickness = 1
	}

	imd := imdraw.New(nil)
	imd.Color = col
	imd.EndShape = imdraw.RoundEndShape
	imd.Push(pixel.V(startPixelX, startPixelY))
	imd.Push(pixel.V(endPixelX, endPixelY))
	imd.Line(thickness)
	imd.Draw(r.win)

	r.currentX = endX
	r.currentY = endY
}

func (r *OSLwinRender) Rect(args ...any) {
	if r.win == nil || r.rectSprite == nil {
		return
	}

	width := OSLcastNumber(args[0])
	height := OSLcastNumber(args[1])
	rounding := OSLcastNumber(args[2])

	col := r.color
	if col == nil {
		col = color.White
	}

	winBounds := r.win.Bounds()
	winWidth := winBounds.Max.X
	winHeight := winBounds.Max.Y

	pixelX := (winWidth / 2) + r.currentX
	pixelY := (winHeight / 2) + r.currentY

	centerX := pixelX
	centerY := pixelY

	mat := pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(width, height)).
		Moved(pixel.V(centerX, centerY))

	r.rectSprite.DrawColorMask(r.win, mat, col)

	if rounding > 0 {
		prevThickness := r.thickness
		r.thickness = rounding
		centerX = r.currentX
		centerY = r.currentY
		r.Goto(centerX+width/2, centerY-height/2)
		r.LineTo(centerX+width/2, centerY+height/2)
		r.LineTo(centerX-width/2, centerY+height/2)
		r.LineTo(centerX-width/2, centerY-height/2)
		r.LineTo(centerX+width/2, centerY-height/2)
		r.thickness = prevThickness
	}
}

func (r *OSLwinRender) Icon(icon any, size float64) {
	if r.win == nil {
		return
	}

	iconStr := OSLtoString(icon)

	starting_pos := []float64{r.currentX, r.currentY}

	parts := []string{}

	part := ""
	for i := 0; i < len(iconStr); i++ {
		char := iconStr[i]
		if char == ' ' || char == '\n' || char == '\t' {
			if part != "" {
				parts = append(parts, part)
				part = ""
			}
			continue
		}
		part += string(char)
	}

	if part != "" {
		parts = append(parts, part)
	}

	offsetX := starting_pos[0]
	offsetY := starting_pos[1]

	for i := 0; i < len(parts); i++ {
		part := parts[i]
		switch part {
		case "c":
			if i+1 < len(parts) {
				r.Color(parts[i+1])
				i++
			}
		case "w":
			if i+1 < len(parts) {
				r.SetThickness(OSLcastNumber(parts[i+1]) * size)
				i++
			}
		case "line":
			if i+4 < len(parts) {
				startX := OSLcastNumber(parts[i+1])*size + offsetX
				startY := OSLcastNumber(parts[i+2])*size + offsetY
				endX := OSLcastNumber(parts[i+3])*size + offsetX
				endY := OSLcastNumber(parts[i+4])*size + offsetY
				r.Goto(startX, startY)
				r.LineTo(endX, endY)
				i += 4
			}
		case "cont":
			if i+2 < len(parts) {
				endX := OSLcastNumber(parts[i+1])*size + offsetX
				endY := OSLcastNumber(parts[i+2])*size + offsetY
				r.LineTo(endX, endY)
				i += 2
			}
		case "dot":
			if i+2 < len(parts) {
				curX := OSLcastNumber(parts[i+1])*size + offsetX
				curY := OSLcastNumber(parts[i+2])*size + offsetY
				r.Goto(curX, curY)
				r.LineTo(curX+0.001, curY+0.001)
				i += 2
			}
		}
	}

	r.currentX = starting_pos[0]
	r.currentY = starting_pos[1]
}

func (r *OSLwinRender) Text(text string, size any) {
	if r.win == nil {
		return
	}

	text = OSLtoString(text)
	sizeNum := OSLcastNumber(size) * 2

	font := OSLfont

	if font == nil {
		return
	}

	startX := r.currentX
	startY := r.currentY

	for _, char := range text {
		if char == '\n' {
			r.currentX = startX
			r.currentY -= sizeNum
			continue
		}
		if char == ' ' {
			r.currentX += sizeNum
			continue
		}
		glyph, ok := font[string(char)]
		if !ok {
			continue
		}
		r.Goto(r.currentX, r.currentY)
		originalX := r.currentX
		r.Icon(glyph, sizeNum/40)
		r.currentX = originalX + sizeNum
	}
	r.Goto(r.currentX, startY)
}

func (r *OSLwinRender) SetThickness(thickness float64) {
	r.thickness = thickness
}

func (r *OSLwinRender) Change(offsetX, offsetY float64) {
	r.currentX += offsetX
	r.currentY += offsetY
}

func (r *OSLwinRender) Direction(direction float64) {
	r.direction = direction
}

var OSLkeyMap = map[string]pixelgl.Button{
	// Letters
	"a": pixelgl.KeyA, "b": pixelgl.KeyB, "c": pixelgl.KeyC, "d": pixelgl.KeyD,
	"e": pixelgl.KeyE, "f": pixelgl.KeyF, "g": pixelgl.KeyG, "h": pixelgl.KeyH,
	"i": pixelgl.KeyI, "j": pixelgl.KeyJ, "k": pixelgl.KeyK, "l": pixelgl.KeyL,
	"m": pixelgl.KeyM, "n": pixelgl.KeyN, "o": pixelgl.KeyO, "p": pixelgl.KeyP,
	"q": pixelgl.KeyQ, "r": pixelgl.KeyR, "s": pixelgl.KeyS, "t": pixelgl.KeyT,
	"u": pixelgl.KeyU, "v": pixelgl.KeyV, "w": pixelgl.KeyW, "x": pixelgl.KeyX,
	"y": pixelgl.KeyY, "z": pixelgl.KeyZ,

	// Numbers
	"0": pixelgl.Key0, "1": pixelgl.Key1, "2": pixelgl.Key2, "3": pixelgl.Key3,
	"4": pixelgl.Key4, "5": pixelgl.Key5, "6": pixelgl.Key6, "7": pixelgl.Key7,
	"8": pixelgl.Key8, "9": pixelgl.Key9,

	// Special keys
	"space":     pixelgl.KeySpace,
	"enter":     pixelgl.KeyEnter,
	"escape":    pixelgl.KeyEscape,
	"esc":       pixelgl.KeyEscape,
	"backspace": pixelgl.KeyBackspace,
	"tab":       pixelgl.KeyTab,

	// Modifiers
	"shift":      pixelgl.KeyLeftShift,
	"leftshift":  pixelgl.KeyLeftShift,
	"rightshift": pixelgl.KeyRightShift,
	"ctrl":       pixelgl.KeyLeftControl,
	"control":    pixelgl.KeyLeftControl,
	"leftctrl":   pixelgl.KeyLeftControl,
	"rightctrl":  pixelgl.KeyRightControl,
	"alt":        pixelgl.KeyLeftAlt,
	"leftalt":    pixelgl.KeyLeftAlt,
	"rightalt":   pixelgl.KeyRightAlt,

	// Arrow keys
	"up":    pixelgl.KeyUp,
	"down":  pixelgl.KeyDown,
	"left":  pixelgl.KeyLeft,
	"right": pixelgl.KeyRight,
}

func (w *OSLWindow) KeyPressed(key string) bool {
	if w.win == nil {
		return false
	}

	if btn, ok := OSLkeyMap[strings.ToLower(key)]; ok {
		return w.win.Pressed(btn)
	}

	return false
}