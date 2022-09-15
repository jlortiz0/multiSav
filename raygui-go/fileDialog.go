package rayguigo

import (
	"os"
	"path"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	fdstate_active = 1 << iota
	fdstate_pathedit
	fdstate_fileslistedit
	fdstate_filenameedit
	fdstate_selected
	fdstate_save
)

type FileDialog struct {
	position, size  rl.Vector2
	state           int
	path, pathBak   string
	fListInd        int
	fListActive     int
	fName, fNameBak string
	fTypeActive     int
	itemFocused     int
	files           []string
}

func InitGuiFileDialog(width, height int, initpath string, active bool, save bool) *FileDialog {
	fd := new(FileDialog)
	if height == -1 {
		height = 440
	}
	if width == -1 {
		width = 310
	}
	fd.size = rl.Vector2{X: float32(width), Y: float32(height)}
	fd.position = rl.Vector2{X: float32(rl.GetScreenWidth()-width) / 2, Y: float32(rl.GetScreenHeight()-height) / 2}
	if active {
		fd.state = fdstate_active
	}
	if save {
		fd.state |= fdstate_save
	}
	fd.fListActive = -1
	if initpath != "" {
		stat, err := os.Stat(initpath)
		if err == nil {
			if stat.IsDir() {
				fd.path = initpath
			} else {
				fd.path = path.Dir(initpath)
				fd.fName = path.Base(initpath)
				fd.fNameBak = fd.fName
			}
		} else {
			fd.path, _ = os.Getwd()
		}
	} else {
		fd.path, _ = os.Getwd()
	}
	fd.pathBak = fd.path
	return fd
}

func (fd *FileDialog) GuiFileDialog() {
	if fd.state&fdstate_active == 0 {
		return
	}
	if fd.files == nil {
		data, _ := os.ReadDir(fd.path)
		fd.files = make([]string, len(data))
		for i, x := range data {
			fd.files[i] = x.Name()
			if x.Name() == fd.fName {
				fd.fListActive = len(fd.files) - 1
				fd.fListInd = len(fd.files) - 1
			}
		}
	}
	rl.DrawRectangle(0, 0, int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight()), rl.Fade(rl.GetColor(uint(GuiGetStyle(DEFAULT, BACKGROUND_COLOR))), 0.85))
	if GuiWindowBox(rl.Rectangle{X: fd.position.X, Y: fd.position.Y, Width: fd.size.X, Height: fd.size.Y}, "#198# Select File Dialog") {
		fd.state &= ^fdstate_active
	}

	if GuiButton(rl.Rectangle{X: fd.position.X + fd.size.X - 50, Y: fd.position.Y + 35, Width: 40, Height: 25}, "< ..") {
		fd.path = path.Join(fd.path, "..")
		fd.reloadDirectoryFiles()
		fd.fListActive = -1
		fd.fName = ""
		fd.fNameBak = ""
	}

	var b bool
	b, fd.path = GuiTextBox(rl.Rectangle{X: fd.position.X + 10, Y: fd.position.Y + 35, Width: fd.size.Y - 65, Height: 25}, fd.path, 256, fd.state&fdstate_pathedit != 0)
	if b {
		if fd.state&fdstate_pathedit != 0 {
			stat, err := os.Stat(fd.path)
			if err == nil && stat.IsDir() {
				fd.reloadDirectoryFiles()
				fd.pathBak = fd.path
			} else {
				fd.path = fd.pathBak
			}
		}
		fd.state ^= fdstate_pathedit
	}

	prevTextAlign := GuiGetStyle(LISTVIEW, TEXT_ALIGNMENT)
	prevElementHeight := GuiGetStyle(LISTVIEW, LIST_ITEMS_HEIGHT)
	GuiSetStyle(LISTVIEW, TEXT_ALIGNMENT, TEXT_ALIGN_LEFT)
	GuiSetStyle(LISTVIEW, LIST_ITEMS_HEIGHT, 24)
	prevFListActive := fd.fListActive
	fd.fListActive = GuiListViewEx(rl.Rectangle{X: fd.position.X + 10, Y: fd.position.Y + 70, Width: fd.size.X - 20, Height: fd.size.Y - 135}, nil, len(fd.files), &fd.itemFocused, &fd.fListInd, fd.fListActive)
	GuiSetStyle(LISTVIEW, TEXT_ALIGNMENT, prevTextAlign)
	GuiSetStyle(LISTVIEW, LIST_ITEMS_HEIGHT, prevElementHeight)

	if fd.fListActive >= 0 && fd.fListActive != prevFListActive {
		fd.fName = fd.files[fd.fListActive]
		stat, err := os.Stat(path.Join(fd.path, fd.fName))
		if err != nil && stat.IsDir() {
			fd.path = path.Join(fd.path, fd.fName)
			fd.pathBak = fd.path
			fd.reloadDirectoryFiles()
			// Is this a mistake?
			fd.pathBak = fd.path
			fd.fListActive = -1
			fd.fName = ""
			fd.fNameBak = ""
		}
	}

	GuiLabel(rl.Rectangle{X: fd.position.X + 10, Y: fd.position.Y + fd.size.Y - 60, Width: 68, Height: 25}, "File name:")
	b, fd.fName = GuiTextBox(rl.Rectangle{X: fd.position.X + 75, Y: fd.position.Y + fd.size.Y - 60, Width: fd.size.X - 200, Height: 25}, fd.fName, 128, fd.state&fdstate_filenameedit != 0)
	if b {
		if fd.state&fdstate_filenameedit != 0 && fd.fName != "" {
			_, err := os.Stat(path.Join(fd.path, fd.fName))
			if err != nil {
				for i, x := range fd.files {
					if x == fd.fName {
						fd.fListActive = i
						fd.fNameBak = fd.fName
						break
					}
				}
			}
		} else if fd.state&fdstate_save == 0 {
			fd.fName = fd.fNameBak
		}
		fd.state ^= fdstate_filenameedit
	}

	fd.fTypeActive = GuiComboBox(rl.Rectangle{X: fd.position.X + 75, Y: fd.position.Y + fd.size.Y - 30, Width: fd.size.X - 200, Height: 25}, "All files", fd.fTypeActive)
	GuiLabel(rl.Rectangle{X: fd.position.X + 10, Y: fd.position.Y + fd.size.Y - 30, Width: 68, Height: 25}, "File filter:")
	if GuiButton(rl.Rectangle{X: fd.position.X + fd.size.X - 120, Y: fd.position.Y + fd.size.Y - 60, Width: 110, Height: 25}, "Select") {
		fd.state ^= fdstate_active | fdstate_selected
	}
	if GuiButton(rl.Rectangle{X: fd.position.X + fd.size.X - 120, Y: fd.position.Y + fd.size.Y - 30, Width: 110, Height: 25}, "Cancel") {
		fd.state ^= fdstate_active
	}
}

func (fd *FileDialog) reloadDirectoryFiles() {
	data, _ := os.ReadDir(fd.path)
	fd.files = make([]string, len(data))
	for i, x := range data {
		fd.files[i] = x.Name()
	}
	fd.itemFocused = 0
}

func (fd *FileDialog) IsActive() bool {
	return fd.state&fdstate_active != 0
}

func (fd *FileDialog) GetResult() string {
	if fd.state&fdstate_active != 0 || fd.state&fdstate_selected == 0 {
		return ""
	}
	return path.Join(fd.path, fd.fName)
}
