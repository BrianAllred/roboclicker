package main

import (
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/therecipe/qt/gui"

	"github.com/go-vgo/robotgo"
	"github.com/matryer/runner"
	"github.com/therecipe/qt/widgets"
)

type mButton int

const (
	mButtonLeft   mButton = mButton(0)
	mButtonRight  mButton = mButton(1)
	mButtonMiddle mButton = mButton(2)
)

type mClick int

const (
	mClickSingle mClick = mClick(0)
	mClickDouble mClick = mClick(1)
)

// QT widgets
var (
	application               = widgets.NewQApplication(len(os.Args), os.Args)
	intervalGroup             = widgets.NewQGroupBox2("Click interval", nil)
	intervalHoursEdit         = widgets.NewQLineEdit(nil)
	intervalMinsEdit          = widgets.NewQLineEdit(nil)
	intervalSecsEdit          = widgets.NewQLineEdit(nil)
	intervalMillisecondsEdit  = widgets.NewQLineEdit(nil)
	intervalLayout            = widgets.NewQGridLayout2()
	clickOptsGroup            = widgets.NewQGroupBox2("Click options", nil)
	clickMouseButtonLabel     = widgets.NewQLabel2("Mouse button:", nil, 0)
	clickMouseButtonComboBox  = widgets.NewQComboBox(nil)
	clickClickTypeLabel       = widgets.NewQLabel2("Click type:", nil, 0)
	clickClickTypeComboBox    = widgets.NewQComboBox(nil)
	clickOptsLayout           = widgets.NewQGridLayout2()
	clickRepeatGroup          = widgets.NewQGroupBox2("Click repeat", nil)
	clickRepeatTimesRadio     = widgets.NewQRadioButton2("Repeat", nil)
	clickRepeatInfRadio       = widgets.NewQRadioButton2("Repeat until stopped", nil)
	clickRepeatTimesSpin      = widgets.NewQSpinBox(nil)
	clickRepeatTimesSpinLabel = widgets.NewQLabel2("times", nil, 0)
	clickRepeatLayout         = widgets.NewQGridLayout2()
	cursorGroup               = widgets.NewQGroupBox2("Cursor position", nil)
	cursorCurrentLocRadio     = widgets.NewQRadioButton2("Current location", nil)
	cursorPickLocRadio        = widgets.NewQRadioButton(nil)
	cursorPickLocButton       = widgets.NewQPushButton2("Pick location", nil)
	cursorLocXEdit            = widgets.NewQLineEdit(nil)
	cursorLocYEdit            = widgets.NewQLineEdit(nil)
	cursorLayout              = widgets.NewQGridLayout2()
	startButton               = widgets.NewQPushButton2("Start (F6)", nil)
	stopButton                = widgets.NewQPushButton2("Stop (F6)", nil)
	layout                    = widgets.NewQGridLayout2()
	window                    = widgets.NewQMainWindow(nil, 0)
	centralWidget             = widgets.NewQWidget(window, 0)
)

var clickTask, startStopListenerTask *runner.Task
var interval time.Duration
var mouseButton mButton
var clickType mClick
var xPos, yPos, repeat int
var stopKeyLog, stopListen chan struct{}
var wg sync.WaitGroup

