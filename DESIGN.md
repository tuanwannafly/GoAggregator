# GoAggregator Design System

**Design tokens, typography, spacing, and component guidelines for the GoAggregator frontend.**

---

## Table of Contents

- [Overview](#overview)
- [Design Philosophy](#design-philosophy)
- [Color System](#color-system)
- [Typography](#typography)
- [Spacing](#spacing)
- [Border Radius](#border-radius)
- [Components](#components)
- [Usage Guidelines](#usage-guidelines)
- [Implementation](#implementation)

---

## Overview

The GoAggregator design system provides a cohesive visual language for the flight and hotel search frontend. It emphasizes clarity, speed, and trust — essential qualities for a price comparison service.

### Design Characteristics

- **Premium and Professional**: Clean, trustworthy appearance suitable for financial transactions
- **Fast and Efficient**: Clear visual hierarchy for quick scanning
- **Athletic and Dynamic**: Subtle sport-inspired accents reflecting travel energy
- **Accessible**: WCAG AA compliant contrast ratios

---

## Design Philosophy

### Core Principles

1. **Clarity First**: Every element serves a purpose
2. **Fast Scanning**: Visual hierarchy enables quick decision-making
3. **Trust-Building**: Professional design instills confidence
4. **Consistency**: Unified experience across all pages
5. **Accessibility**: Inclusive design for all users

### Visual Identity

The design draws inspiration from:
- Premium travel booking platforms
- Athletic club memberships
- Diplomatic passport aesthetics
- Sunlit, warm color temperatures

---

## Color System

### Color Palette

| Token | Hex | Usage |
|-------|-----|-------|
| Primary | `#426785` | Main brand color, key actions |
| Primary Dark | `#2D465A` | Hover states, emphasis |
| Primary Light | `#AABBC8` | Subtle backgrounds, accents |
| Secondary | `#8F4D39` | Labels, secondary highlights |
| Tertiary | `#294258` | Decorative accents |
| Surface | `#D3C7B4` | Page backgrounds |
| Surface Muted | `#C7BFB0` | Alternate sections |
| Surface Elevated | `#FFFFFF` | Cards, foreground panels |
| On Surface | `#000000` | Primary text |
| On Muted | `#757E81` | Secondary text, captions |
| Border | `#B4AFA3` | Dividers, outlines |
| Border Strong | `#999A95` | Emphasis borders |
| Error | `#B3261E` | Error states |

### Color Usage Guidelines

**Primary Color Usage**:
- Primary buttons and key actions
- Active navigation states
- Important data highlights
- Signature panels and headers

**Secondary Color Usage**:
- Product tags and labels
- Secondary highlights
- Illustrated details
- Mid-emphasis UI states

**Surface Color Usage**:
- `surface`: Main page backgrounds
- `surface-muted`: Alternate sections, secondary cards
- `surface-elevated`: Cards, modals, floating elements

### Contrast Requirements

| Context | Minimum Ratio | Token Pair |
|---------|----------------|------------|
| Body text | 4.5:1 | `on-surface` on `surface` |
| Large text | 3:1 | `on-surface` on `surface` |
| UI components | 3:1 | `on-primary` on `primary` |
| Error states | 4.5:1 | `on-surface` on `error` |

---

## Typography

### Font Families

| Usage | Font | Fallback |
|-------|------|----------|
| Headlines | Montserrat | system-ui, sans-serif |
| Body/UI | Inter | system-ui, sans-serif |

### Type Scale

| Token | Font | Size | Weight | Line Height | Letter Spacing | Usage |
|-------|------|------|--------|-------------|----------------|-------|
| headline-display | Montserrat | 64px | 600 | 1.05 | -0.03em | Hero titles, campaign statements |
| headline-lg | Montserrat | 48px | 600 | 1.1 | -0.02em | Section introductions |
| headline-md | Montserrat | 36px | 600 | 1.15 | -0.01em | Card headings, feature titles |
| headline-sm | Montserrat | 28px | 600 | 1.2 | 0em | Compact headings |
| body-lg | Inter | 18px | 400 | 1.6 | 0em | Editorial content |
| body-md | Inter | 16px | 400 | 1.6 | 0em | Body text, descriptions |
| body-sm | Inter | 14px | 400 | 1.5 | 0em | Supporting text, captions |
| label-lg | Inter | 14px | 600 | 1.2 | 0.04em | Navigation, metadata |
| label-md | Inter | 12px | 600 | 1.2 | 0.08em | Labels, badges |
| label-sm | Inter | 11px | 600 | 1.1 | 0.1em | Small labels, tags |
| button | Inter | 14px | 700 | 1 | 0.02em | Button text |

### Typography Guidelines

**Hierarchy**:
- Use `headline-*` for major section headings
- Use `body-*` for content and descriptions
- Use `label-*` for UI elements and annotations

**Best Practices**:
- Maintain consistent hierarchy throughout
- Avoid mixing font weights unexpectedly
- Use letter spacing sparingly
- Keep body text at comfortable reading size (16px minimum)

---

## Spacing

### Spacing Scale

| Token | Value | Usage |
|-------|-------|-------|
| xs | 4px | Tight spacing, icon gaps |
| sm | 8px | Compact elements |
| md | 16px | Standard padding |
| lg | 24px | Section spacing |
| xl | 32px | Large gaps |
| 2xl | 48px | Major sections |
| 3xl | 64px | Hero sections |
| 4xl | 96px | Page margins |

### Layout Spacing

| Token | Value | Usage |
|-------|-------|-------|
| gutter | 24px | Column spacing |
| section | 96px | Vertical rhythm between sections |
| container | 1200px | Max content width |

### Spacing Guidelines

**Consistent Rhythm**:
- Use multiples of 4px for all spacing
- Establish clear vertical rhythm (24px or 32px base)
- Maintain consistent padding within components

**Responsive**:
- Mobile: Reduce spacing tokens by one level
- Tablet: Use standard spacing
- Desktop: May increase for emphasis

---

## Border Radius

### Radius Scale

| Token | Value | Usage |
|-------|-------|-------|
| none | 0px | Technical panels, certificates |
| sm | 4px | Tight elements, inputs |
| md | 8px | Buttons, cards, normal UI |
| lg | 12px | Feature cards, large panels |
| xl | 20px | Promotional cards |
| full | 9999px | Pills, badges, avatars |

### Radius Usage Guidelines

| Component | Radius | Token |
|-----------|--------|-------|
| Buttons | 8px | md |
| Inputs | 8px | md |
| Cards | 12px | lg |
| Badges | full | full |
| Images | 12px | lg |
| Modals | 12px | lg |

### Best Practices

- Use smaller radius for functional elements
- Use larger radius for emphasis and promotion
- Pills and badges should use `full` radius
- Avoid mixing many radius sizes in same context

---

## Components

### Button

**Primary Button**:
- Background: `primary`
- Text: `on-primary`
- Typography: `button`
- Radius: `md` (8px)
- Padding: 14px 24px
- Height: 48px

**States**:
- Default: Primary background
- Hover: Primary Dark background
- Active: Slight scale down
- Disabled: Reduced opacity (50%)

**Secondary Button**:
- Background: `surface-elevated`
- Text: `on-surface`
- Border: `border`
- Same dimensions as primary

**Ghost Button**:
- Background: transparent
- Text: `on-surface`
- Border: none

### Card

**Default Card**:
- Background: `surface-elevated`
- Text: `on-surface`
- Radius: `lg` (12px)
- Padding: 32px
- Shadow: subtle elevation

**Card Variants**:
- `default`: White background
- `muted`: Muted surface background
- `primary-light`: Light primary accent
- `neutral`: Neutral surface

**Usage**:
- Product information cards
- Feature highlights
- Result displays
- Dashboard panels

### Badge

**Pill Badge**:
- Background: `surface-muted`
- Text: `on-surface`
- Typography: `label-sm`
- Radius: `full` (pill shape)
- Padding: 6px 10px

**Badge Variants**:
- `default`: Neutral styling
- `primary`: Primary accent
- `secondary`: Secondary accent
- `muted`: Muted appearance
- `neutral`: Surface color

**Usage**:
- Status indicators
- Category labels
- Price tags
- Feature highlights

### Input

**Text Input**:
- Background: `surface-elevated`
- Text: `on-surface`
- Typography: `body-md`
- Radius: `md` (8px)
- Padding: 12px 14px
- Border: `border`

**States**:
- Default: Standard border
- Focus: Primary border, subtle shadow
- Invalid: Error border, error message
- Disabled: Reduced opacity

### Section Heading

**Headline Wrapper**:
- Responsive typography scale
- Consistent spacing above heading
- Optional subtitle

**Variants**:
- `display`: `headline-display` (64px)
- `lg`: `headline-lg` (48px)
- `md`: `headline-md` (36px)
- `sm`: `headline-sm` (28px)

---

## Usage Guidelines

### Layout Principles

**Container**:
- Max width: 1200px
- Horizontal padding: 24px
- Centered alignment

**Grid**:
- 4-column mobile
- 8-column tablet
- 12-column desktop
- 24px gutters

**Vertical Rhythm**:
- 96px between major sections
- 48px between card groups
- 24px between related elements

### Responsive Breakpoints

| Breakpoint | Min Width | Columns | Usage |
|-------------|-----------|---------|-------|
| Mobile | 0px | 4 | Smartphones |
| Tablet | 768px | 8 | Tablets, small laptops |
| Desktop | 1024px | 12 | Standard screens |
| Large | 1280px | 12 | Wide screens |

### Accessibility

**Keyboard Navigation**:
- All interactive elements focusable
- Visible focus indicators
- Logical tab order
- Keyboard shortcuts for common actions

**Screen Readers**:
- Semantic HTML elements
- ARIA labels where needed
- Alt text for images
- Form labels

---

## Implementation

### File Structure

```
frontend/
├── tokens/
│   └── design-tokens.ts    # Source of truth for all tokens
├── lib/
│   ├── tokens.ts           # Re-export tokens
│   └── cn.ts               # Class name utility
├── components/ui/
│   ├── Button.tsx
│   ├── Card.tsx
│   ├── Badge.tsx
│   ├── Input.tsx
│   └── SectionHeading.tsx
├── tailwind.config.ts      # Tailwind theme extension
└── app/globals.css         # CSS variables
```

### Using Tokens in Code

**TypeScript**:
```typescript
import { colors, typography } from '@/tokens/design-tokens';

// Use in component
const styles = {
  color: colors.primary,
  ...typography.bodyMd,
};
```

**Tailwind**:
```tsx
<div className="bg-primary text-on-primary p-md rounded-lg">
  Content
</div>
```

**CSS Variables**:
```css
.my-component {
  background-color: var(--color-primary);
  padding: var(--spacing-md);
}
```

### Creating New Components

1. Import design tokens
2. Define component props interface
3. Implement component with token-based styles
4. Export component with variants
5. Add Storybook story (if applicable)
6. Write tests

---

## Do's and Don'ts

### Do

- Use design tokens instead of hardcoded values
- Maintain consistent spacing throughout
- Use appropriate typography hierarchy
- Ensure sufficient contrast ratios
- Test on multiple screen sizes
- Keep components focused and small

### Don't

- Use colors outside the palette
- Mix font families inconsistently
- Use arbitrary spacing values
- Create components without tokens
- Overload pages with elements
- Ignore accessibility guidelines

---

## Changelog

### Version 1.0.0 (Initial)

- Core color palette
- Typography system
- Spacing scale
- Border radius scale
- Button component
- Card component
- Badge component
- Input component
- SectionHeading component

---

## License

This design system is part of the GoAggregator project and is licensed under the MIT License. See [LICENSE.md](LICENSE.md) for details.
