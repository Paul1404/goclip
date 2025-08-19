//go:build linux

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bendahl/uinput"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type windowInfo struct {
	ID    string
	Title string
}

var backend string // "x11" or "uinput"

// detectBackend checks if we are on X11 or Wayland
func detectBackend() string {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		return "uinput" // Wayland → use uinput
	}
	if os.Getenv("DISPLAY") != "" {
		return "x11"
	}
	return "uinput" // fallback
}

// listWindows returns visible windows (X11 only)
func listWindows() ([]windowInfo, error) {
	if backend != "x11" {
		return nil, fmt.Errorf("window listing not supported on Wayland")
	}

	cmd := exec.Command("xdotool", "search", "--onlyvisible", "--name", ".")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var wins []windowInfo
	for _, id := range lines {
		if id == "" {
			continue
		}
		nameCmd := exec.Command("xdotool", "getwindowname", id)
		var buf bytes.Buffer
		nameCmd.Stdout = &buf
		if err := nameCmd.Run(); err != nil {
			continue
		}
		title := strings.TrimSpace(buf.String())
		if title != "" {
			wins = append(wins, windowInfo{ID: id, Title: title})
		}
	}

	sort.Slice(wins, func(i, j int) bool {
		return strings.ToLower(wins[i].Title) < strings.ToLower(wins[j].Title)
	})

	return wins, nil
}

// typeText sends text to a window (X11) or globally (uinput)
func typeText(windowID, text string) error {
	if backend == "x11" {
		if windowID == "" {
			return exec.Command("xdotool", "type", "--delay", "7", text).Run()
		}
		exec.Command("xdotool", "windowactivate", "--sync", windowID).Run()
		return exec.Command("xdotool", "type", "--window", windowID, "--delay", "7", text).Run()
	}

	if backend == "uinput" {
		return typeWithUinput(text)
	}

	return fmt.Errorf("unsupported backend")
}

// typeWithUinput injects keystrokes via /dev/uinput
func typeWithUinput(text string) error {
	keyboard, err := uinput.CreateKeyboard("/dev/uinput", []byte("goclip-virtual-keyboard"))
	if err != nil {
		return fmt.Errorf("failed to create uinput keyboard: %w", err)
	}
	defer keyboard.Close()

	for _, r := range text {
		key, shift := runeToKey(r)
		if key == 0 {
			continue
		}
		if shift {
			_ = keyboard.KeyDown(uinput.KeyLeftshift)
		}
		if err := keyboard.KeyPress(key); err != nil {
			return err
		}
		if shift {
			_ = keyboard.KeyUp(uinput.KeyLeftshift)
		}
		time.Sleep(7 * time.Millisecond)
	}
	return nil
}

// runeToKey maps runes to uinput key codes (basic ASCII)
func runeToKey(r rune) (int, bool) {
	switch r {
	case 'a', 'A':
		return uinput.KeyA, r == 'A'
	case 'b', 'B':
		return uinput.KeyB, r == 'B'
	case 'c', 'C':
		return uinput.KeyC, r == 'C'
	case 'd', 'D':
		return uinput.KeyD, r == 'D'
	case 'e', 'E':
		return uinput.KeyE, r == 'E'
	case 'f', 'F':
		return uinput.KeyF, r == 'F'
	case 'g', 'G':
		return uinput.KeyG, r == 'G'
	case 'h', 'H':
		return uinput.KeyH, r == 'H'
	case 'i', 'I':
		return uinput.KeyI, r == 'I'
	case 'j', 'J':
		return uinput.KeyJ, r == 'J'
	case 'k', 'K':
		return uinput.KeyK, r == 'K'
	case 'l', 'L':
		return uinput.KeyL, r == 'L'
	case 'm', 'M':
		return uinput.KeyM, r == 'M'
	case 'n', 'N':
		return uinput.KeyN, r == 'N'
	case 'o', 'O':
		return uinput.KeyO, r == 'O'
	case 'p', 'P':
		return uinput.KeyP, r == 'P'
	case 'q', 'Q':
		return uinput.KeyQ, r == 'Q'
	case 'r', 'R':
		return uinput.KeyR, r == 'R'
	case 's', 'S':
		return uinput.KeyS, r == 'S'
	case 't', 'T':
		return uinput.KeyT, r == 'T'
	case 'u', 'U':
		return uinput.KeyU, r == 'U'
	case 'v', 'V':
		return uinput.KeyV, r == 'V'
	case 'w', 'W':
		return uinput.KeyW, r == 'W'
	case 'x', 'X':
		return uinput.KeyX, r == 'X'
	case 'y', 'Y':
		return uinput.KeyY, r == 'Y'
	case 'z', 'Z':
		return uinput.KeyZ, r == 'Z'
	case ' ':
		return uinput.KeySpace, false
	case '\n':
		return uinput.KeyEnter, false
	case '0':
		return uinput.Key0, false
	case '1':
		return uinput.Key1, false
	case '2':
		return uinput.Key2, false
	case '3':
		return uinput.Key3, false
	case '4':
		return uinput.Key4, false
	case '5':
		return uinput.Key5, false
	case '6':
		return uinput.Key6, false
	case '7':
		return uinput.Key7, false
	case '8':
		return uinput.Key8, false
	case '9':
		return uinput.Key9, false
	default:
		return 0, false
	}
}

