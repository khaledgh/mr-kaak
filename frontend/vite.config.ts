import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import { VitePWA } from "vite-plugin-pwa";
import path from "node:path";

// Vite + React + PWA. The dev server proxies /api to the Go backend so the
// frontend and API share an origin in development (avoids CORS friction).
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "VITE_");
  const port = parseInt(env.VITE_PORT ?? "5173", 10);
  const backendUrl = env.VITE_BACKEND_URL ?? "http://127.0.0.1:8080";
  const appTitle = env.VITE_APP_TITLE ?? "Arabic Sweets & Kaak";
  const themeColor = env.VITE_THEME_COLOR ?? "#b45309";

  return {
    plugins: [
      react(),
      VitePWA({
        registerType: "autoUpdate",
        includeAssets: ["favicon.svg"],
        manifest: {
          name: appTitle,
          short_name: env.VITE_APP_NAME ?? "Sweets&Kaak",
          description: "Order Lebanese kaak and Arabic sweets",
          theme_color: themeColor,
          background_color: "#fffaf3",
          display: "standalone",
          start_url: "/",
          icons: [
            { src: "/icon-192.png", sizes: "192x192", type: "image/png" },
            { src: "/icon-512.png", sizes: "512x512", type: "image/png" },
          ],
        },
        workbox: {
          // App-shell precache + runtime cache for images and the menu (offline browsing).
          globPatterns: ["**/*.{js,css,html,svg,png,woff2}"],
          runtimeCaching: [
            {
              urlPattern: ({ url }) => url.pathname.startsWith("/api/v1/menu"),
              handler: "StaleWhileRevalidate",
              options: { cacheName: "menu-cache" },
            },
            {
              urlPattern: ({ request }) => request.destination === "image",
              handler: "CacheFirst",
              options: { cacheName: "image-cache", expiration: { maxEntries: 200 } },
            },
          ],
        },
      }),
    ],
    resolve: { alias: { "@": path.resolve(__dirname, "src") } },
    server: {
      port,
      proxy: { "/api": backendUrl },
    },
  };
});

