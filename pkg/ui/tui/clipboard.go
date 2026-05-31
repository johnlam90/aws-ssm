package tui

import (
	"fmt"
	"os/exec"
	"runtime"
)

// writeToClipboard sends s to the system clipboard via the platform's
// command-line clipboard tool. Returns an error if no working tool is
// available; the caller is expected to surface that error in the UI.
//
// On Linux/WSL this prefers xclip → xsel → wl-copy. On macOS it uses
// pbcopy. On Windows it uses clip.exe. Each tool reads from stdin.
func writeToClipboard(s string) error {
	cands := clipboardCandidates()
	if len(cands) == 0 {
		return fmt.Errorf("no clipboard tool available on %s", runtime.GOOS)
	}

	var firstErr error
	for _, c := range cands {
		if _, err := exec.LookPath(c.bin); err != nil {
			continue
		}
		cmd := exec.Command(c.bin, c.args...)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := cmd.Start(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if _, err := stdin.Write([]byte(s)); err != nil {
			_ = stdin.Close()
			_ = cmd.Wait()
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		_ = stdin.Close()
		if err := cmd.Wait(); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		return nil
	}

	if firstErr != nil {
		return firstErr
	}
	return fmt.Errorf("no clipboard tool succeeded")
}

type clipboardTool struct {
	bin  string
	args []string
}

func clipboardCandidates() []clipboardTool {
	switch runtime.GOOS {
	case "darwin":
		return []clipboardTool{{bin: "pbcopy"}}
	case "windows":
		return []clipboardTool{{bin: "clip"}}
	default:
		return []clipboardTool{
			{bin: "xclip", args: []string{"-selection", "clipboard"}},
			{bin: "xsel", args: []string{"--clipboard", "--input"}},
			{bin: "wl-copy"},
			{bin: "clip.exe"}, // WSL bridge
		}
	}
}
