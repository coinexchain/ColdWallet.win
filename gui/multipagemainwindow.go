// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/base64"
	"github.com/lxn/win"
	"image/png"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type MultiPageMainWindowConfig struct {
	Name                 string
	Enabled              Property
	Visible              Property
	Font                 Font
	MinSize              Size
	MaxSize              Size
	ContextMenuItems     []MenuItem
	OnKeyDown            walk.KeyEventHandler
	OnKeyPress           walk.KeyEventHandler
	OnKeyUp              walk.KeyEventHandler
	OnMouseDown          walk.MouseEventHandler
	OnMouseMove          walk.MouseEventHandler
	OnMouseUp            walk.MouseEventHandler
	OnSizeChanged        walk.EventHandler
	OnCurrentPageChanged walk.EventHandler
	Title                string
	Size                 Size
	MenuItems            []MenuItem
	ToolBar              ToolBar
	PageCfgs             []PageConfig
}

type PageConfig struct {
	Title   string
	Image   string
	NewPage PageFactoryFunc
}

type PageFactoryFunc func(parent walk.Container, extraInfo interface{}) (Page, error)

type Page interface {
	// Provided by Walk
	walk.Container
	Parent() walk.Container
	SetParent(parent walk.Container) error
}

type MultiPageMainWindow struct {
	*walk.MainWindow
	navTB                       *walk.ToolBar
	pageCom                     *walk.Composite
	action2NewPage              map[*walk.Action]PageFactoryFunc
	pageActions                 []*walk.Action
	currentAction               *walk.Action
	currentPage                 Page
	currentPageChangedPublisher walk.EventPublisher
	TextToSign                  string
}

func NewMultiPageMainWindow(cfg *MultiPageMainWindowConfig) (*MultiPageMainWindow, error) {
	mpmw := &MultiPageMainWindow{
		action2NewPage: make(map[*walk.Action]PageFactoryFunc),
	}

	if err := (MainWindow{
		AssignTo:         &mpmw.MainWindow,
		Name:             cfg.Name,
		Title:            cfg.Title,
		Enabled:          cfg.Enabled,
		Visible:          cfg.Visible,
		Font:             cfg.Font,
		MinSize:          cfg.MinSize,
		MaxSize:          cfg.MaxSize,
		MenuItems:        cfg.MenuItems,
		ToolBar:          cfg.ToolBar,
		ContextMenuItems: cfg.ContextMenuItems,
		OnKeyDown:        cfg.OnKeyDown,
		OnKeyPress:       cfg.OnKeyPress,
		OnKeyUp:          cfg.OnKeyUp,
		OnMouseDown:      cfg.OnMouseDown,
		OnMouseMove:      cfg.OnMouseMove,
		OnMouseUp:        cfg.OnMouseUp,
		OnSizeChanged:    cfg.OnSizeChanged,
		Layout:           HBox{MarginsZero: true, SpacingZero: true},
		Children: []Widget{
			ScrollView{
				HorizontalFixed: true,
				Layout:          VBox{MarginsZero: true},
				Children: []Widget{
					Composite{
						Layout: VBox{MarginsZero: true},
						Children: []Widget{
							ToolBar{
								AssignTo:    &mpmw.navTB,
								Orientation: Vertical,
								ButtonStyle: ToolBarButtonImageAboveText,
								MaxTextRows: 2,
							},
						},
					},
				},
			},
			Composite{
				AssignTo: &mpmw.pageCom,
				Name:     "pageCom",
				Layout:   HBox{MarginsZero: true, SpacingZero: true},
			},
		},
	}).Create(); err != nil {
		return nil, err
	}

	succeeded := false
	defer func() {
		if !succeeded {
			mpmw.Dispose()
		}
	}()

	for _, pc := range cfg.PageCfgs {
		action, err := mpmw.newPageAction(pc.Title, pc.Image, pc.NewPage)
		if err != nil {
			return nil, err
		}

		mpmw.pageActions = append(mpmw.pageActions, action)
	}

	if err := mpmw.updateNavigationToolBar(); err != nil {
		return nil, err
	}

	if len(mpmw.pageActions) > 0 {
		if err := mpmw.setCurrentAction(mpmw.pageActions[0]); err != nil {
			return nil, err
		}
	}

	if cfg.OnCurrentPageChanged != nil {
		mpmw.CurrentPageChanged().Attach(cfg.OnCurrentPageChanged)
	}

	succeeded = true

	return mpmw, nil
}

