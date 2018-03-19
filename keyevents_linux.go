// +build linux

package main

import (
	"log"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/keybind"
	"github.com/BurntSushi/xgbutil/xevent"
)

type keyBinding struct {
	Key       string
	Callback  func(string)
	XConn     *xgbutil.XUtil
	Connected bool
}

var keyBindings = make(map[string]*keyBinding)

func startKeybind(name string, key string, callback func(string)) bool {
	if keyBind, ok := keyBindings[name]; ok {
		if !keyBind.Connected {
			newBinding := keyBinding{keyBind.Key, keyBind.Callback, nil, false}
			delete(keyBindings, name)
			return startKeybind(name, newBinding.Key, newBinding.Callback)
		}

		return true
	}

	x, err := xgbutil.NewConn()

	if err != nil {
		log.Fatal(err)
		return false
	}

	keyBind := keyBinding{key, callback, x, false}

	keybind.Initialize(keyBind.XConn)

	callbackFunc := keybind.KeyPressFun(
		func(X *xgbutil.XUtil, e xevent.KeyPressEvent) {
			keyBind.Callback(key)
		})

	err = callbackFunc.Connect(keyBind.XConn, keyBind.XConn.RootWin(), keyBind.Key, true)

	if err != nil {
		log.Fatal(err)
		return false
	}

	keyBind.Connected = true

	pingBefore, pingAfter, pingQuit := xevent.MainPing(keyBind.XConn)

	go func() {
		for {
			select {
			case <-pingBefore:
				<-pingAfter
			case <-pingQuit:
				keyBind.Connected = false
				return
			}
		}
	}()

	keyBindings[name] = &keyBind

	return true
}

func stopKeybind(name string) bool {
	if keyBind, ok := keyBindings[name]; ok {
		if keyBind.Connected {
			keyBind.detachEventsAndDisconnect()
		}
	}

	return true
}

func (keyBind *keyBinding) detachEventsAndDisconnect() {
	keybind.Detach(keyBind.XConn, keyBind.XConn.RootWin())
	xevent.Detach(keyBind.XConn, keyBind.XConn.RootWin())
	xevent.Quit(keyBind.XConn)
}
