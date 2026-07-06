package todo

import (
	"todo_cli/internal/styles"

	"charm.land/lipgloss/v2"
)

func (m State) View() string {
	s := ""

	for i, todo := range m.Todos {

		itemStyle := styles.Task
		cursor := " "

		if m.Cursor == i {
			itemStyle = styles.TaskHover
			cursor = styles.Cursor.Render("▶")
		}

		checked := "[ ]"
		if m.Todos[i].Done == true {
			checked = styles.Completed.UnsetStrikethrough().Render("[✓]")
			itemStyle = styles.Completed
		}

		// s += fmt.Sprintf("%s %s %s\n", cursor, checked, itemStyle.Render(todo.Title))
		line := lipgloss.JoinHorizontal(
			lipgloss.Left,
			" ",
			cursor,
			" ",
			checked,
			" ",
			lipgloss.PlaceHorizontal(50, lipgloss.Left, itemStyle.Render(todo.Title)),
			// "created_at: ",
			// todo.CreatedAt.Format("15:04:05"),
			// " updated_at: ",
			// todo.UpdatedAt.Format("15:04:05"),
		)

		s += line + "\n"
	}

	return s + "\n" +m.help.MenuRender()
}
