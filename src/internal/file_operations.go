package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rkoesters/xdg/trash"
)

// Move file or directory
func moveElement(src, dst string) error {
	err := os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("failed to move file: %v", err)
	}
	return nil
}

// Move file to trash can and can auto switch macos trash can or linux trash can
func trashMacOrLinux(src string) error {
	if runtime.GOOS == "darwin" {
		err := moveElement(src, HomeDir+"/.Trash/"+filepath.Base(src))
		if err != nil {
			outPutLog("Delete single item function move file to trash can error", err)
		}
	} else {
		err := trash.Trash(src)
		if err != nil {
			outPutLog("Paste item function move file to trash can error", err)
		}
	}
	return nil
}

// Paste all item in directory
func pasteDir(src, dst string, id string, m model) (model, error) {
	// Check if destination directory already exists
	dst, err := renameIfDuplicate(dst)
	if err != nil {
		return m, err
	}

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		newPath := filepath.Join(dst, relPath)
		
		if info.IsDir() {
			newPath, err = renameIfDuplicate(newPath)
			if err != nil {
				return err
			}
			err = os.MkdirAll(newPath, info.Mode())
			if err != nil {
				return err
			}
		} else {
			p := m.processBarModel.process[id]
			if m.copyItems.cut {
				p.name = "󰆐 " + filepath.Base(path)
			} else {
				p.name = "󰆏 " + filepath.Base(path)
			}

			if len(channel) < 5 {
				channel <- channelMessage{
					messageId:       id,
					processNewState: p,
				}
			}

			err := pasteFile(path, newPath)
			if err != nil {
				p.state = failure
				channel <- channelMessage{
					messageId:       id,
					processNewState: p,
				}
				return err
			}
			p.done++
			if len(channel) < 5 {
				channel <- channelMessage{
					messageId:       id,
					processNewState: p,
				}
			}
			m.processBarModel.process[id] = p
		}

		return nil
	})

	if err != nil {
		return m, err
	}

	return m, nil
}