import type { Config } from "tailwindcss";

// Tailwind v3 with a warm amber brand palette matching the menu photography.
// RTL is handled via logical utilities (ps-/pe-/ms-/me-) + dir="rtl" on <html>.
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          50: "#fffaf3",
          100: "#fdeedb",
          200: "#fad7a8",
          300: "#f6bb6f",
          400: "#f0993c",
          500: "#e07b1a",
          600: "#b45309",
          700: "#8a3f0c",
          800: "#6f3410",
          900: "#5b2c11",
        },
      },
      fontFamily: {
        sans: ["Inter", "system-ui", "sans-serif"],
        arabic: ["Cairo", "Tajawal", "system-ui", "sans-serif"],
      },
    },
  },
  plugins: [],
} satisfies Config;
