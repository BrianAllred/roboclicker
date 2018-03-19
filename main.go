package main

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/therecipe/qt/gui"

	"github.com/matryer/runner"
	"github.com/therecipe/qt/widgets"
)

type mButton byte

const (
	mButtonLeft   mButton = mButton(0)
	mButtonMiddle mButton = mButton(1)
	mButtonRight  mButton = mButton(2)
)

type mClick byte

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
	clickMouseButtonComboBox.AddItems([]string{"Left", "Middle", "Right"})
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
	screenXMin, screenYMin, screenXMax, screenYMax := getScreenSize()
	cursorLocXEdit.SetValidator(gui.NewQIntValidator2(screenXMin, screenXMax, cursorLocXEdit))
	cursorLocXEdit.SetPlaceholderText("X")
	cursorLocYEdit.SetValidator(gui.NewQIntValidator2(screenYMin, screenYMax, cursorLocYEdit))
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
		go startPickLocation()
	})

	startButton.ConnectClicked(func(checked bool) {
		go toggleAutoClick()
	})

	stopButton.ConnectClicked(func(checked bool) {
		go toggleAutoClick()
	})

	window.Show()

	go listenForStartStop()

	widgets.QApplication_Exec()
}

func listenForStartStop() {
	startKeybind("listenForStartStop", "f6", func(key string) {
		toggleAutoClick()
	})
}

func startPickLocation() {
	cursorPickLocButton.SetEnabled(false)
	startButton.SetEnabled(false)
	cursorPickLocRadio.SetChecked(true)
	cursorPickLocButton.SetText("Press Esc")

	time.Sleep(10 * time.Millisecond)

	startKeybind("pickLocation", "escape", func(key string) {
		stopPickLocation()
	})

	updating := false
	startMouseMotionBind("pickLocation", func(x int, y int) {
		if !updating {
			updating = true
			cursorLocXEdit.SetText(strconv.Itoa(x))
			cursorLocYEdit.SetText(strconv.Itoa(y))
			time.Sleep(10 * time.Millisecond)
			updating = false
		}
	})
}

func stopPickLocation() {
	stopKeybind("pickLocation")
	stopMousebind("pickLocation")

	time.Sleep(10 * time.Millisecond)

	cursorPickLocButton.SetEnabled(true)
	startButton.SetEnabled(true)
	cursorPickLocButton.SetText("Pick location")

	xPos, _ = strconv.Atoi(cursorLocXEdit.Text())
	yPos, _ = strconv.Atoi(cursorLocYEdit.Text())

	time.Sleep(10 * time.Millisecond)
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
			xPos, yPos = getMousePos()
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
	count := 0

	for {
		if cursorCurrentLocRadio.IsChecked() {
			xPos, yPos = getMousePos()
		} else {
			moveMouse(xPos, yPos)
		}

		go mouseClick(xPos, yPos, mouseButton, clickType)

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

func taskRunning(task *runner.Task) bool {
	return task != nil && task.Running()
}
