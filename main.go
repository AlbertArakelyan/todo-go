package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Task struct {
	Id          uint
	Title       string
	Description string
}

func main() {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Task Manager")
	w.Resize(fyne.NewSize(500, 600))
	w.CenterOnScreen()

	var tasks []Task // storage of tasks
	var createContent *fyne.Container
	var tasksContent *fyne.Container
	var tasksList *widget.List

	// Connect to DB (_ is for err variable, but I will not handle it)
	DB, _ := gorm.Open(sqlite.Open("todo.db"), &gorm.Config{})
	DB.AutoMigrate(&Task{}) // php artisan migrate :D automatically

	// Get all tasks from DB on start
	DB.Find(&tasks)

	noTasksLabel := widget.NewLabel("No Tasks")
	// noTasksLabel := canvas.NewText("No tasks", color.Black)

	if len(tasks) != 0 {
		noTasksLabel.Hide()
	}

	// Icons
	newTaskIcon, _ := fyne.LoadResourceFromPath("./icons/plus.png")
	editTaskIcon, _ := fyne.LoadResourceFromPath("./icons/edit.png")
	backIcon, _ := fyne.LoadResourceFromPath("./icons/back.png")
	saveIcon, _ := fyne.LoadResourceFromPath("./icons/save.png")
	deleteTaskIcon, _ := fyne.LoadResourceFromPath("./icons/delete.png")

	// Header

	tasksBar := container.NewHBox( // NewHBox() from left to right (flex-direction: row)
		// widget.NewLabel("Your Tasks"),
		canvas.NewText("Your Tasks", color.Black),
		layout.NewSpacer(), // like margin-left: auto
		widget.NewButtonWithIcon("", newTaskIcon, func() {
			w.SetContent(createContent)
		}),
	)

	// Main

	tasksList = widget.NewList(
		func() int {
			return len(tasks)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Default")
		},
		func(lii widget.ListItemID, co fyne.CanvasObject) {
			co.(*widget.Label).SetText(tasks[lii].Title)
		},
	)

	tasksList.OnSelected = func(id widget.ListItemID) {
		detailsBar := container.NewHBox(
			canvas.NewText(
				fmt.Sprintf("Details about \"%s\"", tasks[id].Title),
				color.Black,
			),
			layout.NewSpacer(),
			widget.NewButtonWithIcon("", backIcon, func() {
				w.SetContent(tasksContent)
				tasksList.Unselect(id)
			}),
		)

		taskTitle := widget.NewLabel(tasks[id].Title)
		taskTitle.TextStyle = fyne.TextStyle{Bold: true}

		taskDescription := widget.NewLabel(tasks[id].Description)
		taskDescription.TextStyle = fyne.TextStyle{Italic: true}
		taskDescription.Wrapping = fyne.TextWrapBreak

		buttonsBox := container.NewHBox(
			// DELETE
			widget.NewButtonWithIcon("", deleteTaskIcon, func() {
				// dialog.ShowConfirm(
				// 	"Delete Task",
				// 	fmt.Sprintf("Are you sure you want to delete \"%s\"?", tasks[id].Title),
				// 	func(b bool) {
				// 		if b {
				// 			DB.Delete(&Task{}, "Id = ?", tasks[id].Id)
				// 			DB.Find(&tasks)

				// 			if len(tasks) == 0 {
				// 				noTasksLabel.Show()
				// 			} else {
				// 				noTasksLabel.Hide()
				// 			}
				// 		}
				// 		w.SetContent(tasksContent)
				// 	},
				// 	w,
				// ),

				DB.Delete(&Task{}, "Id = ?", tasks[id].Id)
				DB.Find(&tasks)

				if len(tasks) == 0 {
					noTasksLabel.Show()
				} else {
					noTasksLabel.Hide()
				}

				w.SetContent(tasksContent)
			}),

			// EDIT
			widget.NewButtonWithIcon("", editTaskIcon, func() {
				editBar := container.NewHBox(
					canvas.NewText(
						fmt.Sprintf("Editing \"%s\"", tasks[id].Title),
						color.Black,
					),
					layout.NewSpacer(),
					widget.NewButtonWithIcon("", backIcon, func() {
						w.SetContent(tasksContent)
						tasksList.Unselect(id)
					}),
				)

				editTitleEntry := widget.NewEntry()
				editTitleEntry.SetText(tasks[id].Title)

				editDescriptionEntry := widget.NewMultiLineEntry()
				editDescriptionEntry.SetText(tasks[id].Description)

				editButton := widget.NewButtonWithIcon("Save Task", saveIcon, func() {
					DB.Find(&Task{}, "Id = ?", tasks[id].Id).Updates(
						Task{
							Title:       editTitleEntry.Text,
							Description: editDescriptionEntry.Text,
						},
					)

					DB.Find(&tasks)

					w.SetContent(tasksContent)
					tasksList.Unselect(id)
				})

				editContent := container.NewVBox(
					editBar,
					canvas.NewLine(color.Black),
					editTitleEntry,
					editDescriptionEntry,
					editButton,
				)

				w.SetContent(editContent)
			}),
		)

		detailsVBox := container.NewVBox(
			detailsBar,
			canvas.NewLine(color.Black),
			taskTitle,
			taskDescription,
			buttonsBox,
		)

		w.SetContent(detailsVBox)
	}

	tasksScroll := container.NewScroll(tasksList)
	tasksScroll.SetMinSize(fyne.NewSize(500, 500))

	tasksContent = container.NewVBox( // NewVBox() from up to down (flex-direction: column)
		tasksBar,
		canvas.NewLine(color.Black), // <hr />
		noTasksLabel,
		tasksScroll,
	)

	titleEntry := widget.NewEntry()
	titleEntry.SetPlaceHolder("Task Title...")

	descriptionEntry := widget.NewMultiLineEntry()
	descriptionEntry.SetPlaceHolder("Test Description...")

	saveTaskButton := widget.NewButtonWithIcon("Save Task", saveIcon, func() {
		task := Task{
			Title:       titleEntry.Text,
			Description: descriptionEntry.Text,
		}

		DB.Create(&task)
		DB.Find(&tasks)

		titleEntry.Text = ""
		titleEntry.Refresh()

		descriptionEntry.Text = ""
		descriptionEntry.Refresh()

		w.SetContent(tasksContent)
		tasksList.UnselectAll()

		if len(tasks) == 0 {
			noTasksLabel.Show()
		} else {
			noTasksLabel.Hide()
		}
	})

	createBar := container.NewHBox(
		canvas.NewText("Create New Task", color.Black),
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", backIcon, func() {
			titleEntry.Text = ""
			titleEntry.Refresh()

			descriptionEntry.Text = ""
			descriptionEntry.Refresh()

			w.SetContent(tasksContent)
			tasksList.UnselectAll()
		}),
	)

	createContent = container.NewVBox(
		createBar,
		canvas.NewLine(color.Black),
		container.NewVBox(
			titleEntry,
			descriptionEntry,
			saveTaskButton,
		),
	)

	w.SetContent(tasksContent) // something like appendChild()
	w.Show()
	a.Run()
}
