package main

import (
	//"fmt"
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"path"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/marcsauter/single"

	"github.com/coinexchain/ColdWallet.win/keykeeper"
)

type AppMainWindow struct {
	*MultiPageMainWindow
	prevDir string
}

func (mw *AppMainWindow) updateTitle(prefix string) {
	var buf bytes.Buffer

	if prefix != "" {
		buf.WriteString(prefix)
		buf.WriteString(" - ")
	}

	buf.WriteString(T("appName"))

	mw.SetTitle(buf.String())
}

func (mw *AppMainWindow) aboutActionTriggered() {
	walk.MsgBox(mw, T("aboutTitle"), T("aboutContent"), walk.MsgBoxOK|walk.MsgBoxIconInformation)
}

func (mw *AppMainWindow) openActionTriggered(withCreation bool) {
	dlg := new(walk.FileDialog)
	dlg.Filter = "json files (*.json)|*.json|All files (*.*)|*.*"
	dlg.FilterIndex = 1

	if len(mw.prevDir) != 0 {
		dlg.InitialDirPath = mw.prevDir
	}

	var ok bool
	var err error
	if withCreation {
		dlg.Title = T("selectSave")
		ok, err = dlg.ShowSave(mw)
	} else {
		dlg.Title = T("selectOpen")
		ok, err = dlg.ShowOpen(mw)
	}

	if err != nil {
		walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}
	if !ok {
		return
	}

	fname := dlg.FilePath
	if !strings.HasSuffix(dlg.FilePath, ".json") {
		fname = fname + ".json"
	}
	err = keykeeper.OpenKeybase(fname)
	if err != nil {
		walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}

	mw.updateTitle(fname)
	mw.prevDir, _ = path.Split(fname)
}

func (mw *AppMainWindow) scanQRCode() {
	ShowQRCodeScanDialog(mw, func(text string) {
		mw.MultiPageMainWindow.TextToSign = text
		action := mw.MultiPageMainWindow.pageActions[2]
		mw.MultiPageMainWindow.setCurrentAction(action)
	})
}

func (mw *AppMainWindow) CheckKBOpened() bool {
	if keykeeper.KeybaseOpened() {
		return true
	}
	walk.MsgBox(MainWin, T("notOpen"), T("kbNotOpen"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
	return false
}

var MainWin *AppMainWindow

//  go build -ldflags="-H windowsgui"
func main() {
	mw := new(AppMainWindow)
	MainWin = mw

	cfg := &MultiPageMainWindowConfig{
		Name:    "PrivateKeyKeeper",
		MinSize: Size{800, 600},
		MenuItems: []MenuItem{
			Menu{
				Text: T("file"),
				Items: []MenuItem{
					Action{
						Text:        T("open"),
						OnTriggered: func() { mw.openActionTriggered(false) },
					},
					Action{
						Text:        T("create&open"),
						OnTriggered: func() { mw.openActionTriggered(true) },
					},
					Separator{},
					Action{
						Text:        T("exit"),
						OnTriggered: func() {
							res := walk.MsgBox(MainWin, T("exit?"), T("sureToExit"),
								walk.MsgBoxYesNo|walk.MsgBoxIconQuestion|walk.MsgBoxApplModal)
							if res == walk.DlgCmdYes {
								keykeeper.CloseKeybase()
								walk.App().Exit(0)
							}
						},
					},
				},
			},
			Action{
				Text:        T("scanQRCode"),
				OnTriggered: func() {
					mw.scanQRCode()
				},
			},
			Menu{
				Text: T("help"),
				Items: []MenuItem{
					Action{
						Text:        T("about"),
						OnTriggered: func() { mw.aboutActionTriggered() },
					},
				},
			},
		},
		PageCfgs: []PageConfig{
			{T("ca"), AccountCreateImgData, newCreateAccountPage},
			{T("la"), AccountListImgData, newListAccountsPage},
			{T("sign"), AccountSignImgData, newSignPage},
		},
	}

	mpmw, err := NewMultiPageMainWindow(cfg)
	if err != nil {
		panic(err)
	}

	mw.MultiPageMainWindow = mpmw

	src := base64.NewDecoder(base64.StdEncoding, strings.NewReader(CoinexChainImgData))
	pngImg, err := png.Decode(src)
	if err != nil {
		panic(err)
	}
	icon, err := walk.NewIconFromImage(pngImg)
	if err != nil {
		panic(err)
	}
	mw.SetIcon(icon)

	mw.SetTitle(T("appName"))

	mw.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		res := walk.MsgBox(MainWin, T("exit?"), T("sureToExit"),
			walk.MsgBoxYesNo|walk.MsgBoxIconQuestion|walk.MsgBoxApplModal)
		if res == walk.DlgCmdNo {
			*canceled = true
		}
		keykeeper.CloseKeybase()
	})

	s := single.New("CoinEx_Chain_Cold_Wallet")
	if err := s.CheckLock(); err != nil && err == single.ErrAlreadyRunning {
		walk.MsgBox(MainWin, T("error!"), T("alreadyRun"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		walk.App().Exit(0)
	}

	font := mw.Font()
	fmt.Printf("%#v\n", font)
	newFont, err := walk.NewFont(font.Family(), font.PointSize()+1, font.Style())
	if err == nil {
		mw.SetFont(newFont)
	}

	action := mw.MultiPageMainWindow.pageActions[0]
	mw.MultiPageMainWindow.setCurrentAction(action)

	mw.Run()
}


