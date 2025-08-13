# goclip

A tiny Windows tool that types text into **any** focused window (even web/VNC/VM consoles) using **real keyboard events**.  
Built with [Fyne](https://fyne.io/) for a clean dark-mode GUI.

---

## Why?

Some apps and browser-embedded consoles (e.g. VMware/KVM) ignore Unicode paste or `WM_CHAR` messages. **goclip** simulates **physical key presses (scan codes)**, so those consoles receive input exactly like a real keyboard would.

---

## Features

- 🖱️ **Target window selection** from a dropdown  
  - Or click **Clear** → nothing selected means **“use last active window”** automatically.
- ⌨️ **Layout-aware typing** using OS keyboard layouts (via `VkKeyScanExW`/`MapVirtualKeyExW`)  
  - Sends **scan codes** with modifiers (Shift/Ctrl/Alt) for each character.
  - **Unicode fallback** for unmappable characters.
- 🕶️ **Modern dark-mode GUI** (Fyne)
- ⚙️ **No install required** – single `.exe`

---

## Supported keyboard layouts (selector)

- Auto (Use System)
- English (US)
- US International
- English (UK)
- German (DE)
- French (FR)
- Spanish (ES)
- Italian (IT)
- Dutch (NL)
- Portuguese (BR - ABNT2)
- Portuguese (PT)
- Danish (DA)
- Swedish (SV)
- Finnish (FI)
- Norwegian (NO)
- Swiss German (DE-CH)
- Swiss French (FR-CH)
- Polish (Programmers)
- Czech (CS)
- Slovak (SK)
- Hungarian (HU)
- Turkish (Q)
- Russian (RU)
- Ukrainian (UK)
- Hebrew (HE)
- Arabic (AR)
- Japanese (JP)
- Korean (KO)

> Tip: If your target system uses a different layout than your local PC, pick the layout that matches the **target**. The mapping is performed using that layout’s OS keyboard table.

---

## How it works (high level)

- Resolves each character (based on the chosen layout) with `VkKeyScanExW` → **virtual key** + required **modifiers**.
- Converts VK → hardware **scan code** via `MapVirtualKeyExW`.
- Sends **press/release** events with `SendInput` and `KEYEVENTF_SCANCODE`.
- If mapping fails (e.g., emoji), falls back to **Unicode injection**.

This is why web consoles and VMs that ignore paste/Unicode still receive keystrokes.

---

## Requirements

- Windows 10/11 (x64)
- Go 1.22+ (to build)
- CGO toolchain (MinGW-w64) for Fyne on Windows

---

## Build

```powershell
# in the project root
go mod tidy
go build -trimpath -ldflags="-H=windowsgui -s -w" -o goclip.exe .
```

> The `-H=windowsgui` flag hides the console window for a cleaner UX.

If you need MinGW-w64 for CGO on the GitHub runner, see the provided workflow below.

---

## GitHub Actions (preconfigured)

This repo can include a workflow to build and publish a Windows `.exe` and a zipped asset on push and tags:

```
.github/workflows/build-windows.yml
```

- Runs on `windows-latest`
- Installs **MinGW-w64** for CGO
- Builds `goclip-windows-amd64.exe`
- Uploads artifacts
- On tags (`v*`) also creates a **GitHub Release** and attaches the files

---

## Run

1. Launch **goclip**.
2. Pick **Keyboard Layout** (or keep “Auto (Use System)”).
3. Select a **Target Window** from the dropdown, or press **Clear** so no selection → it will use the **last active** window.
4. Type your text in the big box.
5. Click **Type**.  
   goclip briefly focuses the target window and injects keystrokes.

---

## Notes & limitations

- **Elevation:** Windows blocks sending input from a non-elevated process to an **elevated** target (UAC). If you need to type into admin apps, run goclip **as Administrator**.
- **Focus rules:** Windows sometimes restricts focus changes. We try to foreground the target just before typing, but if the target is stubborn, click it once to focus, then press **Type**.
- **CJK/IME:** For Japanese/Korean/Chinese and other IME-based input, ASCII works via scan codes. Composed characters may require IME state; Unicode fallback helps, but some web consoles ignore Unicode entirely.
- **Browser consoles:** Ensure the console iframe has focus (click into it once).

---

## Add / customize layouts

Layouts are loaded by **KLID** (keyboard layout ID) using `LoadKeyboardLayoutW`. To add more entries, extend the `loadHKLByName` switch with the appropriate KLID:

```go
func loadHKLByName(name string) windows.Handle {
  if name == "Auto (Use System)" || name == "" {
    h, _, _ := procGetKeyboardLayout.Call(0)
    return windows.Handle(h)
  }
  klid := ""
  switch name {
  case "Belgian (Period)":
    klid = "0000080C" // example
  // add more here...
  default:
    h, _, _ := procGetKeyboardLayout.Call(0)
    return windows.Handle(h)
  }
  ptr, _ := windows.UTF16PtrFromString(klid)
  h, _, _ := procLoadKeyboardLayoutW.Call(uintptr(unsafe.Pointer(ptr)), uintptr(0))
  return windows.Handle(h)
}
```

---

## License

MIT
