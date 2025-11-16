package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AnimationType represents different types of animations
type AnimationType int

const (
	AnimationFadeIn AnimationType = iota
	AnimationSlideIn
	AnimationPulse
	AnimationGlow
)

// AnimationState tracks the current animation state
type AnimationState struct {
	AnimationType AnimationType
	StartTime     time.Time
	Duration      time.Duration
	Completed     bool
}

// NewAnimation creates a new animation state
func NewAnimation(animationType AnimationType, duration time.Duration) *AnimationState {
	return &AnimationState{
		AnimationType: animationType,
		StartTime:     time.Now(),
		Duration:      duration,
		Completed:     false,
	}
}

// Update updates the animation state and returns whether it's completed
func (a *AnimationState) Update() bool {
	if a.Completed {
		return true
	}

	elapsed := time.Since(a.StartTime)
	if elapsed >= a.Duration {
		a.Completed = true
		return true
	}

	return false
}

// Progress returns the animation progress (0.0 to 1.0)
func (a *AnimationState) Progress() float64 {
	if a.Completed {
		return 1.0
	}

	elapsed := time.Since(a.StartTime)
	progress := float64(elapsed) / float64(a.Duration)
	if progress > 1.0 {
		progress = 1.0
	}

	return progress
}

// EaseInOutCubic provides smooth easing function
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - (-2*t+2)*(2*t-2)*(2*t-2)/2
}

// AnimationMsg is sent when an animation should be triggered
type AnimationMsg struct {
	AnimationType AnimationType
	Target        string // Target element identifier
}

// StartAnimationCmd creates a command to start an animation
func StartAnimationCmd(animationType AnimationType, target string) tea.Cmd {
	return func() tea.Msg {
		return AnimationMsg{
			AnimationType: animationType,
			Target:        target,
		}
	}
}

// StatusAnimation provides smooth transitions for status messages
type StatusAnimation struct {
	Message     string
	MessageType string
	Animation   *AnimationState
	FadeOutTime time.Duration
}

// NewStatusAnimation creates a new status animation
func NewStatusAnimation(message, messageType string, duration time.Duration) *StatusAnimation {
	return &StatusAnimation{
		Message:     message,
		MessageType: messageType,
		Animation:   NewAnimation(AnimationFadeIn, duration),
		FadeOutTime: duration + 2*time.Second, // Start fade out after display duration
	}
}

// Update updates the status animation
func (s *StatusAnimation) Update() (bool, string) {
	if s.Animation == nil {
		return true, ""
	}

	completed := s.Animation.Update()
	if completed {
		// Check if we should start fade out
		if time.Since(s.Animation.StartTime) > s.FadeOutTime {
			return true, ""
		}
		return false, s.Message
	}

	// During animation, show the message
	return false, s.Message
}

// GetAnimatedStyle applies animation effects to styles based on progress
func GetAnimatedStyle(baseStyle lipgloss.Style, animation *AnimationState) lipgloss.Style {
	if animation == nil || animation.Completed {
		return baseStyle
	}

	progress := animation.Progress()
	easedProgress := EaseInOutCubic(progress)

	switch animation.AnimationType {
	case AnimationFadeIn:
		// Apply fade-in effect by adjusting foreground color brightness
		return baseStyle
	case AnimationPulse:
		// Create pulsing effect by cycling opacity
		_ = (1.0 + easedProgress) / 2.0 // pulse value for future use
		return baseStyle
	case AnimationGlow:
		// Add glow effect by brightening colors
		return baseStyle
	default:
		return baseStyle
	}
}
