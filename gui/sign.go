package main

import (
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"

	"github.com/coinexchain/ColdWallet.win/keykeeper"
	"github.com/coinexchain/ColdWallet.win/msg"
)

type SignPage struct {
	*walk.Composite
	origTextEdit     *walk.TextEdit
	readableTextEdit *walk.TextEdit
	parsedMsg        *msg.Msg
}

func newSignPage(parent walk.Container, extraInfo interface{}) (Page, error) {
	p := new(SignPage)
	origText := extraInfo.(string)

	if err := (Composite{
		AssignTo: &p.Composite,
		Name:     "SignPage",
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout:        Grid{Columns: 2},
				StretchFactor: 4,
				Children: []Widget{
					Label{Text: T("origMsg")},
					TextEdit{
						AssignTo: &p.origTextEdit,
						ReadOnly: true,
						OnTextChanged: func() {
							parseOrigMsg(p)
						},
					},
					Label{Text: T("readableMsg")},
					TextEdit{
						AssignTo: &p.readableTextEdit,
						ReadOnly: true,
					},
				},
			},
			PushButton{
				Text: T("sign"),
				MinSize: Size{100, 70},
				OnClicked: func() {
					runSign(p)
				},
			},
			VSpacer{},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}


	if len(origText) != 0 {
		p.origTextEdit.SetText(origText)
		MainWin.Show()
		MainWin.BringToTop()
		MainWin.SetFocus()
		win.SetForegroundWindow(MainWin.Handle())
		win.SetActiveWindow(MainWin.Handle())

		win.AttachThreadInput(
			int32(win.GetWindowThreadProcessId(win.GetForegroundWindow(), nil)),
			int32(win.GetCurrentThreadId()), true)
	}

	return p, nil
}

func parseOrigMsg(p *SignPage) {
	jsonStr := p.origTextEdit.Text()
	msg, err := msg.ParseJson(jsonStr)
	if err != nil {
		p.parsedMsg = nil
		p.readableTextEdit.SetText(err.Error())
	}
	p.parsedMsg = &msg
}

func runSign(p *SignPage) {
	if !MainWin.CheckKBOpened() {
		return
	}
	signer := p.parsedMsg.GetSigner()
	if !keykeeper.HasAccount(signer) {
		walk.MsgBox(MainWin, T("error!"), T("notHaveAcc")+signer, walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}
	signBytes := p.parsedMsg.GetSignBytes()

	doSign := func(passphrase string) {
		signedResult, err := keykeeper.Sign(signer, passphrase, signBytes)
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			return
		}
		err = ShowQRCodeDialog(MainWin, signedResult, T("successSign"), T("seeSignQRBelow"))
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		}
	}

	passphrase, ok := keykeeper.GetCachedPassphrase(signer)
	if ok {
		doSign(passphrase)
		return
	}

	ShowPassphraseDialog(MainWin, func(passphrase string) {
		err := keykeeper.AddCachedPassphrase(signer, passphrase)
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		}
		doSign(passphrase)
	})
}

