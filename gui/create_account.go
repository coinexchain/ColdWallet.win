package main

import (
	"fmt"
	"runtime"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"

	"github.com/coinexchain/ColdWallet.win/keykeeper"
)

type CreateAccountPage struct {
	*walk.Composite
	prefixLineEdit *walk.LineEdit
	suffixLineEdit *walk.LineEdit
	pass1LineEdit *walk.LineEdit
	pass2LineEdit *walk.LineEdit
	memoLineEdit *walk.LineEdit
	progressTextEdit *walk.TextEdit
	caButton *walk.PushButton
}

func newCreateAccountPage(parent walk.Container, _ interface{}) (Page, error) {
	p := new(CreateAccountPage)

	if err := (Composite{
		AssignTo: &p.Composite,
		Name:     "createAccountPage",
		Layout:   VBox{},
		Children: []Widget{
			Label{Text: T("caline1")},
			Label{Text: T("caline2")},
			Label{Text: T("caline3")},
			Label{Text: T("caline4")},
			Label{Text: T("caline5")},
			Label{Text: T("caline6")},
			Label{Text: T("caline7")},
			Composite{
				Layout:        Grid{Columns: 2},
				StretchFactor: 4,
				Children: []Widget{
					Label{Text: T("prefix")},
					LineEdit{AssignTo: &p.prefixLineEdit},
					Label{Text: T("suffix")},
					LineEdit{AssignTo: &p.suffixLineEdit},
					Label{Text: T("encryptPassphrase")},
					LineEdit{
						AssignTo: &p.pass1LineEdit,
						PasswordMode: true,
					},
					Label{Text: T("retypeEncryptPassphrase")},
					LineEdit{
						AssignTo: &p.pass2LineEdit,
						PasswordMode: true,
					},
					Label{Text: T("memo")},
					LineEdit{AssignTo: &p.memoLineEdit},
					Label{Text: T("progress")},
					TextEdit{
						AssignTo: &p.progressTextEdit,
						ReadOnly: true,
					},
				},
			},
			PushButton{
				Text: T("ca"),
				AssignTo: &p.caButton,
				MinSize: Size{100, 70},
				OnClicked: func() {
					runCreateAccount(p)
				},
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}

	return p, nil
}

func runCreateAccount(p *CreateAccountPage) {
	if !MainWin.CheckKBOpened() {
		return
	}
	pass1 := p.pass1LineEdit.Text()
	pass2 := p.pass2LineEdit.Text()
	if pass1 != pass2 {
		walk.MsgBox(MainWin, T("error!"), T("mismatchPassphrase"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}
	memo := p.memoLineEdit.Text()
	if len(memo) == 0 {
		walk.MsgBox(MainWin, T("error!"), T("emptyMemo"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}

	prefix := p.prefixLineEdit.Text()
	s, ok := keykeeper.CheckValid(prefix)
	if !ok {
		walk.MsgBox(MainWin, T("invalid_prefix"), fmt.Sprintf(T("invalid_char"), s),
			walk.MsgBoxIconWarning|walk.MsgBoxApplModal)
		return
	}
	prefix = keykeeper.AddrPrefix + prefix

	suffix := p.suffixLineEdit.Text()
	s, ok = keykeeper.CheckValid(suffix)
	if !ok {
		walk.MsgBox(MainWin, T("invalid_suffix"), fmt.Sprintf(T("invalid_char"), s),
			walk.MsgBoxIconWarning|walk.MsgBoxApplModal)
		return
	}

	if n := len(prefix + suffix); n > len(keykeeper.AddrPrefix)+7 {
		s := fmt.Sprintf(T("long_run_time"), n)
		walk.MsgBox(MainWin, T("warn"), s,
			walk.MsgBoxIconWarning|walk.MsgBoxApplModal)
	}
	coreCount := runtime.NumCPU()
	p.caButton.SetEnabled(false)
	go func() {
		addr, mnemonic := keykeeper.GenerateMnemonic(prefix, suffix, func(count uint64, percent float64) {
			MainWin.Synchronize(func() {
				s := fmt.Sprintf(T("estimate_progress"), count, percent)
				p.progressTextEdit.SetText(s)
				p.progressTextEdit.SetFocus()
			})
		}, coreCount)
		MainWin.Synchronize(func() {
			p.progressTextEdit.AppendText(fmt.Sprintf("===== %s ======\r\n", T("mnemonic")))
			p.progressTextEdit.AppendText(fmt.Sprintf("%s\r\n", mnemonic))
			p.progressTextEdit.AppendText(fmt.Sprintf("===== %s ======\r\n", T("address")))
			p.progressTextEdit.AppendText(fmt.Sprintf("%s\r\n", addr))
			p.progressTextEdit.SetFocus()

			if !keykeeper.KB.IsOpen() {
				walk.MsgBox(MainWin, T("error!"), T("notOpen"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
				return
			}

			_, err := keykeeper.CreateAccount(memo, mnemonic, pass1)
			if err != nil {
				walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			} else {
				walk.MsgBox(MainWin, T("success"), T("successCA")+addr, walk.MsgBoxIconInformation|walk.MsgBoxApplModal)
			}
			p.caButton.SetEnabled(true)
		})
	}()
}
