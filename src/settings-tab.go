package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func makeSettingsTab(dept *widget.Select) *fyne.Container {
	// Settings
	pms := widget.NewMultiLineEntry()
	pms.SetText(myApp.Preferences().StringWithFallback("ProductManagers", "[]"))
	pms.SetMinRowsVisible(7)
	savepath := widget.NewEntry()
	savepath.SetText(myApp.Preferences().StringWithFallback("SavePath", ""))
	return container.NewVBox(
		widget.NewForm(
			widget.NewFormItem("Domains", dept),
			widget.NewFormItem("Product Managers", pms),
			widget.NewFormItem("Save path", savepath),
		),
		widget.NewButton("Save", func() {
			myApp.Preferences().SetString("Department", dept.Selected)
			myApp.Preferences().SetString("ProductManagers", PrettyJSONString(pms.Text))
			myApp.Preferences().SetString("SavePath", savepath.Text)
		}))
}