func (mpmw *MultiPageMainWindow) CurrentPage() Page {
	return mpmw.currentPage
}

func (mpmw *MultiPageMainWindow) CurrentPageTitle() string {
	if mpmw.currentAction == nil {
		return ""
	}

	return mpmw.currentAction.Text()
}

func (mpmw *MultiPageMainWindow) CurrentPageChanged() *walk.Event {
	return mpmw.currentPageChangedPublisher.Event()
}

func (mpmw *MultiPageMainWindow) newPageAction(title, imageData string, newPage PageFactoryFunc) (*walk.Action, error) {
	src := base64.NewDecoder(base64.StdEncoding, strings.NewReader(imageData))
	pngImg, err := png.Decode(src)
	if err != nil {
		return nil, err
	}
	walkImg, err := walk.NewBitmapFromImageForDPI(pngImg,96)
	if err != nil {
		return nil, err
	}
	img, err := walk.NewIconFromImageWithSize(walkImg, walk.Size{64,64})
	if err != nil {
		return nil, err
	}

	action := walk.NewAction()
	action.SetCheckable(true)
	action.SetExclusive(true)
	action.SetImage(img)
	action.SetText(title)

	mpmw.action2NewPage[action] = newPage

	action.Triggered().Attach(func() {
		mpmw.setCurrentAction(action)
	})

	return action, nil
}

func (mpmw *MultiPageMainWindow) setCurrentAction(action *walk.Action) error {
	defer func() {
		if !mpmw.pageCom.IsDisposed() {
			mpmw.pageCom.RestoreState()
		}
	}()

	mpmw.SetFocus()

	if prevPage := mpmw.currentPage; prevPage != nil {
		mpmw.pageCom.SaveState()
		prevPage.SetVisible(false)
		prevPage.(walk.Widget).SetParent(nil)
		prevPage.Dispose()
	}

	newPage := mpmw.action2NewPage[action]

	page, err := newPage(mpmw.pageCom, mpmw.TextToSign)
	if err != nil {
		return err
	}

	font := page.Font()
	newFont, err := walk.NewFont(font.Family(), font.PointSize()+1, font.Style())
	if err == nil {
		page.SetFont(newFont)
	}

	for _, a := range mpmw.pageActions {
		a.SetChecked(false)
	}
	action.SetChecked(true)

	mpmw.currentPage = page
	mpmw.currentAction = action

	mpmw.currentPageChangedPublisher.Publish()

	return nil
}

func (mpmw *MultiPageMainWindow) updateNavigationToolBar() error {
	mpmw.navTB.SetSuspended(true)
	defer mpmw.navTB.SetSuspended(false)

	actions := mpmw.navTB.Actions()

	if err := actions.Clear(); err != nil {
		return err
	}

	for _, action := range mpmw.pageActions {
		if err := actions.Add(action); err != nil {
			return err
		}
	}

	if mpmw.currentAction != nil {
		if !actions.Contains(mpmw.currentAction) {
			for _, action := range mpmw.pageActions {
				if action != mpmw.currentAction {
					if err := mpmw.setCurrentAction(action); err != nil {
						return err
					}

					break
				}
			}
		}
	}

	lParam := uintptr(win.MAKELONG(uint16(120), 120))
	mpmw.navTB.SendMessage(win.TB_SETBUTTONSIZE, 0, lParam)
	//lParam = uintptr(win.MAKELONG(uint16(64), 64))
	//mpmw.navTB.SendMessage(win.TB_SETBITMAPSIZE, 0, lParam)

	return nil
}
