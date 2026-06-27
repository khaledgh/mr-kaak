import type { Config } from "tailwindcss";

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
    },
  },
  plugins: [],
} satisfies Config;
