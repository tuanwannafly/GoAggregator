import type { Config } from "tailwindcss";
import { colors, rounded, spacing } from "./tokens/design-tokens";

const config: Config = {
  content: [
    "./app/**/*.{ts,tsx}",
    "./components/**/*.{ts,tsx}",
    "./lib/**/*.{ts,tsx}",
  ],
  theme: {
    extend: {
      colors,
      borderRadius: {
        none: rounded.none,
        sm: rounded.sm,
        md: rounded.md,
        lg: rounded.lg,
        xl: rounded.xl,
        full: rounded.full,
      },
      spacing: {
        xs: spacing.xs,
        sm: spacing.sm,
        md: spacing.md,
        lg: spacing.lg,
        xl: spacing.xl,
        "2xl": spacing["2xl"],
        "3xl": spacing["3xl"],
        "4xl": spacing["4xl"],
        gutter: spacing.gutter,
        section: spacing.section,
      },
      maxWidth: {
        container: spacing.container,
      },
      fontFamily: {
        display: ["var(--font-montserrat)", "system-ui", "sans-serif"],
        sans: ["var(--font-inter)", "system-ui", "sans-serif"],
      },
      container: {
        center: true,
        padding: {
          DEFAULT: "1rem",
          lg: "1.5rem",
        },
        screens: {
          sm: "640px",
          md: "768px",
          lg: "1024px",
          xl: "1280px",
          "2xl": "1200px",
        },
      },
    },
  },
  plugins: [],
};

export default config;
