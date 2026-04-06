package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/leeozaka/gommits/internal/models"
)

var (
	toastStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#38A169")).
			Padding(1, 3).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#2F855A")).
			Bold(true).
			Align(lipgloss.Center)

	toastErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#E53E3E")).
			Padding(1, 3).
			Margin(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#C53030")).
			Bold(true).
			Align(lipgloss.Center)
)

const (
	toastTickInterval            = 50 * time.Millisecond
	slideInDuration              = 300 * time.Millisecond
	fadeInDuration               = 200 * time.Millisecond
	fadeOutDuration              = 500 * time.Millisecond
	toastBgR, toastBgG, toastBgB = 0x1A, 0x1A, 0x1A
)

type ToastManager struct {
	toast models.Toast
}

func NewToastManager() ToastManager {
	return ToastManager{
		toast: models.Toast{
			Visible:  false,
			Opacity:  0.0,
			Position: 0.0,
			Duration: 3 * time.Second,
		},
	}
}

func (tm ToastManager) IsVisible() bool {
	return tm.toast.Visible && tm.toast.Opacity > 0
}

func (tm ToastManager) Update(msg tea.Msg) (ToastManager, tea.Cmd) {
	switch msg := msg.(type) {
	case models.ShowToastMsg:
		tm.toast = models.Toast{
			Message:   msg.Message,
			Type:      msg.Type,
			Visible:   true,
			Opacity:   0.0,
			Position:  0.0,
			StartTime: time.Now(),
			Duration:  msg.Duration,
		}
		return tm, tea.Batch(
			hideToastCmd(msg.Duration),
			tickCmd(toastTickInterval),
		)

	case models.HideToastMsg:
		tm.toast.Visible = false
		tm.toast.Opacity = 0.0
		tm.toast.Position = 0.0
		return tm, nil

	case models.TickMsg:
		if !tm.toast.Visible {
			return tm, nil
		}

		elapsed := time.Since(tm.toast.StartTime)
		if elapsed >= tm.toast.Duration {
			tm.toast.Visible = false
			tm.toast.Opacity = 0.0
			tm.toast.Position = 0.0
			return tm, nil
		}

		if elapsed < slideInDuration {
			p := float64(elapsed) / float64(slideInDuration)
			tm.toast.Position = 1 - (1-p)*(1-p)*(1-p)
		} else {
			tm.toast.Position = 1.0
		}

		if elapsed < fadeInDuration {
			tm.toast.Opacity = float64(elapsed) / float64(fadeInDuration)
		} else {
			fadeOutStart := tm.toast.Duration - fadeOutDuration
			if elapsed >= fadeOutStart {
				fadeProgress := float64(elapsed-fadeOutStart) / float64(fadeOutDuration)
				tm.toast.Opacity = 1.0 - fadeProgress
			} else {
				tm.toast.Opacity = 1.0
			}
		}

		return tm, tickCmd(toastTickInterval)
	}

	return tm, nil
}

func (tm ToastManager) View() string {
	if !tm.IsVisible() {
		return ""
	}

	style := toastStyle
	if tm.toast.Type == models.ToastError {
		style = toastErrorStyle
	}

	if tm.toast.Opacity < 1.0 {
		style = tm.applyOpacity(style)
	}

	return style.Render(tm.toast.Message)
}

// toastViewModel wraps the toast rendered string as a tea.Model for use with bubbletea-overlay.
type toastViewModel struct {
	content string
}

func (t toastViewModel) Init() tea.Cmd                       { return nil }
func (t toastViewModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return t, nil }
func (t toastViewModel) View() string                        { return t.content }

// backgroundViewModel wraps a pre-rendered string as a tea.Model for use with bubbletea-overlay.
type backgroundViewModel struct {
	content string
}

func (b backgroundViewModel) Init() tea.Cmd                       { return nil }
func (b backgroundViewModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return b, nil }
func (b backgroundViewModel) View() string                        { return b.content }

func (tm ToastManager) applyOpacity(style lipgloss.Style) lipgloss.Style {
	opacity := tm.toast.Opacity

	var fgR, fgG, fgB int
	var borderR, borderG, borderB int
	if tm.toast.Type == models.ToastSuccess {
		fgR, fgG, fgB = 0x38, 0xA1, 0x69
		borderR, borderG, borderB = 0x2F, 0x85, 0x5A
	} else {
		fgR, fgG, fgB = 0xE5, 0x3E, 0x3E
		borderR, borderG, borderB = 0xC5, 0x30, 0x30
	}

	style = style.
		Background(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
			int(float64(fgR)*opacity+float64(toastBgR)*(1-opacity)),
			int(float64(fgG)*opacity+float64(toastBgG)*(1-opacity)),
			int(float64(fgB)*opacity+float64(toastBgB)*(1-opacity))))).
		BorderForeground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
			int(float64(borderR)*opacity+float64(toastBgR)*(1-opacity)),
			int(float64(borderG)*opacity+float64(toastBgG)*(1-opacity)),
			int(float64(borderB)*opacity+float64(toastBgB)*(1-opacity)))))

	textR, textG, textB := 0xFA, 0xFA, 0xFA
	style = style.
		Foreground(lipgloss.Color(fmt.Sprintf("#%02x%02x%02x",
			int(float64(textR)*opacity+float64(toastBgR)*(1-opacity)),
			int(float64(textG)*opacity+float64(toastBgG)*(1-opacity)),
			int(float64(textB)*opacity+float64(toastBgB)*(1-opacity)))))

	return style
}

func hideToastCmd(delay time.Duration) tea.Cmd {
	return tea.Tick(delay, func(t time.Time) tea.Msg {
		return models.HideToastMsg{}
	})
}

func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return models.TickMsg(t)
	})
}
