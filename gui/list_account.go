package main

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"strings"

	"github.com/coinexchain/ColdWallet.win/keykeeper"
)

type AccModel struct {
	walk.ListModelBase
	items []string
}

func NewAccModel(infoList []keys.Info) *AccModel {
	m := &AccModel{items: make([]string, len(infoList))}
	for i, info := range infoList {
		m.items[i] = info.GetName()
	}
	return m
}

func (m *AccModel) ItemCount() int {
	return len(m.items)
}

func (m *AccModel) Value(index int) interface{} {
	return m.items[index]
}

type ListAccountsPage struct {
	*walk.Composite
	accountListBox *walk.ListBox
	model *AccModel
}

func newListAccountsPage(parent walk.Container, _ interface{}) (Page, error) {
	p := new(ListAccountsPage)
	var listErr error
	p.model = &AccModel{items: nil}
	if keykeeper.KB.IsOpen() {
		p.model = &AccModel{items: keykeeper.KB.GetStringItems()}
	}

	if err := (Composite{
		AssignTo: &p.Composite,
		Name:     "listAccountsPage",
		Layout:   VBox{},
		Children: []Widget{
			ListBox{
				AssignTo: &p.accountListBox,
				Model:    p.model,
			},
			Composite{
				Layout:   HBox{},
				Children: []Widget{
					PushButton{
						Text: T("delete"),
						OnClicked: func() {
							runDeleteAccount(p)
						},
					},
					PushButton{
						Text: T("changePassphrase"),
						OnClicked: func() {
							runChangePassphrase(p)
						},
					},
					PushButton{
						Text: T("showMnemonic"),
						OnClicked: func() {
							runShowMnemonic(p)
						},
					},
					PushButton{
						Text: T("showAddrQRCode"),
						OnClicked: func() {
							runShowAddrQRCode(p)
						},
					},
				},
			},
		},
	}).Create(NewBuilder(parent)); err != nil {
		return nil, err
	}

	if err := walk.InitWrapperWindow(p); err != nil {
		return nil, err
	}

	MainWin.Synchronize(func() {
		if listErr != nil {
			walk.MsgBox(MainWin, T("error!"), listErr.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			return
		}
	})

	return p, nil
}

func getSelectedAddr(p *ListAccountsPage) (item string, ok bool) {
	if !MainWin.CheckKBOpened() {
		return "", false
	}
	if len(p.model.items) == 0 {
		walk.MsgBox(MainWin, T("error!"), T("noAccYet"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return "", false
	}
	idx := p.accountListBox.CurrentIndex()
	if idx == -1 || len(p.model.items[idx]) == 0 {
		walk.MsgBox(MainWin, T("error!"), T("noSelAcc"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return "", false
	}
	item = p.model.items[idx]
	splitterPos := strings.Index(item, " - ") //bech32 address is before splitter
	if splitterPos == -1 {
		return "", false
	}
	return item[:splitterPos], true
}

func runDeleteAccount(p *ListAccountsPage) {
	addr, ok := getSelectedAddr(p)
	if !ok {
		return
	}
	ShowPassphraseDialog(MainWin, func(pass string) {
		err := keykeeper.DeleteAccount(addr, pass)
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			return
		}
	})
}

func runChangePassphrase(p *ListAccountsPage) {
	addr, ok := getSelectedAddr(p)
	if !ok {
		return
	}
	ShowChangePassphraseDialog(MainWin, func(oldPass, newPass string) {
		err := keykeeper.ChangePassphrase(addr, oldPass, newPass)
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			return
		}
	})
}

func runShowMnemonic(p *ListAccountsPage) {
	addr, ok := getSelectedAddr(p)
	if !ok {
		return
	}
	ShowPassphraseDialog(MainWin, func(pass string) {
		mnemonic, err := keykeeper.GetMnemonic(addr, pass)
		if err != nil {
			walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
			return
		}
		walk.MsgBox(MainWin, fmt.Sprintf(T("mnemonicOf"), addr),
			mnemonic, walk.MsgBoxIconError|walk.MsgBoxApplModal)
	})
}

func runShowAddrQRCode(p *ListAccountsPage) {
	addr, ok := getSelectedAddr(p)
	if !ok {
		return
	}
	err := ShowQRCodeDialog(MainWin, addr,
		fmt.Sprintf(T("qrCodeOfAddr"), addr),
		fmt.Sprintf(T("qrCodeOfAddrBelow"), addr),
	)
	if err != nil {
		walk.MsgBox(MainWin, T("error!"), err.Error(), walk.MsgBoxIconError|walk.MsgBoxApplModal)
	}
}

