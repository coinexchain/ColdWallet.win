package main

import (
	"bytes"
	"errors"
	"image/png"
	"sync/atomic"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	qrenc "github.com/skip2/go-qrcode"
	"gocv.io/x/gocv"
)

// prompt user to enter a passphrase
func ShowPassphraseDialog(owner walk.Form, okCallback func(pass string)) {
	var dlg *walk.Dialog
	var okPB, cancelPB *walk.PushButton
	var passLineEdit *walk.LineEdit

	var dialog = Dialog{}
	dialog.AssignTo = &dlg
	dialog.Title = T("enterEncryptPassphrase")
	dialog.MinSize = Size{300, 200}
	dialog.Layout = VBox{}
	dialog.DefaultButton = &okPB
	dialog.CancelButton = &cancelPB

	childrens := []Widget{
		Composite{
			Layout: Grid{Columns: 2},
			Children: []Widget{
				Label{Text: T("enterEncryptPassphrase")},
				LineEdit{
					AssignTo: &passLineEdit,
					PasswordMode: true,
				},
			},
		},
		Composite{
			Layout: HBox{},
			Children: []Widget{
				HSpacer{},
				PushButton{
					AssignTo: &okPB,
					Text:     T("ok"),
					OnClicked: func() {
						okCallback(passLineEdit.Text())
						dlg.Accept()
					},
				},
				PushButton{
					AssignTo:  &cancelPB,
					Text:      T("cancel"),
					OnClicked: func() { dlg.Cancel() },
				},
			},
		},
	}
	dialog.Children = childrens
	dialog.Run(owner)
}

