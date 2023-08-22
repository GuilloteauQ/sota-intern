package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2).MarginRight(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)

	titleStylePager = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStylePager = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStylePager.Copy().BorderStyle(b)
	}()

	modelStyle = lipgloss.NewStyle().
			Width(80).
			Height(80).
			Align(lipgloss.Center, lipgloss.Center).
			BorderStyle(lipgloss.HiddenBorder())

	focusedModelStyle = lipgloss.NewStyle().
				Width(80).
				Height(80).
				Align(lipgloss.Center, lipgloss.Center).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("69"))
)

func max(x int, y int) int {
	if x > y {
		return x
	}
	return y
}

func (m model) headerView() string {
	title := titleStylePager.Render("Abstact")
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(title)))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line)
}

func (m model) footerView() string {
	info := infoStylePager.Render(fmt.Sprintf("%3.f%%", m.viewport.ScrollPercent()*100))
	line := strings.Repeat("─", max(0, m.viewport.Width-lipgloss.Width(info)))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info)
}

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type statusMsg int

type HalResponse struct {
	Response Response `json:"response"`
}

type Response struct {
	NumFound      int        `json:"numFound"`
	Start         int        `json:"start"`
	MaxScore      float32    `json:"maxScore"`
	NumFoundExact bool       `json:"numFoundExact"`
	Documents     []Document `json:"docs"`
}

type Document struct {
	Title    []string `json:"title_s"`
	Abstract []string `json:"abstract_s"`
	HalId    string   `json:"halId_s"`
	Domains  []string `json:"domain_s"`
	SubDate  string   `json:"submittedDate_tdate"`
}

func send_get_req(keywords []string, response *HalResponse) tea.Msg {
	domain := "1.info.info-dc"
	fields := make([]string, 0)
	for _, kw := range keywords {
		fields = append(fields, fmt.Sprintf("\"%s\"~", kw))
	}
	title_request := strings.Join(fields, "||")
	url := fmt.Sprintf("https://api.archives-ouvertes.fr/search/?q=abstract_t:(%s)&fq=title_t:(%s)&fq=openAccess_bool:true&wt=json&fq=domain_s:%s&fl=title_s,submittedDate_tdate,abstract_s,halId_s,domain_s&rows=100000", title_request, title_request, domain)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("get")
		panic(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("read")
		panic(err)
	}

	err = ioutil.WriteFile("output.txt", data, 0644)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, response)
	if err != nil {
		fmt.Println("json")
		panic(err)
	}
	return statusMsg(0)
}

type model struct {
	keyword   string
	textInput textinput.Model
	queryDone bool
	response  *HalResponse
	list      list.Model
	choice    string
	quitting  bool
	// showAbstract bool
	viewport       viewport.Model
	content        string
	ready          bool
	viewOnAbstract bool
}

func initialModel() model {
    ti := textinput.New()
	ti.Placeholder = "openmp"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	var response HalResponse
	return model{textInput: ti, response: &response}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case statusMsg:
		switch int(msg) {
		case 0:
			m.queryDone = true

			titles := make([]list.Item, len(m.response.Response.Documents))
			for i, doc := range m.response.Response.Documents {
                width := m.list.Width() - 15
                wrapped := lipgloss.NewStyle().Width(width).Render(doc.Title[0])
				titles[i] = item(wrapped)
				// titles[i] = item(doc.Title[0])
			}
            m.list.SetItems(titles)

			return m, nil
		case 1:
			panic("oops")
		}
	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		if !m.ready {
			m.ready = true

			m.viewport = viewport.New(msg.Width / 2, msg.Height-verticalMarginHeight)
			m.viewport.HighPerformanceRendering = false //useHighPerformanceRenderer
			m.viewport.SetContent(m.content)
			m.viewport.YPosition = headerHeight

			m.list = list.New(nil, itemDelegate{}, msg.Width / 2, msg.Height-verticalMarginHeight)// msg.Height)
			m.list.Title = "What do you want for dinner?"
			m.list.SetShowStatusBar(false)
			m.list.SetFilteringEnabled(false)
			m.list.Styles.Title = titleStyle
			m.list.Styles.PaginationStyle = paginationStyle
			m.list.Styles.HelpStyle = helpStyle
		} else {
			m.viewport.Width = msg.Width / 2
			m.viewport.Height = msg.Height - verticalMarginHeight
            m.list.SetWidth(msg.Width / 2)
            m.list.SetHeight(msg.Height - verticalMarginHeight)

		}
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyTab:
			m.viewOnAbstract = !m.viewOnAbstract
			return m, nil
		case tea.KeyEnter:
			if m.queryDone {
				var halUrl string
				for _, doc := range m.response.Response.Documents {
                    width := m.list.Width() - 15
                    wrapped := lipgloss.NewStyle().Width(width).Render(doc.Title[0])
					if wrapped == m.choice {
						halUrl = fmt.Sprintf("https://hal.science/%s/document", doc.HalId)
						break
					}
				}
				cmdBrowser := exec.Command("xdg-open", halUrl)
				cmdBrowser.Run()
				return m, nil
			} else {
				m.keyword = m.textInput.Value()
				keywords := []string{m.keyword}
				return m, (func() tea.Msg { return send_get_req(keywords, m.response) })
			}
		}
	}
	if m.queryDone {
		if m.viewOnAbstract {
			m.viewport, cmd = m.viewport.Update(msg)
		} else {
			m.list, cmd = m.list.Update(msg)
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				for _, doc := range m.response.Response.Documents {
                    width := m.list.Width() - 15
                    wrapped := lipgloss.NewStyle().Width(width).Render(doc.Title[0])
					if wrapped == m.choice {
						width := m.viewport.Width - 10
						wrapped := lipgloss.NewStyle().Width(width).Render(doc.Abstract[0])
						m.viewport.SetContent(wrapped)
					}
				}
			}
		}
	} else {
		m.textInput, cmd = m.textInput.Update(msg)
	}
	return m, cmd
}

func (m model) View() string {
	if m.queryDone {
		// return lipgloss.JoinVertical(lipgloss.Top, m.list.View(), fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
		return lipgloss.JoinHorizontal(lipgloss.Left, m.list.View(), fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView()))
	} else {
		return fmt.Sprintf(
			"Enter Keyword\n\n%s\n\n%s",
			m.textInput.View(),
			"(esc to quit)",
		) + "\n"
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	_, err := p.Run()
	if err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
