# GoAggregator Frontend

**Next.js 15 (App Router) + Tailwind CSS v4** frontend for the GoAggregator flight and hotel price aggregation service.

## Overview

This is the web interface for GoAggregator, providing a modern, responsive UI for searching and comparing flight and hotel prices across multiple providers. Built with Next.js 15 App Router, TypeScript, and Tailwind CSS v4.

## Design System

The frontend uses a premium design system with the following characteristics:

- **Visual Style**: Athletic, diplomatic, sunlit qualities with refined sport-club utility
- **Design Source**: [DESIGN.md](../DESIGN.md) — all tokens are mirrored in [`tokens/design-tokens.ts`](./tokens/design-tokens.ts)
- **Theme Exposure**: Design tokens exposed as Tailwind theme extensions and CSS variables

## Quick Start

### Prerequisites

- Node.js 18+
- npm or yarn

### Installation

```bash
# Navigate to frontend directory
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) to view the landing page.

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start Next.js development server (port 3000) |
| `npm run build` | Create production build |
| `npm run start` | Run production server |
| `npm run lint` | Run ESLint via `next lint` |
| `npm run typecheck` | Run TypeScript type checking (`tsc --noEmit`) |

## Project Structure

```
frontend/
├── app/
│   ├── layout.tsx          # Root layout with Montserrat/Inter fonts
│   ├── page.tsx            # Landing page (6 sections)
│   └── globals.css         # Tailwind base styles + CSS theme variables
├── components/
│   └── ui/
│       ├── Button.tsx       # primary / secondary / ghost variants
│       ├── Card.tsx         # default / muted / primary-light / neutral variants
│       ├── Badge.tsx        # pill-style badge (5 variants)
│       ├── Input.tsx        # text input with focus + invalid states
│       └── SectionHeading.tsx  # headline-display/lg/md/sm wrapper
├── lib/
│   ├── tokens.ts            # Re-export design tokens
│   └── cn.ts               # clsx + tailwind-merge utility helper
├── tokens/
│   └── design-tokens.ts     # Source-of-truth design tokens
├── tailwind.config.ts       # Tailwind v4 theme extension
├── postcss.config.mjs       # PostCSS configuration
├── next.config.ts          # Next.js configuration
├── tsconfig.json           # TypeScript configuration
└── package.json            # Dependencies and scripts
```

## Design Tokens

### Colors

| Token | Value | Usage |
|-------|-------|-------|
| Primary | `#426785` | Key actions, active navigation, important states |
| Primary Dark | `#2D465A` | Hover states |
| Primary Light | `#AABBC8` | Subtle accents |
| Secondary | `#8F4D39` | Labels, secondary highlights |
| Tertiary | `#294258` | Atmospheric accents |
| Surface | `#D3C7B4` | Base background |
| Surface Muted | `#C7BFB0` | Alternate sections |
| Surface Elevated | `#FFFFFF` | Cards, foreground panels |
| Border | `#B4AFA3` | Dividers, card outlines |
| Error | `#B3261E` | Error states |

### Typography

| Style | Font | Size | Weight | Usage |
|-------|------|------|--------|-------|
| headline-display | Montserrat | 64px | 600 | Hero titles, campaign statements |
| headline-lg | Montserrat | 48px | 600 | Section introductions |
| headline-md | Montserrat | 36px | 600 | Card headings |
| headline-sm | Montserrat | 28px | 600 | Compact headings |
| body-lg | Inter | 18px | 400 | Editorial content |
| body-md | Inter | 16px | 400 | Body text |
| body-sm | Inter | 14px | 400 | Supporting text |
| label-lg | Inter | 14px | 600 | Navigation, metadata |
| label-md | Inter | 12px | 600 | Labels |
| label-sm | Inter | 11px | 600 | Small labels |
| button | Inter | 14px | 700 | Action text |

### Spacing

| Token | Value |
|-------|-------|
| xs | 4px |
| sm | 8px |
| md | 16px |
| lg | 24px |
| xl | 32px |
| 2xl | 48px |
| 3xl | 64px |
| 4xl | 96px |

### Border Radius

| Token | Value |
|-------|-------|
| none | 0px |
| sm | 4px |
| md | 8px |
| lg | 12px |
| xl | 20px |
| full | 9999px |

## Components

### Button

Primary buttons for key actions:
```tsx
import { Button } from '@/components/ui/Button';

<Button variant="primary">Search Flights</Button>
<Button variant="secondary">View Details</Button>
<Button variant="ghost">Cancel</Button>
```

### Card

Versatile card component:
```tsx
import { Card } from '@/components/ui/Card';

<Card variant="default">Content here</Card>
<Card variant="muted">Muted background</Card>
<Card variant="primary-light">Light accent</Card>
<Card variant="neutral">Neutral surface</Card>
```

### Badge

Pill-style badges for labels:
```tsx
import { Badge } from '@/components/ui/Badge';

<Badge variant="default">Available</Badge>
<Badge variant="primary">Featured</Badge>
<Badge variant="secondary">Popular</Badge>
<Badge variant="muted">Default</Badge>
<Badge variant="neutral">Neutral</Badge>
```

### Input

Form inputs with validation:
```tsx
import { Input } from '@/components/ui/Input';

<Input placeholder="Enter city" />
<Input invalid error="This field is required" />
```

### SectionHeading

Typography hierarchy:
```tsx
import { SectionHeading } from '@/components/ui/SectionHeading';

<SectionHeading variant="display">Hero Title</SectionHeading>
<SectionHeading variant="lg">Large Heading</SectionHeading>
<SectionHeading variant="md">Medium Heading</SectionHeading>
<SectionHeading variant="sm">Small Heading</SectionHeading>
```

## Backend Connection

The frontend connects to the GoAggregator API at `http://localhost:8081`.

### API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /healthz` | Health check |
| `GET /search/flights` | Search flights |
| `GET /search/hotels` | Search hotels |
| `GET /providers/status` | Provider circuit breaker status |

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_API_BASE` | `http://localhost:8081` | Backend API base URL |

## Development Roadmap

### Sprint 0 — Foundation (Complete)
- Project scaffolding
- 5 UI primitives (Button, Card, Badge, Input, SectionHeading)
- Landing page

### Sprint 1 — Core Features (In Progress)
- Flight search UI
- Hotel search UI
- API integration with Go backend

### Sprint 2 — Dashboard (Planned)
- Provider status dashboard
- Chaos control panel
- Real-time metrics

### Sprint 3 — Enhanced Features (Planned)
- User authentication
- Booking flow
- Payment integration

## License

MIT License - See [../LICENSE.md](../LICENSE.md)
