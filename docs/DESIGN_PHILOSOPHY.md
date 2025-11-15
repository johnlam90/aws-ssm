# TUI Design Philosophy

## Jony Ive's Minimalist Approach

> "Simplicity is not the absence of clutter, that's a consequence of simplicity. Simplicity is somehow essentially describing the purpose and place of an object and product. The absence of clutter is just a clutter-free product. That's not simple."
> — Jony Ive

This TUI follows Jony Ive's design principles: purposeful minimalism, functional color usage, and refined aesthetics that enhance usability without distraction.

## Core Design Principles

### 1. **Purposeful Color Usage**
Color is used sparingly and only where it adds functional value. Every color choice has a specific semantic meaning that helps users understand the interface at a glance.

### 2. **Visual Hierarchy**
Color creates clear hierarchy and guides user attention to the most important elements:
- **White** (#FFFFFF): Primary content, titles
- **Light Gray** (#E5E7EB): Body text, readable content
- **Cool Gray** (#9CA3AF): Secondary information
- **Muted Gray** (#6B7280): Tertiary text, de-emphasized content

### 3. **Functional Accents**
Subtle color accents communicate state and interactivity:
- **Blue** (#60A5FA): Interactive elements, focus, selection
- **Green** (#34D399): Success states, active/running resources
- **Amber** (#FBBF24): Warnings, pending/transitional states
- **Red** (#F87171): Errors, stopped/terminated states
- **Indigo** (#818CF8): Information, keybindings, helpful references

### 4. **High Contrast for Accessibility**
All color combinations meet WCAG AA standards for contrast, ensuring readability in various lighting conditions and for users with visual impairments.

### 5. **Consistency**
Each color has a consistent semantic meaning throughout the interface, creating a predictable and learnable system.

## Color Palette

### Monochromatic Base
```
#FFFFFF  Pure White       Primary text, titles
#E5E7EB  Light Gray       Body text, content
#9CA3AF  Cool Gray        Secondary text
#6B7280  Muted Gray       Tertiary text
#374151  Border Gray      Subtle borders
#1F2937  Selection Gray   Selected item background
#1E3A8A  Focus Blue       Focused element background
#000000  Pure Black       Background
```

### Functional Accents
```
#60A5FA  Soft Blue        Interactive, selection, focus
#34D399  Refined Green    Success, active, running
#FBBF24  Warm Amber       Warning, pending, attention
#F87171  Soft Red         Error, stopped, inactive
#818CF8  Subtle Indigo    Information, keybindings
```

## Application of Principles

### Interactive Elements
- **Selected menu items**: Blue foreground + deep blue background
- **Selected list items**: Blue foreground + subtle gray background
- **Keybindings**: Indigo color makes them scannable
- **Selection indicator**: Blue arrow (▸) guides the eye

### State Communication
- **Running instances**: Green - healthy and operational
- **Stopped instances**: Gray - inactive, neutral
- **Pending states**: Amber - transitional, needs attention
- **Terminated**: Muted gray - ended, archived
- **Errors**: Soft red - clear but not alarming

### Information Hierarchy
- **Titles**: Bold white - maximum prominence
- **Table headers**: Bold gray with bottom border - clear structure
- **Body content**: Light gray - readable, not harsh
- **Help text**: Muted gray - available but not distracting

### Feedback & Messages
- **Success messages**: Green + bold
- **Error messages**: Red + bold
- **Warnings**: Amber + bold
- **Info messages**: Indigo
- **Status bar**: Muted gray - stays out of the way

## Design Rationale

### Why Blue for Interactive Elements?
Blue is universally recognized as the color of interactivity in digital interfaces. It signals "you can act here" without being aggressive or demanding attention.

### Why Green for Running States?
Green has strong cultural associations with "go," "healthy," and "operational." It provides instant visual feedback that a resource is active and functioning correctly.

### Why Amber for Warnings?
Amber (yellow-orange) is the universal color for caution. It draws attention without the urgency of red, perfect for transitional states that need monitoring but aren't critical.

### Why Indigo for Keybindings?
Indigo is distinct from the primary blue used for selection, making keybindings easily scannable while maintaining the refined, professional aesthetic.

### Why Muted Tones?
Bright, saturated colors cause visual fatigue during extended use. Muted tones are easier on the eyes while still providing clear differentiation and semantic meaning.

## Comparison to Other TUIs

### k9s
- Uses bright, saturated colors
- More colorful overall
- Our approach: More refined, less visual noise

### Claude Code CLI
- Minimal color usage
- Professional aesthetic
- Our approach: Similar restraint with purposeful accents

### Droid CLI
- Clean, monochromatic
- Subtle highlights
- Our approach: Aligned philosophy with semantic color meanings

## Testing the Design

To see the design in action:
```bash
./aws-ssm tui
```

Navigate through different views to observe:
- Blue selection indicators
- Green running instances
- Amber pending states
- Indigo keybindings in help text
- Refined gray hierarchy

## Future Enhancements

Potential areas for subtle color refinement:
- Capacity metrics (highlight when near limits)
- Search results (subtle highlight on matches)
- Diff indicators (for configuration changes)
- Time-based indicators (recent vs. old data)

All future enhancements will follow the same principles: color must serve a functional purpose and enhance usability without adding visual clutter.

