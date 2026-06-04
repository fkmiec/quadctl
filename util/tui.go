package util

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// listModel holds the state for our Bubble Tea program
type listModel struct {
	choices  []string // The items passed in
	cursor   int      // The currently highlighted item index
	selected string   // The final selected item
}

// Init is called when the program starts. We don't need any initial I/O, so we return nil.
func (m listModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming events like key presses.
func (m listModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch msg.String() {
		// Exit without selecting anything
		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		// Move cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// Move cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// Confirm selection
		case "enter":
			m.selected = m.choices[m.cursor]
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI to the terminal.
func (m listModel) View() string {
	s := "Select an item (use arrow keys/j/k, press Enter to select):\n\n"

	for i, choice := range m.choices {
		// Render the cursor pointing at the current item
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		s += fmt.Sprintf("%s%s\n", cursor, choice)
	}

	s += "\nPress 'q' or 'esc' to quit.\n"
	return s
}

// SelectFromList takes a slice of strings, displays a UI, and returns the selected string.
func SelectFromList(items []string) (string, error) {
	if len(items) == 0 {
		return "", fmt.Errorf("the provided list is empty")
	}

	// Initialize our Bubble Tea program with our initial model
	p := tea.NewProgram(listModel{
		choices: items,
	})

	// Run the program and wait for it to finish
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running list UI: %w", err)
	}

	// Cast the returned tea.Model back to our specific listModel
	if m, ok := finalModel.(listModel); ok {
		if m.selected != "" {
			return m.selected, nil
		}
	}

	return "", fmt.Errorf("no item was selected")
}

/*
func main() {
	// Example usage
	options := []string{
		"Initialize database",
		"Run migrations",
		"Start server",
		"Run test suite",
	}

	selected, err := SelectFromList(options)
	if err != nil {
		fmt.Printf("Operation cancelled or failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nYou selected: %s\n", selected)
}
*/
