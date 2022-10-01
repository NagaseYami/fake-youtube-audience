package main

import (
	"flag"
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
	"github.com/op/go-logging"
	"os"
	"path"
	"time"
)

func main() {

	// Get flags
	homedir, _ := os.UserHomeDir()
	if homedir == "" {
		homedir = "test.jpg"
	}
	var streamURL = flag.String("url", "", "Youtube stream URL.")
	var timeout = flag.Int("timeout", 30, "Timeout in Second. Default is 30.")
	var debug = flag.Bool("debug", false, "Debug mode switch.")
	var debugScreenShotPath = flag.String("debug-screenshot-path", homedir, "Debug mode screenshot path.")
	var browserPath = flag.String("browser", "", "Chrome or Edge browser executable file path.")
	flag.Parse()

	// Init logger
	log := logging.MustGetLogger("main")
	format := logging.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)
	backend := logging.NewLogBackend(os.Stderr, "", 0)
	formatted := logging.NewBackendFormatter(backend, format)
	logging.SetBackend(formatted)
	log.Debug("Log initialization Finished.")

	if *streamURL == "" {
		log.Fatal("Please set URL.")
	}

	// Init Chrome
	var has bool
	if *browserPath == "" {
		*browserPath, has = launcher.LookPath()
		if !has {
			log.Fatal("Please set browser path.")
		}
	}

	u, err := launcher.New().Bin(*browserPath).Launch()
	if err != nil {
		log.Fatal(err)
	}
	browser := rod.New().ControlURL(u)
	err = browser.Connect()
	if err != nil {
		log.Fatal(err)
	}
	closeBrowser := func() {
		err = browser.Close()
		if err != nil {
			log.Fatal(err)
		}

	}
	defer closeBrowser()
	log.Debug("Chrome initialization Finished.")

	if *debug {
		err = browser.SetCookies(nil)
		if err != nil {
			log.Fatal(err)
		}
	}
	log.Debug("Cookies cleared.")

	// Open page
	page, err := browser.Page(proto.TargetCreateTarget{URL: *streamURL})
	if err != nil {
		log.Fatal(err)
	}
	page.MustWaitLoad()
	log.Debugf("Page %s Opened.\n", *streamURL)
	if *debug {
		page.MustScreenshot(path.Join(*debugScreenShotPath, "Screenshot1.jpg"))
	}

	// Play Video
	pageLoaded, isPlaying, playingAd := false, false, true
	timeoutCount := 0
	for !pageLoaded || !isPlaying || playingAd {
		if timeoutCount >= *timeout {
			log.Fatal("Timeout. Can't play stream.")
		}
		time.Sleep(time.Second)
		timeoutCount++

		if !pageLoaded {
			hasVideoContainer, _, err := page.Has(".html5-video-container")
			if err != nil {
				log.Fatal(err)
			}
			if !hasVideoContainer {
				log.Warning("Can't find video container element. Retrying...")
				continue
			}
			pageLoaded = true
		}

		if !isPlaying {
			hasPlayBtn, playBtn, err := page.Has(".ytp-play-button")
			if err != nil {
				log.Fatal(err)
			}
			if !hasPlayBtn {
				log.Warning("Can't find play button element. Retrying...")
				continue
			}

			tooltip, err := playBtn.Attribute("data-title-no-tooltip")
			if err != nil {
				log.Fatal("Can't find attribute data-title-no-tooltip from play button. It's impossible.")
			}
			if *tooltip == "Play" {
				err = playBtn.Click(proto.InputMouseButtonLeft, 1)
				log.Info("Video is paused. Try to play video...")
				if err != nil {
					log.Fatal(err)
				}
			}
			isPlaying = true
		}

		if playingAd {
			hasAds, _, err := page.Has(".ytp-ad-player-overlay > div")
			if err != nil {
				log.Fatal(err)
			}
			hasSkipBtn, skipBtn, err := page.Has(".ytp-ad-skip-button")
			if err != nil {
				log.Fatal(err)
			}

			if !hasAds {
				log.Info("No ads.")
				playingAd = false
			} else if hasSkipBtn {
				if *debug {
					page.MustScreenshot(path.Join(*debugScreenShotPath, fmt.Sprintf("Screenshot2-%d.jpg", timeoutCount+1)))
				}
				err = skipBtn.Click(proto.InputMouseButtonLeft, 1)
				log.Info("Found ads. Try to skip...")
				if err != nil {
					log.Error(err)
				}
			} else if !hasSkipBtn {
				if *debug {
					page.MustScreenshot(path.Join(*debugScreenShotPath, fmt.Sprintf("Screenshot2-%d.jpg", timeoutCount+1)))
				}
				log.Warning("Found ads but can't skip. Waiting for 10 sec...")
				time.Sleep(10 * time.Second)
			}
		}
	}

	if *debug {
		time.Sleep(5 * time.Second)
		page.MustScreenshot(path.Join(*debugScreenShotPath, "Screenshot3.jpg"))
	} else {
		utils.Pause()
	}
}
