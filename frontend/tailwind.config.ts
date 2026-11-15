import type { Config } from "tailwindcss";

const config: Config = {
  darkMode: ["class"],
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        background: "hsl(var(--background))",
        foreground: "hsl(var(--foreground))",
        card: {
          DEFAULT: "hsl(var(--card))",
          foreground: "hsl(var(--card-foreground))",
        },
        muted: {
          DEFAULT: "hsl(var(--muted))",
          foreground: "hsl(var(--muted-foreground))",
        },
        border: "hsl(var(--border))",
        input: "hsl(var(--input))",
        ring: "hsl(var(--ring))",
        accent: {
          DEFAULT: "hsl(var(--accent))",
          foreground: "hsl(var(--accent-foreground))",
        },
        primary: {
          DEFAULT: "hsl(var(--primary))",
          foreground: "hsl(var(--primary-foreground))",
        },
        destructive: {
          DEFAULT: "hsl(var(--destructive))",
          foreground: "hsl(var(--destructive-foreground))",
        },
        cyan: {
          DEFAULT: "hsl(var(--cyan))",
          muted: "hsl(var(--cyan-muted))",
        },
        quantum: {
          DEFAULT: "hsl(var(--quantum))",
          muted: "hsl(var(--quantum-muted))",
        },
        elevated: "hsl(var(--elevated))",
        lower: "hsl(var(--lower))",
        unknown: "hsl(var(--unknown))",
        review: "hsl(var(--review))",
      },
      fontFamily: {
        sans: ["var(--font-geist-sans)", "IBM Plex Sans", "Segoe UI", "sans-serif"],
        mono: ["var(--font-geist-mono)", "IBM Plex Mono", "ui-monospace", "monospace"],
        display: ["var(--font-geist-sans)", "IBM Plex Sans", "sans-serif"],
      },
      boxShadow: {
        panel: "0 0 0 1px hsl(var(--border)), 0 12px 40px rgba(0,0,0,0.35)",
      },
      borderRadius: {
        lg: "var(--radius)",
        md: "calc(var(--radius) - 2px)",
        sm: "calc(var(--radius) - 4px)",
      },
    },
  },
  plugins: [],
};

export default config;
