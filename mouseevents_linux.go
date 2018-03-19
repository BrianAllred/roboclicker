// +build linux

package main

import (
	"log"
	"time"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgb/xtest"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/mousebind"
	"github.com/BurntSushi/xgbutil/xevent"
)

type mouseBinding struct {
	Key       string
	Callback  func(int, int)
	XConn     *xgbutil.XUtil
	Connected bool
}

type pointerLocation struct {
	X int
	Y int
}

var mouseBindings = make(map[string]*mouseBinding)

func startMouseMotionBind(name string, callback func(int, int)) bool {
	if mouseBind, ok := mouseBindings[name]; ok {
		if !mouseBind.Connected {
			newBinding := mouseBinding{"", mouseBind.Callback, nil, false}
			delete(mouseBindings, name)
			return startMouseMotionBind(name, newBinding.Callback)
		}

		return true
	}

	x, err := xgbutil.NewConn()

	if err != nil {
		log.Fatal(err)
		return false
	}

	mouseBind := mouseBinding{"", callback, x, false}

	mousebind.Initialize(mouseBind.XConn)

	mouseBind.Connected = true

	mouseBindings[name] = &mouseBind

	go mouseBind.queryPointer()

	return true
}

func stopMousebind(name string) bool {
	if mouseBind, ok := mouseBindings[name]; ok {
		if mouseBind.Connected {
			mouseBind.Connected = false
			mouseBind.detachEventsAndDisconnect()
		}
	}

	return true
}

func (mouseBind *mouseBinding) detachEventsAndDisconnect() {
	mousebind.Detach(mouseBind.XConn, mouseBind.XConn.RootWin())
	xevent.Quit(mouseBind.XConn)
}

func getScreenSize() (xRoot, yRoot, xSize, ySize int) {
	xConn, err := xgbutil.NewConn()
	defer xevent.Quit(xConn)

	if err != nil {
		log.Fatal(err)
		return 0, 0, 0, 0
	}

	// X screen always starts at 0, 0
	return 0, 0, int(xConn.Screen().WidthInPixels), int(xConn.Screen().HeightInPixels)
}

func (mouseBind *mouseBinding) queryPointer() {
	for mouseBind.Connected && !xevent.Quitting(mouseBind.XConn) {
		pointer, err := xproto.QueryPointer(mouseBind.XConn.Conn(), mouseBind.XConn.RootWin()).Reply()

		if err != nil {
			log.Fatal(err)
			continue
		}

		mouseBind.Callback(int(pointer.RootX), int(pointer.RootY))

		time.Sleep(10 * time.Millisecond)
	}
}

func mouseClick(x int, y int, button mButton, click mClick) {
	xConn, err := xgbutil.NewConn()

	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		xevent.Quit(xConn)
	}()

	// Translate mButton struct to X11 mouse button
	xButton := byte(button) + 1

	xtest.Init(xConn.Conn())

	doMouseClick(xConn, xButton, x, y)

	if click == mClickDouble {
		doMouseClick(xConn, xButton, x, y)
	}

	xevent.Main(xConn)
}

func doMouseClick(xConn *xgbutil.XUtil, button byte, x int, y int) {
	xtest.FakeInput(xConn.Conn(), xproto.ButtonPress, button, 0, xConn.RootWin(), int16(x), int16(y), 0)
	xConn.Sync()
	xtest.FakeInput(xConn.Conn(), xproto.ButtonRelease, button, 0, xConn.RootWin(), int16(x), int16(y), 0)
	xConn.Sync()
}

func getMousePos() (x, y int) {
	xConn, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
		return
	}

	defer xevent.Quit(xConn)

	pointer, err := xproto.QueryPointer(xConn.Conn(), xConn.RootWin()).Reply()

	return int(pointer.RootX), int(pointer.RootY)
}

func moveMouse(x int, y int) {
	xConn, err := xgbutil.NewConn()

	if err != nil {
		log.Fatal(err)
		return
	}

	defer func() {
		xevent.Quit(xConn)
	}()

	xproto.WarpPointer(xConn.Conn(), xproto.Window(0), xConn.RootWin(), 0, 0, 0, 0, int16(x), int16(y))
}