func main() {
	switch runtime.GOOS {
	case "windows":
		application.SetStyle(widgets.QStyleFactory_Create("Fusion"))
	}

	// Click interval section
	intervalHoursEdit.SetPlaceholderText("hours")
	intervalHoursEdit.SetValidator(gui.NewQIntValidator(intervalHoursEdit))
	intervalMinsEdit.SetPlaceholderText("minutes")
	intervalMinsEdit.SetValidator(gui.NewQIntValidator(intervalMinsEdit))
	intervalSecsEdit.SetPlaceholderText("seconds")
	intervalSecsEdit.SetValidator(gui.NewQIntValidator(intervalSecsEdit))
	intervalMillisecondsEdit.SetPlaceholderText("milliseconds")

	var msecondsValidator = gui.NewQIntValidator(intervalMillisecondsEdit)
	msecondsValidator.SetBottom(1)
	intervalMillisecondsEdit.SetValidator(msecondsValidator)

	intervalLayout.AddWidget(intervalHoursEdit, 0, 0, 0)
	intervalLayout.AddWidget(intervalMinsEdit, 0, 1, 0)
	intervalLayout.AddWidget(intervalSecsEdit, 0, 2, 0)
	intervalLayout.AddWidget(intervalMillisecondsEdit, 0, 3, 0)
	intervalGroup.SetLayout(intervalLayout)

	// Click options section
	clickMouseButtonComboBox.AddItems([]string{"Left", "Right", "Middle"})
	clickClickTypeComboBox.AddItems([]string{"Single", "Double"})

	clickOptsLayout.AddWidget(clickMouseButtonLabel, 0, 0, 0)
	clickOptsLayout.AddWidget(clickMouseButtonComboBox, 0, 1, 0)
	clickOptsLayout.AddWidget(clickClickTypeLabel, 1, 0, 0)
	clickOptsLayout.AddWidget(clickClickTypeComboBox, 1, 1, 0)
	clickOptsGroup.SetLayout(clickOptsLayout)

	// Click repeat section
	clickRepeatTimesRadio.SetChecked(true)
	clickRepeatTimesSpin.SetMinimum(1)

	clickRepeatLayout.AddWidget(clickRepeatTimesRadio, 0, 0, 0)
	clickRepeatLayout.AddWidget(clickRepeatTimesSpin, 0, 1, 0)
	clickRepeatLayout.AddWidget(clickRepeatTimesSpinLabel, 0, 2, 0)
	clickRepeatLayout.AddWidget(clickRepeatInfRadio, 1, 0, 0)
	clickRepeatGroup.SetLayout(clickRepeatLayout)

	// Cursor position section
	screenXMax, screenYMax := robotgo.GetScreenSize()
	cursorLocXEdit.SetValidator(gui.NewQIntValidator2(0, screenXMax, cursorLocXEdit))
	cursorLocXEdit.SetPlaceholderText("X")
	cursorLocYEdit.SetValidator(gui.NewQIntValidator2(0, screenYMax, cursorLocYEdit))
	cursorLocYEdit.SetPlaceholderText("Y")
	cursorCurrentLocRadio.SetChecked(true)

	cursorLayout.AddWidget(cursorCurrentLocRadio, 0, 0, 0)
	cursorLayout.AddWidget(cursorPickLocRadio, 0, 1, 0)
	cursorLayout.AddWidget(cursorPickLocButton, 0, 2, 0)
	cursorLayout.AddWidget(cursorLocXEdit, 0, 3, 0)
	cursorLayout.AddWidget(cursorLocYEdit, 0, 4, 0)
	cursorGroup.SetLayout(cursorLayout)

	// Button section
	stopButton.SetEnabled(false)

	// Setup main layout
	layout.AddWidget3(intervalGroup, 0, 0, 1, 0, 0)
	layout.AddWidget(clickOptsGroup, 1, 0, 0)
	layout.AddWidget(clickRepeatGroup, 1, 1, 0)
	layout.AddWidget3(cursorGroup, 2, 0, 1, 0, 0)
	layout.AddWidget(startButton, 3, 0, 0)
	layout.AddWidget(stopButton, 3, 1, 0)

	window.SetWindowTitle("roboclicker")

	centralWidget.SetLayout(layout)
	window.SetCentralWidget(centralWidget)

	// Setup button events
	cursorPickLocButton.ConnectClicked(func(checked bool) {
		go pickLocation()
	})

	startButton.ConnectClicked(func(checked bool) {
		go toggleAutoClick()
	})

	stopButton.ConnectClicked(func(checked bool) {
		go toggleAutoClick()
	})

	window.Show()

	stopListen = make(chan struct{})
	go listenForStartStop(stopListen)

	widgets.QApplication_Exec()
}

func keylog(keyPress chan int, done <-chan struct{}, key string) {
	select {
	case keyPress <- robotgo.AddEvent(key):
	case <-stopKeyLog:
		keyPress <- 1
	}
}

func pickLocation() {
	defer func() {
		stopListen = make(chan struct{})
		go listenForStartStop(stopListen)
	}()

	close(stopListen)

	wg.Wait()

	time.Sleep(10 * time.Millisecond)

	cursorPickLocButton.SetEnabled(false)
	startButton.SetEnabled(false)
	cursorPickLocRadio.SetChecked(true)
	cursorPickLocButton.SetText("Press Esc")

	time.Sleep(10 * time.Millisecond)

	loop := true

	escChan := make(chan int)
	stopKeyLog = make(chan struct{})
	go keylog(escChan, stopKeyLog, "esc")

	for loop {
		x, y := robotgo.GetMousePos()

		time.Sleep(100 * time.Millisecond)

		cursorLocXEdit.SetText(strconv.Itoa(x))
		cursorLocYEdit.SetText(strconv.Itoa(y))

		time.Sleep(10 * time.Millisecond)

		select {
		case key := <-escChan:
			if key == 0 {
				loop = false
			}
		default:
		}
	}

	time.Sleep(10 * time.Millisecond)

	cursorPickLocButton.SetEnabled(true)
	startButton.SetEnabled(true)
	cursorPickLocButton.SetText("Pick location")

	time.Sleep(10 * time.Millisecond)
}

