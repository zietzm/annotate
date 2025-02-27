package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type model struct {
	list          list.Model
	viewport      viewport.Model
	textarea      textarea.Model
	records       []item
	currentItem   int
	mode          string
	width         int
	height        int
	viewportStyle lipgloss.Style
	textareaStyle lipgloss.Style
	headerStyle   lipgloss.Style
}

type item struct {
	number     int
	title      string
	text       string
	annotation string
}

func (i item) Title() string       { return fmt.Sprintf("%d. %s", i.number, i.title) }
func (i item) Description() string { return i.text }
func (i item) FilterValue() string { return i.title }

func initialModel(records []item) model {
	items := make([]list.Item, len(records))
	for i, record := range records {
		items[i] = record
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Items to Annotate"

	ta := textarea.New()
	ta.Placeholder = "Type your annotation here..."
	ta.CharLimit = 0 // No limit
	ta.Focus()

	vp := viewport.New(0, 0)
	vp.MouseWheelEnabled = false

	return model{
		list:     l,
		viewport: vp,
		textarea: ta,
		records:  records,
		mode:     "list",
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case m.mode == "list" && (key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c")))):
			return m, tea.Quit
		case m.mode == "list" && msg.Type == tea.KeyEnter:
			m.mode = "edit"
			m.currentItem = m.list.Index()
			m.viewport.SetContent(m.records[m.currentItem].text)
			m.textarea.SetValue(m.records[m.currentItem].annotation)
			m.textarea.Focus()
			return m, textarea.Blink
		case m.mode == "edit" && (key.Matches(msg, key.NewBinding(key.WithKeys("esc", "ctrl+c")))):
			m.mode = "list"
			m.saveAnnotation()
			return m, nil
		case m.mode == "edit" && (key.Matches(msg, key.NewBinding(key.WithKeys("tab")))):
			m.saveAnnotation()
			m.nextItem()
			return m, textarea.Blink
		case m.mode == "edit" && (key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab")))):
			m.saveAnnotation()
			m.previousItem()
			return m, textarea.Blink
		case m.mode == "edit" && msg.Type == tea.KeyPgUp:
			m.viewport.HalfViewUp()
			return m, nil
		case m.mode == "edit" && msg.Type == tea.KeyPgDown:
			m.viewport.HalfViewDown()
			return m, nil
		case m.mode == "edit":
			// Only give the message to the textarea, not the viewport
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height)

		contentHeight := msg.Height - 4
		contentWidth := msg.Width/2 - 4
		m.viewport.Width = contentWidth
		m.viewport.Height = contentHeight
		m.textarea.SetWidth(contentWidth)
		m.textarea.SetHeight(contentHeight)

		m.viewportStyle = lipgloss.NewStyle().
			Width(m.width/2 - 4).
			Height(m.height - 4).
			Border(lipgloss.RoundedBorder())
		m.textareaStyle = lipgloss.NewStyle().
			Width(m.width/2 - 4).
			Height(m.height - 4).
			Border(lipgloss.RoundedBorder())

		m.headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#777777"))
	case tea.MouseMsg:
		return m, nil
	}

	if m.mode == "list" {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.mode == "list" {
		return m.list.View()
	}

	headerContent := fmt.Sprintf("Item %d/%d", m.currentItem+1, len(m.records))

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.headerStyle.Render(headerContent),
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			m.viewportStyle.Render(m.viewport.View()),
			m.textareaStyle.Render(m.textarea.View()),
		),
	)
}

func (m *model) saveAnnotation() {
	m.records[m.currentItem].annotation = m.textarea.Value()
	m.list.SetItem(m.currentItem, m.records[m.currentItem])
}

func (m *model) nextItem() {
	if m.currentItem < len(m.records)-1 {
		m.currentItem++
		m.list.Select(m.currentItem)
		m.viewport.SetContent(m.records[m.currentItem].text)
		m.textarea.SetValue(m.records[m.currentItem].annotation)
		m.viewport.GotoTop()
	}
}

func (m *model) previousItem() {
	if m.currentItem > 0 {
		m.currentItem--
		m.viewport.SetContent(m.records[m.currentItem].text)
		m.textarea.SetValue(m.records[m.currentItem].annotation)
		m.viewport.GotoTop()
	}
}

func main() {
	var inputFile string
	var outputFile string
	var textColumn string
	var annotationColumn string

	var rootCmd = &cobra.Command{
		Use:   "annotate",
		Short: "A brief description of your application",
		Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application.`,
		Run: func(cmd *cobra.Command, args []string) {
			if inputFile == "" || outputFile == "" || textColumn == "" {
				fmt.Println("Please provide input file, output file, and column to annotate")
				pflag.PrintDefaults()
				os.Exit(1)
			}

			records, err := readCsv(inputFile, textColumn, annotationColumn)
			if err != nil {
				fmt.Println("Error reading CSV:", err)
				os.Exit(1)
			}

			m := initialModel(records)
			p := tea.NewProgram(m, tea.WithAltScreen())

			if _, err := p.Run(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}

			if annotationColumn == "" {
				annotationColumn = "annotation"
			}

			err = writeCsv(outputFile, m.records, textColumn, annotationColumn)
			if err != nil {
				fmt.Println("Error writing CSV:", err)
				os.Exit(1)
			}

			fmt.Println("Annotations saved to", outputFile)
		},
	}

	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input CSV file (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output CSV file (required)")
	rootCmd.Flags().StringVarP(&textColumn, "text", "t", "", "Column to annotate (required)")
	rootCmd.Flags().StringVarP(
		&annotationColumn,
		"annotation",
		"a",
		"",
		"Column containing annotations",
	)
	rootCmd.MarkFlagRequired("input")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("text")
	rootCmd.Execute()
}