func main() {
	backend = detectBackend()
	log.Println("Detected backend:", backend)

	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	w := myApp.NewWindow("goclip (Linux)")
	w.Resize(fyne.NewSize(800, 460))

	inputEntry := widget.NewMultiLineEntry()
	inputEntry.SetPlaceHolder("Type here…")
	inputEntry.Wrapping = fyne.TextWrapWord

	status := widget.NewLabel("Ready.")

	windowOptions := []string{}
	windowMap := map[string]string{}
	windowSelect := widget.NewSelect(windowOptions, nil)
	windowSelect.PlaceHolder = "None (use last active)"

	var laMu sync.RWMutex
	lastActiveTitle := "(none)"
	lastActiveID := ""
	lastActiveLabel := widget.NewLabel("Last active: (none)")

	refreshBtn := widget.NewButton("Refresh windows", func() {
		if backend != "x11" {
			status.SetText("Window listing not supported on Wayland (uinput mode).")
			return
		}
		wins, err := listWindows()
		if err != nil {
			status.SetText("Error listing windows: " + err.Error())
			return
		}
		windowOptions = []string{}
		windowMap = map[string]string{}
		for _, wi := range wins {
			label := fmt.Sprintf("%s (%s)", truncateRunes(wi.Title, 30), wi.ID)
			windowOptions = append(windowOptions, label)
			windowMap[label] = wi.ID
		}
		windowSelect.Options = windowOptions
		windowSelect.Refresh()
		status.SetText(fmt.Sprintf("Found %d windows.", len(wins)))
	})

	// Track last active window (X11 only)
	if backend == "x11" {
		go func() {
			for {
				cmd := exec.Command("xdotool", "getactivewindow")
				out, err := cmd.Output()
				if err == nil {
					id := strings.TrimSpace(string(out))
					if id != "" {
						nameCmd := exec.Command("xdotool", "getwindowname", id)
						var buf bytes.Buffer
						nameCmd.Stdout = &buf
						if err := nameCmd.Run(); err == nil {
							title := strings.TrimSpace(buf.String())
							if title != "" {
								laMu.Lock()
								lastActiveID = id
								lastActiveTitle = truncateRunes(title, 30)
								laMu.Unlock()
								lastActiveLabel.SetText("Last active: " + lastActiveTitle)
							}
						}
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
	}

	typeBtn := widget.NewButton("Type", func() {
		txt := inputEntry.Text
		if txt == "" {
			status.SetText("Nothing to type.")
			return
		}

		laMu.RLock()
		curID := lastActiveID
		curTitle := lastActiveTitle
		laMu.RUnlock()

		var targetID string
		if windowSelect.Selected == "" {
			targetID = curID
		} else {
			var ok bool
			targetID, ok = windowMap[windowSelect.Selected]
			if !ok {
				status.SetText("Selected window no longer available.")
				return
			}
		}

		if backend == "uinput" {
			// Wayland: ignore window selection, just type globally
			targetID = ""
			curTitle = "(focused window)"
		}

		if err := typeText(targetID, txt); err != nil {
			status.SetText("Error typing: " + err.Error())
			return
		}

		status.SetText("Typed to: " + curTitle)
	})

	left := container.NewVBox(
		widget.NewLabelWithStyle("Target Window", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewHBox(windowSelect, refreshBtn),
		lastActiveLabel,
	)

	body := container.NewVBox(
		widget.NewLabelWithStyle("Text to type", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		inputEntry,
		container.NewHBox(typeBtn),
		status,
	)

	content := container.NewBorder(left, nil, nil, nil, body)
	w.SetContent(content)

	if backend == "x11" {
		refreshBtn.OnTapped() // initial load
	}
	w.ShowAndRun()
}

func truncateRunes(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) <= n {
		return s
	}
	if n <= 3 {
		return string(r[:n])
	}
	return string(r[:n]) + "..."
}