package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/shurcooL/trayhost"
	"github.com/tmc/icns"
)

func main() {

	// On macOS, when you run an app bundle, the working directory of the executed process
	// is the root directory (/), not the app bundle's Contents/Resources directory.
	// Change directory to Resources so that we can load resources from there.
	ep, err := os.Executable()
	if err != nil {
		log.Fatalln("os.Executable:", err)
	}
	err = os.Chdir(filepath.Join(filepath.Dir(ep), "..", "Resources"))
	if err != nil {
		log.Fatalln("os.Chdir:", err)
	}

	// Load tray icon.
	iconData, err := ioutil.ReadFile("AppIcon.icns")
	if err != nil {
		log.Fatalln(err)
	}
	icons, err := icns.Parse(bytes.NewReader(iconData))
	if err != nil {
		log.Fatalln(err)
	}

	menuItems := []trayhost.MenuItem{
		{
			Title: "Example Item",
			Handler: func() {
				fmt.Println("do stuff")
			},
		},
		{
			Title: "Get Clipboard Content",
			Handler: func() {
				cc, err := trayhost.GetClipboardContent()
				if err != nil {
					fmt.Printf("GetClipboardContent() error: %v\n", err)
					return
				}

				fmt.Printf("Text: %q\n", cc.Text)
				fmt.Printf("Image: %v len(%v)\n", cc.Image.Kind, len(cc.Image.Bytes))
				fmt.Printf("Files: len(%v) %v\n", len(cc.Files), cc.Files)
			},
		},
		{
			Title: "Set Clipboard Text",
			Handler: func() {
				const text = "this text gets copied"

				trayhost.SetClipboardText(text)
				fmt.Printf("Text %q got copied into your clipboard.\n", text)
			},
		},
		{
			// Displaying notifications requires a proper app bundle and won't work without one.
			// See https://godoc.org/github.com/shurcooL/trayhost#hdr-Notes.

			Title: "Display Notification",
			Handler: func() {
				notification := trayhost.Notification{
					Title:   "Example Notification",
					Body:    "Notification body text is here.",
					Timeout: 3 * time.Second,
					Image: trayhost.Image{
						Kind:  "png",
						Bytes: icons[len(icons)-1].Data,
					},
					Handler: func() {
						fmt.Println("do stuff when notification is clicked")
					},
				}
				if cc, err := trayhost.GetClipboardContent(); err == nil && cc.Image.Kind != "" {
					// Use image from clipboard as notification image.
					notification.Image = cc.Image
				}
				notification.Display()
			},
		},
		trayhost.SeparatorMenuItem(),
		{
			Title:   "Quit",
			Handler: trayhost.Exit,
		},
	}

	trayhost.Initialize("Example App", icons[len(icons)-1].Data, menuItems)

	go func() {
		for {
			time.Sleep(2 * time.Second)

			myimage := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{256, 256}})

			// This loop just fills the image with random data
			c := color.NRGBA{uint8(0), uint8(rand.Intn(155) + 50), uint8(rand.Intn(55)), 255}
			for x := 0; x < 256; x++ {
				for y := 0; y < 256; y++ {
					myimage.SetNRGBA(x, y, c)
				}
			}

			buf := new(bytes.Buffer)
			png.Encode(buf, myimage)
			trayhost.SetStatusIcon(buf.Bytes())

			time.Sleep(2 * time.Second)
			trayhost.SetStatusIcon(icons[len(icons)-1].Data)
		}
	}()

	trayhost.EnterLoop()
}
