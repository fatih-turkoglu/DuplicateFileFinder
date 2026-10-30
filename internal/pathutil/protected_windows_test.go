//go:build windows

package pathutil

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWindowsProtectedPaths(t *testing.T) {
	for _, path := range []string{
		`C:\Windows\System32\kernel32.dll`, `c:\windows\winsxs\a.dll`, `C:\Program Files\App\app.exe`,
		`C:\Program Files (x86)\App\lib.dll`, `C:\ProgramData\App\data.bin`, `C:\$Recycle.Bin\S-1\a.txt`,
		`C:\System Volume Information\x`, `C:\Users\me\AppData\Roaming\App\settings.json`,
		`C:\Users\me\AppData\Local\Programs\App\app.exe`, `C:\Users\me\AppData`,
		`D:\Windows\System32\a.dll`, `D:\Program Files\App\a.dll`,
	} {
		if !IsProtected(path) {
			t.Errorf("not protected: %s", path)
		}
	}
	for _, path := range []string{
		`C:\`, `D:\`, `C:\Users\me\Documents\a.txt`, `C:\Users\me\Pictures\a.jpg`, `D:\Photos\a.jpg`,
		`C:\Users\me\Documents\Windows\a.txt`, `C:\Users\me\AppData\Local\Temp\a.txt`,
		`C:\Users\me\Documents\Program Files notes.txt`,
		filepath.Join(t.TempDir(), "test.txt"),
	} {
		if IsProtected(path) {
			t.Errorf("personal path blocked: %s", path)
		}
	}
}

func TestWindowsProtectedJunction(t *testing.T) {
	sys := os.Getenv("SystemRoot")
	if sys == "" {
		t.Skip("SystemRoot not set")
	}
	link := filepath.Join(t.TempDir(), "sys")
	// Junctions need no special privilege, unlike symlinks.
	if out, err := exec.Command("cmd", "/c", "mklink", "/J", link, sys).CombinedOutput(); err != nil {
		t.Skipf("could not create junction: %v %s", err, out)
	}
	if !IsProtected(filepath.Join(link, "win.ini")) {
		t.Fatal("junction bypassed system folder protection")
	}
}
