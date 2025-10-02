// name: window
// description: PixelGL window wrapper
// author: Mist
// requires: github.com/faiface/pixel, github.com/faiface/pixel/pixelgl, github.com/faiface/pixel/imdraw, image, image/color

type OSLwinpkg struct{}

var mouse_x float64 = 0.0
var mouse_y float64 = 0.0
var mouse_down = false

type OSLWindow struct {
	win  *pixelgl.Window
	cfg  pixelgl.WindowConfig
	loop func(w *OSLWindow)
}

type OSLwinRender struct {
	win        *pixelgl.Window
	color      color.Color
	rectSprite *pixel.Sprite
	currentX   float64
	currentY   float64
	thickness  float64
}

var OSLdrawctx *OSLwinRender

func (OSLwinpkg) Create(setup func(w *OSLWindow)) {
	pixelgl.Run(func() {
		cfg := pixelgl.WindowConfig{
			Title:     "Untitled",
			Bounds:    pixel.R(0, 0, 800, 600),
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
		}

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

func (w *OSLWindow) Resize(width, height float64) {
	w.cfg.Bounds = pixel.R(0, 0, width, height)
	if w.win != nil {
		w.win.SetBounds(w.cfg.Bounds)
	}
}

func (w *OSLWindow) Goto(x, y int) {
	if w.win != nil {
		w.win.SetPos(pixel.V(float64(x), float64(y)))
	}
}

func (w *OSLWindow) Clear(col color.Color) {
	if w.win != nil {
		w.win.Clear(col)
	}
}

func (w *OSLWindow) Update() {
	if w.win != nil {
		w.win.Update()
	}
}

func (w *OSLWindow) Closed() bool {
	return w.win == nil || w.win.Closed()
}

func (w *OSLWindow) Close() {
	if w.win != nil {
		w.win.SetClosed(true)
	}
}

func (r *OSLwinRender) Hex(hex any) color.RGBA {
	h := OSLcastString(hex)
	var rr, gg, bb, aa uint8 = 0, 0, 0, 255
	if len(h) == 7 && h[0] == '#' {
		fmt.Sscanf(h, "#%02x%02x%02x", &rr, &gg, &bb)
	} else if len(h) == 9 && h[0] == '#' {
		fmt.Sscanf(h, "#%02x%02x%02x%02x", &rr, &gg, &bb, &aa)
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

func (r *OSLwinRender) Rect(width, height, rounding float64) {
	if r.win == nil || r.rectSprite == nil {
		return
	}

	col := r.color
	if col == nil {
		col = color.White
	}

	winBounds := r.win.Bounds()
	winWidth := winBounds.Max.X
	winHeight := winBounds.Max.Y

	pixelX := (winWidth / 2) + r.currentX
	pixelY := (winHeight / 2) - r.currentY

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

	iconStr := OSLcastString(icon)

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

func (r *OSLwinRender) SetThickness(thickness float64) {
	r.thickness = thickness
}

func (w *OSLWindow) MousePressed(button pixelgl.Button) bool {
	if w.win == nil {
		return false
	}
	return w.win.Pressed(button)
}

func (w *OSLWindow) MouseJustPressed(button pixelgl.Button) bool {
	if w.win == nil {
		return false
	}
	return w.win.JustPressed(button)
}

func (r *OSLwinRender) LineTo(x, y float64) {
	if r.win == nil {
		return
	}

	col := r.color
	if col == nil {
		col = color.White
	}

	winBounds := r.win.Bounds()
	winWidth := winBounds.Max.X
	winHeight := winBounds.Max.Y

	startX := (winWidth / 2) + r.currentX
	startY := (winHeight / 2) + r.currentY

	endX := (winWidth / 2) + x
	endY := (winHeight / 2) + y

	imd := imdraw.New(nil)
	imd.Color = col
	imd.EndShape = imdraw.RoundEndShape

	imd.Push(pixel.V(startX, startY))
	imd.Push(pixel.V(endX, endY))
	imd.Line(r.thickness)

	imd.Draw(r.win)

	r.currentX = x
	r.currentY = y
}

func (r *OSLwinRender) Text(txt string, size float64) {
	if r.win == nil {
		return
	}

	for _, c := range txt {
		if c == '\n' {
			r.currentX = 0
			r.currentY += size
			continue
		}

		if c == '\t' {
			r.currentX += size * 4
			continue
		}

		r.Goto(r.currentX, r.currentY)
		r.Icon(OSLfont[string(c)], size/30)
		r.currentX += size
	}
}

var window = OSLwinpkg{}