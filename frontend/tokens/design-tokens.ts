// Design tokens extracted from /DESIGN.md (Desert Fox Club — version: alpha)
// These are the normative values used across Tailwind config and TS components.

export const colors = {
  primary: "#426785",
  "primary-dark": "#2D465A",
  "primary-light": "#AABBC8",
  secondary: "#8F4D39",
  tertiary: "#294258",
  neutral: "#D3C7B4",
  surface: "#D3C7B4",
  "surface-muted": "#C7BFB0",
  "surface-elevated": "#FFFFFF",
  "on-surface": "#000000",
  "on-muted": "#000000",
  "on-primary": "#FFFFFF",
  border: "#B4AFA3",
  "border-strong": "#999A95",
  error: "#B3261E",
} as const;

export const typography = {
  "headline-display": {
    fontFamily: "Montserrat",
    fontSize: "64px",
    fontWeight: 600,
    lineHeight: 1.05,
    letterSpacing: "-0.03em",
  },
  "headline-lg": {
    fontFamily: "Montserrat",
    fontSize: "48px",
    fontWeight: 600,
    lineHeight: 1.1,
    letterSpacing: "-0.02em",
  },
  "headline-md": {
    fontFamily: "Montserrat",
    fontSize: "36px",
    fontWeight: 600,
    lineHeight: 1.15,
    letterSpacing: "-0.01em",
  },
  "headline-sm": {
    fontFamily: "Montserrat",
    fontSize: "28px",
    fontWeight: 600,
    lineHeight: 1.2,
    letterSpacing: "0em",
  },
  "body-lg": {
    fontFamily: "Inter",
    fontSize: "18px",
    fontWeight: 400,
    lineHeight: 1.6,
    letterSpacing: "0em",
  },
  "body-md": {
    fontFamily: "Inter",
    fontSize: "16px",
    fontWeight: 400,
    lineHeight: 1.6,
    letterSpacing: "0em",
  },
  "body-sm": {
    fontFamily: "Inter",
    fontSize: "14px",
    fontWeight: 400,
    lineHeight: 1.5,
    letterSpacing: "0em",
  },
  "label-lg": {
    fontFamily: "Inter",
    fontSize: "14px",
    fontWeight: 600,
    lineHeight: 1.2,
    letterSpacing: "0.04em",
  },
  "label-md": {
    fontFamily: "Inter",
    fontSize: "12px",
    fontWeight: 600,
    lineHeight: 1.2,
    letterSpacing: "0.08em",
  },
  "label-sm": {
    fontFamily: "Inter",
    fontSize: "11px",
    fontWeight: 600,
    lineHeight: 1.1,
    letterSpacing: "0.1em",
  },
  button: {
    fontFamily: "Inter",
    fontSize: "14px",
    fontWeight: 700,
    lineHeight: 1,
    letterSpacing: "0.02em",
  },
} as const;

export const rounded = {
  none: "0px",
  sm: "4px",
  md: "8px",
  lg: "12px",
  xl: "20px",
  full: "9999px",
} as const;

export const spacing = {
  xs: "4px",
  sm: "8px",
  md: "16px",
  lg: "24px",
  xl: "32px",
  "2xl": "48px",
  "3xl": "64px",
  "4xl": "96px",
  gutter: "24px",
  section: "96px",
  container: "1200px",
} as const;

export const designTokens = {
  colors,
  typography,
  rounded,
  spacing,
} as const;

export type ColorToken = keyof typeof colors;
export type TypographyToken = keyof typeof typography;
export type SpacingToken = keyof typeof spacing;
