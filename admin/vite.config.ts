import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";
import path from "node:path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "VITE_");
  const port = parseInt(env.VITE_PORT ?? "5174", 10);
  const backendUrl = env.VITE_BACKEND_URL ?? "http://127.0.0.1:8080";

  return {
    plugins: [react()],
    resolve: { alias: { "@": path.resolve(__dirname, "src") } },
    server: {
      port,
      proxy: { "/api": backendUrl, "/uploads": backendUrl },
    },
  };
});

