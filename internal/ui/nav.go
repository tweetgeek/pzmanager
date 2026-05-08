package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type NavSidebar struct {
	onSelect func(section string)
	items    *widget.List
	cont     *fyne.Container
	sections []navItem
}

type navItem struct {
	id    string
	label string
}

func NewNavSidebar(onSelect func(section string)) *NavSidebar {
	sections := []navItem{
		{"saves", "📁 Saves"},
		{"settings", "⚙ Settings"},
		{"about", "ℹ About"},
	}
	n := &NavSidebar{onSelect: onSelect, sections: sections}

	n.items = widget.NewList(
		func() int { return len(n.sections) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(n.sections[i].label)
		},
	)
	n.items.OnSelected = func(i widget.ListItemID) {
		if onSelect != nil {
			onSelect(n.sections[i].id)
		}
	}

	icon := canvas.NewImageFromResource(AppIcon)
	icon.FillMode = canvas.ImageFillContain
	icon.SetMinSize(fyne.NewSize(64, 64))

	launchBtn := widget.NewButton("▶ Launch Game", func() {
		if onSelect != nil {
			onSelect("launch")
		}
	})

	top := container.NewVBox(
		container.NewCenter(icon),
		launchBtn,
	)
	n.cont = container.NewBorder(top, nil, nil, nil, n.items)
	return n
}

func (n *NavSidebar) Container() fyne.CanvasObject { return n.cont }

// Select wywołuje callback bez interakcji UI (do testów).
func (n *NavSidebar) Select(section string) {
	if n.onSelect != nil {
		n.onSelect(section)
	}
}
