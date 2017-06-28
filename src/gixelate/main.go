package main

import (
	"image"
	"image/color"
	"log"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xgraphics"
	"github.com/BurntSushi/xgbutil/xwindow"

	"github.com/disintegration/imaging"

	"github.com/oliamb/cutter"

	"os/exec"
)

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatal(err)
	}

	// Use the "NewDrawable" constructor to create an xgraphics.Image value
	// from a drawable. (Usually this is done with pixmaps, but drawables
	// can also be windows.)
	ximg, err := xgraphics.NewDrawable(X, xproto.Drawable(X.RootWin()))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("grabbed root screen")

	oimg := imaging.New(ximg.Rect.Max.X, ximg.Rect.Max.Y, color.NRGBA{0, 0, 0, 0})
	oimg = imaging.Paste(oimg, ximg, image.Pt(0, 0))

	clients, err := ewmh.ClientListGet(X)
	// Shows the screenshot in a window.
	// ximg.XShowExtra("Screenshot", true)
	// Iterate through each client, find its name and find its size.
	for _, clientid := range clients {
		dgeom, err := xwindow.New(X, clientid).DecorGeometry()
		if err != nil {
			log.Fatalf("Could not get geometry for (0x%X) because: %s",
				clientid, err)
			continue
		}

		attr, err := xproto.GetWindowAttributes(X.Conn(), clientid).Reply()
		if err != nil {
			log.Printf("ERROR: %s", err)
			continue
		}
		if attr.MapState != xproto.MapStateViewable {
			continue
		}

		// Use cutter.Crop since imaging.Crop is painfully slow.
		wimg, _ := cutter.Crop(ximg, cutter.Config{
			Mode:   cutter.TopLeft,
			Anchor: image.Point{X: dgeom.X(), Y: dgeom.Y()},
			Width:  dgeom.Width(),
			Height: dgeom.Height(),
		})

		wimg = imaging.Resize(wimg, dgeom.Width()/9, dgeom.Height()/9, imaging.NearestNeighbor)
		wimg = imaging.Resize(wimg, dgeom.Width(), dgeom.Height(), imaging.Linear)

		oimg = imaging.Paste(oimg, wimg, image.Pt(dgeom.X(), dgeom.Y()))
		log.Printf("processed %d", clientid)
	}

	oimg = imaging.Blur(oimg, 0.25)
	oimg = imaging.Sharpen(oimg, 0.75)
	oimg = imaging.AdjustBrightness(oimg, 0.5)

	log.Printf("saving screen ...")
	imaging.Save(oimg, "/tmp/screen.png")
	log.Printf("saved!")

	cmd := exec.Command("i3lock", "-i", "/tmp/screen.png")
	cmd.Run()
}
