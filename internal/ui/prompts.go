package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	noStyle      = lipgloss.NewStyle()

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

// SelectModel manages the state for a select prompt
type SelectModel struct {
	choices  []string
	cursor  int
	selected map[int]struct{}
	done    bool
	cancelled bool
}

// NewSelectModel creates a new select prompt model
func NewSelectModel(choices []string) SelectModel {
	return SelectModel{
		choices:  choices,
		selected: make(map[int]struct{}),
	}
}

// Init initializes the model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.cancelled = true
			m.done = true
			return m, tea.Quit
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			m.done = true
			return m, tea.Quit
		case "down", "j":
			m.cursor++
			if m.cursor >= len(m.choices) {
				m.cursor = 0
			}
		case "up", "k":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.choices) - 1
			}
		}
	}

	return m, nil
}

// View renders the select prompt
func (m SelectModel) View() string {
	if m.done {
		return ""
	}

	var b strings.Builder
	b.WriteString("Select an option (↑/↓ to move, enter to select):\n\n")

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		style := blurredStyle
		if i == m.cursor {
			style = focusedStyle
		}

		b.WriteString(fmt.Sprintf("%s %s\n", cursor, style.Render(choice)))
	}

	b.WriteString("\nPress q to quit\n")
	return b.String()
}

// Selected returns the selected choice
func (m SelectModel) Selected() string {
	if len(m.selected) == 0 {
		return ""
	}
	for i := range m.selected {
		return m.choices[i]
	}
	return ""
}

// Cancelled reports whether the user cancelled the prompt (e.g., via Ctrl+C or q)
func (m SelectModel) Cancelled() bool {
	return m.cancelled
}

// InputModel manages the state for a text input prompt
type InputModel struct {
	textInput textinput.Model
	done     bool
	cancelled bool
}

// NewInputModel creates a new input prompt model
func NewInputModel(prompt, defaultValue string, isPassword bool) InputModel {
	ti := textinput.New()
	ti.Prompt = prompt + ": "
	ti.SetValue(defaultValue)
	ti.Focus()
	ti.CharLimit = 256

	if isPassword {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	}

	return InputModel{
		textInput: ti,
		done:     false,
	}
}

// Init initializes the model
func (m InputModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles user input
func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.done = true
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			m.done = true
			return m, tea.Quit
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View renders the input prompt
func (m InputModel) View() string {
	return m.textInput.View()
}

// Value returns the input value
func (m InputModel) Value() string {
	return m.textInput.Value()
}

// Cancelled reports whether the user cancelled the input (Ctrl+C or Esc)
func (m InputModel) Cancelled() bool {
	return m.cancelled
}

// ConfirmModel manages the state for a confirmation prompt
type ConfirmModel struct {
	prompt string
	result bool
	done   bool
	cancelled bool
}

// NewConfirmModel creates a new confirmation prompt model
func NewConfirmModel(prompt string, defaultValue bool) ConfirmModel {
	return ConfirmModel{
		prompt: prompt,
		result: defaultValue,
		done:   false,
	}
}

// Init initializes the model
func (m ConfirmModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m ConfirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.cancelled = true
			m.done = true
			return m, tea.Quit
		case "y", "Y":
			m.result = true
			m.done = true
			return m, tea.Quit
		case "n", "N":
			m.result = false
			m.done = true
			return m, tea.Quit
		case "enter":
			m.done = true
			return m, tea.Quit
		case "left", "h":
			m.result = !m.result
		case "right", "l":
			m.result = !m.result
		}
	}

	return m, nil
}

// View renders the confirmation prompt
func (m ConfirmModel) View() string {
	yesNo := "(y/N)"
	if m.result {
		yesNo = "(Y/n)"
	}

	prompt := fmt.Sprintf("%s %s ", m.prompt, yesNo)

	if m.done {
		return ""
	}

	return prompt
}

// Result returns the confirmation result
func (m ConfirmModel) Result() bool {
	return m.result
}

// Cancelled reports whether the user cancelled the confirmation prompt
func (m ConfirmModel) Cancelled() bool {
	return m.cancelled
}