func listenForStartStop(done <-chan struct{}) {
	defer wg.Done()

	wg.Add(1)

	f6Chan := make(chan int)
	stopKeyLog = make(chan struct{})
	go keylog(f6Chan, stopKeyLog, "f6")

	loop := true

	for loop {
		time.Sleep(100 * time.Millisecond)

		select {
		case key := <-f6Chan:
			if key == 0 {
				go toggleAutoClick()
				f6Chan = make(chan int)
				go keylog(f6Chan, stopKeyLog, "f6")
			}
		case <-done:
			loop = false
			close(stopKeyLog)
			robotgo.StopEvent()
		default:
		}
	}
}

func toggleAutoClick() {
	if clicking() {
		stopAutoClick()
	} else {
		startAutoClick()
	}
}

func startAutoClick() {
	if clicking() {
		return
	}

	validInterval := validateInterval()

	if validInterval {
		mouseButton = mButton(clickMouseButtonComboBox.CurrentIndex())
		clickType = mClick(clickClickTypeComboBox.CurrentIndex())

		if cursorCurrentLocRadio.IsChecked() {
			xPos, yPos = robotgo.GetMousePos()
		} else {
			xPos, _ = strconv.Atoi(cursorLocXEdit.Text())
			yPos, _ = strconv.Atoi(cursorLocYEdit.Text())
		}

		if clickRepeatInfRadio.IsChecked() {
			repeat = -1
		} else {
			repeat = clickRepeatTimesSpin.Value()
		}

		clickTask = runner.Go(autoClick)
		startButton.SetEnabled(false)
		cursorPickLocButton.SetEnabled(false)
		stopButton.SetEnabled(true)
	} else {
		intervalMillisecondsEdit.SetFocus2()
	}
}

func stopAutoClick() {
	if clicking() {
		clickTask.Stop()
		<-clickTask.StopChan()
	}

	startButton.SetEnabled(true)
	cursorPickLocButton.SetEnabled(true)
	stopButton.SetEnabled(false)
}

func autoClick(shouldStop runner.S) error {
	var mouseButtonStr string
	var doubleClick bool

	count := 0

	for {
		if cursorCurrentLocRadio.IsChecked() {
			xPos, yPos = robotgo.GetMousePos()
		} else {
			robotgo.MoveMouse(xPos, yPos)
		}

		switch mouseButton {
		case mButtonLeft:
			mouseButtonStr = "left"
		case mButtonRight:
			mouseButtonStr = "right"
		case mButtonMiddle:
			mouseButtonStr = "center"
		}

		doubleClick = clickType == mClickDouble

		robotgo.MouseClick(mouseButtonStr, doubleClick)

		if repeat > 0 {
			count++
			if count >= repeat {
				go stopAutoClick()
				break
			}
		}

		if shouldStop() {
			break
		}

		// move to goroutine
		time.Sleep(interval)
	}

	return nil
}

func clicking() bool {
	return clickTask != nil && clickTask.Running()
}

func validateInterval() bool {
	var intervalHours string
	var intervalMins string
	var intervalSecs string
	var intervalMSecs string

	if len(intervalHoursEdit.Text()) == 0 && len(intervalMinsEdit.Text()) == 0 && len(intervalSecsEdit.Text()) == 0 && len(intervalMillisecondsEdit.Text()) == 0 {
		return false
	}

	if len(intervalHoursEdit.Text()) == 0 {
		intervalHours = "0h"
	} else {
		intervalHours = intervalHoursEdit.Text() + "h"
	}

	if len(intervalMinsEdit.Text()) == 0 {
		intervalMins = "0m"
	} else {
		intervalMins = intervalMinsEdit.Text() + "m"
	}

	if len(intervalSecsEdit.Text()) == 0 {
		intervalSecs = "0s"
	} else {
		intervalSecs = intervalSecsEdit.Text() + "s"
	}

	if len(intervalMillisecondsEdit.Text()) == 0 {
		intervalMSecs = "0ms"
	} else {
		intervalMSecs = intervalMillisecondsEdit.Text() + "ms"
	}

	interval, _ = time.ParseDuration(intervalHours + intervalMins + intervalSecs + intervalMSecs)

	if interval.Nanoseconds() == 0 {
		return false
	}

	return true
}
