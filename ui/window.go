package ui

import (
	"image"
	"image/color"
	"log"

	"golang.org/x/exp/shiny/driver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/image/draw"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/mouse"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
)

type Visualizer struct {
	Title         string
	Debug         bool
	OnScreenReady func(s screen.Screen)

	w    screen.Window
	tx   chan screen.Texture
	done chan struct{}

	sz        size.Event
	pos       image.Rectangle
	figurePos image.Point // Центр фігури "Т"
}

func (pw *Visualizer) Main() {
	// Ініціалізація центру фігури.
	pw.figurePos = image.Point{400, 400} // Початкове положення в центрі.
	pw.tx = make(chan screen.Texture)
	pw.done = make(chan struct{})
	driver.Main(pw.run)
}

func (pw *Visualizer) Update(t screen.Texture) {
	pw.tx <- t
}

func (pw *Visualizer) run(s screen.Screen) {
	w, err := s.NewWindow(&screen.NewWindowOptions{
		Title:  pw.Title,
		Width:  800, // Ширина вікна
		Height: 800, // Висота вікна
	})
	if err != nil {
		log.Fatal("Failed to initialize the app window:", err)
	}
	defer func() {
		w.Release()
		close(pw.done)
	}()

	if pw.OnScreenReady != nil {
		pw.OnScreenReady(s)
	}

	pw.w = w

	events := make(chan any)
	go func() {
		for {
			e := w.NextEvent()
			if pw.Debug {
				log.Printf("new event: %v", e)
			}
			if detectTerminate(e) {
				close(events)
				break
			}
			events <- e
		}
	}()

	var t screen.Texture

	for {
		select {
		case e, ok := <-events:
			if !ok {
				return
			}
			pw.handleEvent(e, t)

		case t = <-pw.tx:
			w.Send(paint.Event{})
		}
	}
}

func detectTerminate(e any) bool {
	switch e := e.(type) {
	case lifecycle.Event:
		if e.To == lifecycle.StageDead {
			return true // Window destroy initiated.
		}
	case key.Event:
		if e.Code == key.CodeEscape {
			return true // Esc pressed.
		}
	}
	return false
}

func (pw *Visualizer) handleEvent(e any, t screen.Texture) {
	switch e := e.(type) {
	case size.Event:
		pw.sz = e
	case mouse.Event:
		if e.Button == mouse.ButtonLeft && e.Direction == mouse.DirPress {
			// Оновлення центру фігури при кліку лівою кнопкою миші.
			pw.figurePos = image.Point{int(e.X), int(e.Y)}
			pw.w.Send(paint.Event{}) // Перемальовка вікна.
		}
	case paint.Event:
		// Малювання контенту вікна.
		if t == nil {
			pw.drawDefaultUI()
		} else {
			pw.w.Scale(pw.sz.Bounds(), t, t.Bounds(), draw.Src, nil)
		}
		pw.w.Publish()
	}
}

func (pw *Visualizer) drawDefaultUI() {
	// Малювання білого фону.
	pw.w.Fill(pw.sz.Bounds(), color.White, draw.Src)

	// Малювання червоної фігури "Т".
	red := color.RGBA{255, 0, 0, 255}
	// Горизонтальний прямокутник.
	hRect := image.Rect(pw.figurePos.X-60, pw.figurePos.Y-50, pw.figurePos.X+60, pw.figurePos.Y-30)
	// Вертикальний прямокутник.
	vRect := image.Rect(pw.figurePos.X-10, pw.figurePos.Y-50, pw.figurePos.X+10, pw.figurePos.Y+50)
	pw.w.Fill(hRect, red, draw.Src)
	pw.w.Fill(vRect, red, draw.Src)
}