// prompt user to enter the old passphrase and re-type the new passphrase twice
func ShowChangePassphraseDialog(owner walk.Form, okCallback func(oldPass, newPass string)) {
	var dlg *walk.Dialog
	var okPB, cancelPB *walk.PushButton
	var passOldLineEdit, passNew1LineEdit, passNew2LineEdit *walk.LineEdit

	var dialog = Dialog{}
	dialog.AssignTo = &dlg
	dialog.Title = T("changeEncryptPassphrase")
	dialog.MinSize = Size{300, 200}
	dialog.Layout = VBox{}
	dialog.DefaultButton = &okPB
	dialog.CancelButton = &cancelPB

	childrens := []Widget{
		Composite{
			Layout: Grid{Columns: 2},
			Children: []Widget{
				Label{Text: T("enterOldEncryptPassphrase")},
				LineEdit{
					AssignTo: &passOldLineEdit,
					PasswordMode: true,
				},
				Label{Text: T("belowNewEncryptPassphrase")},
				Label{},
				Label{Text: T("enterEncryptPassphrase")},
				LineEdit{
					AssignTo: &passNew1LineEdit,
					PasswordMode: true,
				},
				Label{Text: T("retypeEncryptPassphrase")},
				LineEdit{
					AssignTo: &passNew2LineEdit,
					PasswordMode: true,
				},
			},
		},
		Composite{
			Layout: HBox{},
			Children: []Widget{
				HSpacer{},
				PushButton{
					AssignTo: &okPB,
					Text:     T("ok"),
					OnClicked: func() {
						pass1 := passNew1LineEdit.Text()
						pass2 := passNew2LineEdit.Text()
						oldPass := passOldLineEdit.Text()
						if pass1 != pass2 {
							walk.MsgBox(MainWin, T("error!"), T("mismatchPassphrase"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
							return
						}
						okCallback(oldPass, pass1)
						dlg.Accept()
					},
				},
				PushButton{
					AssignTo:  &cancelPB,
					Text:      T("cancel"),
					OnClicked: func() { dlg.Cancel() },
				},
			},
		},
	}
	dialog.Children = childrens
	dialog.Run(owner)
}

// continuously read images from webcam and scan QRCode in the images
// When a valid QRCode is obtained, okCallback will be invoked with it and dlg will be closed
func runJob(webcam *gocv.VideoCapture, imageView *walk.ImageView, mat *gocv.Mat,
	okCallback func(text string), dlg *walk.Dialog, running *int32) {
	for {
		if atomic.LoadInt32(running) == 0 {
			return
		}
		//println("haha1")
		if ok := webcam.Read(mat); !ok {
			continue
		}
		//println("haha2")
		if mat.Empty() {
			continue
		}
		img, err := mat.ToImage()
		//println("haha3")
		if err != nil {
			continue
		}
		bitmap, err := walk.NewBitmapFromImage(img)
		//println("haha4")
		if err != nil {
			continue
		}
		MainWin.Synchronize(func() {
			imageView.SetImage(bitmap)
		})
		//println("haha5")
		bmp, err := gozxing.NewBinaryBitmapFromImage(img)
		if err != nil {
			continue
		}
		// decode image
		//println("haha6")
		qrReader := qrcode.NewQRCodeReader()
		result, err := qrReader.Decode(bmp, nil)
		if err == nil {
			okCallback(result.String())
			dlg.Cancel()
			return
		}
		 time.Sleep(200 * time.Millisecond)
	}
}

// Scan all the webcams to get the one with highest resolution
func getBestWebcam() (*gocv.VideoCapture, error) {
	bestID := -1
	largestHeight := -1
	for i := 0; i < 10; i++ {
		webcam, err := gocv.OpenVideoCapture(i)
		if err != nil {
			break
		}
		webcam.Set(gocv.VideoCaptureFrameWidth, 1920)
		webcam.Set(gocv.VideoCaptureFrameHeight, 1080)
		h := int(webcam.Get(gocv.VideoCaptureFrameHeight))
		if h > largestHeight {
			largestHeight = h
			bestID = i
		}
		webcam.Close()
	}
	if bestID == -1 {
		return nil, errors.New("No webcam was found!")
	}
	webcam, err := gocv.OpenVideoCapture(bestID)
	if err == nil {
		webcam.Set(gocv.VideoCaptureFrameWidth, 1920)
		webcam.Set(gocv.VideoCaptureFrameHeight, 1080)
	}
	return webcam, err
}

// Show a dialog for scanning QRCode
func ShowQRCodeScanDialog(owner walk.Form, okCallback func(text string)) {
	var dlg *walk.Dialog
	var cancelPB *walk.PushButton
	var imageView *walk.ImageView

	webcam, err := getBestWebcam()
	if err != nil {
		walk.MsgBox(MainWin, T("error!"), T("failOpenWebcam"), walk.MsgBoxIconError|walk.MsgBoxApplModal)
		return
	}
	mat := gocv.NewMat()
	running := int32(1)

	h := int(webcam.Get(gocv.VideoCaptureFrameHeight))
	w := int(webcam.Get(gocv.VideoCaptureFrameWidth))

	dialog := Dialog{}
	dialog.AssignTo = &dlg
	dialog.Title = T("scanQRCode")
	dialog.MinSize = Size{w+30, h+100}
	dialog.Layout = VBox{}
	dialog.CancelButton = &cancelPB

	childrens := []Widget{
		ImageView{
			AssignTo:   &imageView,
			Margin:     10,
			Mode:       ImageViewModeStretch,
		},
		PushButton{
			AssignTo:  &cancelPB,
			Text:      T("cancel"),
			MinSize:   Size{100, 50},
			OnClicked: func() { dlg.Cancel() },
		},
	}
	dialog.Children = childrens
	dialog.Create(owner)
	dlg.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		atomic.AddInt32(&running, -1)
		webcam.Close()
		mat.Close()
	})
	imageView.SetDoubleBuffering(true)
	dlg.SetDoubleBuffering(true)
	go runJob(webcam, imageView, &mat, okCallback, dlg, &running)
	dlg.Run()
}

// Shows a QRCode on screen
func ShowQRCodeDialog(owner walk.Form, text, title, hint string) error {
	var dlg *walk.Dialog
	var okPB *walk.PushButton
	var imageView *walk.ImageView

	bz, err := qrenc.Encode(text, qrenc.Medium, 512)
	if err != nil {
		return err
	}
	pngImg, err := png.Decode(bytes.NewReader(bz))
	if err != nil {
		return err
	}
	bitmap, err := walk.NewBitmapFromImage(pngImg)
	if err != nil {
		return err
	}

	dialog := Dialog{}
	dialog.AssignTo = &dlg
	dialog.Title = title
	dialog.MinSize = Size{800, 600}
	dialog.Layout = VBox{}
	dialog.DefaultButton = &okPB

	childrens := []Widget{
		Label{Text: hint},
		ImageView{
			AssignTo:   &imageView,
			Margin:     10,
			Mode:       ImageViewModeStretch,
		},
		PushButton{
			AssignTo:  &okPB,
			Text:      T("ok"),
			OnClicked: func() { dlg.Accept() },
		},
	}
	dialog.Children = childrens
	dialog.Create(owner)
	imageView.SetImage(bitmap)
	dlg.Run()
	return nil
}

