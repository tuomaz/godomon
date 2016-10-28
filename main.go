// +build !js

package main

import (
	"fmt"
	"os"
	"time"
	"net/http"
"runtime"
"runtime/pprof"


	"github.com/pquerna/ffjson/ffjson"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_ttf"
	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

import _ "net/http/pprof"

//DomoticzsData data from Domoticz
type DomoticzsData struct {
	ID      string `json:"id"`
	Svalue1 string `json:"svalue1"`
	Name    string `json:"name"`
}

var blowup bool
var premult bool
var temperature string
var tempKitchen string
var winTitle string = "Text"
var winWidth, winHeight int = 1920, 1200

func main() {
go func() {
	fmt.Println(http.ListenAndServe(":6060", nil))
}()
	f, err := os.Create("mem.prof")
        if err != nil {
            fmt.Printf("could not create memory profile: ", err)
        }
        runtime.GC() // get up-to-date statistics
        if err := pprof.WriteHeapProfile(f); err != nil {
            fmt.Printf("could not write memory profile: ", err)
        }
        f.Close()

	fmt.Printf("Starting up...")

	// Create an MQTT Client.
	cli := client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			fmt.Println(err)
		},
	})

	// Terminate the Client.
	defer cli.Terminate()

	// Connect to the MQTT Server.
	err = cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  "192.168.1.5:1883",
		ClientID: []byte("go-domoticz-monitor-2"),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("Connected to server")

	// Subscribe to topics.
	err = cli.Subscribe(&client.SubscribeOptions{
		SubReqs: []*client.SubReq{
			&client.SubReq{
				TopicFilter: []byte("domoticz/out"),
				QoS:         mqtt.QoS0,
				// Define the processing of the message handler.
				Handler: func(topicName, message []byte) {
					dd := &DomoticzsData{}
					err := ffjson.Unmarshal(message, dd)
					if err != nil {
						fmt.Printf("Failed to unmarshall data %v\n", err)
					}
					fmt.Printf("ID = %s, svalue1 = %s, name = %s\n", dd.ID, dd.Svalue1, dd.Name)
					if dd.ID == "3585" {
						temperature = dd.Svalue1
					}
					if dd.ID == "260" {
						tempKitchen = dd.Svalue1
					}
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	status := run()
	fmt.Printf("Status is %d\n", status)

	// Disconnect the Network Connection.
	if err := cli.Disconnect(); err != nil {
		panic(err)
	}
}

func run() int {
	var window *sdl.Window
	var font *ttf.Font
	var clock *sdl.Surface
	var out *sdl.Surface
	var kitchen *sdl.Surface
	var err error
	var renderer *sdl.Renderer
	var clockTexture *sdl.Texture
	var outTexture *sdl.Texture
	var kitchenTexture *sdl.Texture
	var run bool
	run = true

	sdl.Init(sdl.INIT_EVERYTHING)

	if err := ttf.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize TTF: %s\n", err)
		return 1
	}

	if window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_FULLSCREEN_DESKTOP); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return 2
	}
	defer window.Destroy()

	if renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create renderer: %s\n", err)
	}
	defer renderer.Destroy()

	if font, err = ttf.OpenFont("Ubuntu-R.ttf", 144); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open font: %s\n", err)
		return 4
	}

	sdl.ShowCursor(sdl.DISABLE)

//MainLoop:
	for run {
/*		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				run = false
				break MainLoop
			case *sdl.KeyDownEvent:
				run = false
				break MainLoop

			}

		}*/
		t := time.Now()

		if clock, err = font.RenderUTF8_Blended(t.Format("15.04"), sdl.Color{255, 255, 255, 255}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to render text: %s\n", err)
			return 5
		}
		defer clock.Free()

		if out, err = font.RenderUTF8_Blended(temperature+"°", sdl.Color{255, 255, 255, 255}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to render text: %s\n", err)
			return 5
		}
		defer out.Free()

		if kitchen, err = font.RenderUTF8_Blended(tempKitchen+"°", sdl.Color{255, 255, 255, 255}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to render text: %s\n", err)
			return 5
		}
		defer kitchen.Free()

		if clockTexture, err = renderer.CreateTextureFromSurface(clock); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create texture from surface: %s\n", err)
			return 8
		}

		if outTexture, err = renderer.CreateTextureFromSurface(out); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create texture from surface: %s\n", err)
			return 8
		}

		if kitchenTexture, err = renderer.CreateTextureFromSurface(kitchen); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create texture from surface: %s\n", err)
			return 8
		}

		r3 := &sdl.Rect{
			H: clock.H,
			W: clock.W,
			X: 100,
			Y: 100,
		}

		outRect := &sdl.Rect{
			H: out.H,
			W: out.W,
			X: 100,
			Y: 300,
		}

		kitchenRect := &sdl.Rect{
			H: kitchen.H,
			W: kitchen.W,
			X: 100,
			Y: 500,
		}

		renderer.Clear()
		renderer.Copy(clockTexture, nil, r3)
		renderer.Copy(outTexture, nil, outRect)
		renderer.Copy(kitchenTexture, nil, kitchenRect)
		renderer.Present()
		sdl.Delay(10000)
	}
	font.Close()

	return 0

}
