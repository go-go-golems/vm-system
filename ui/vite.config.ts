import { jsxLocPlugin } from "@builder.io/vite-plugin-jsx-loc";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { defineConfig, loadEnv } from "vite";

const plugins = [react(), tailwindcss(), jsxLocPlugin()];

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, import.meta.dirname, "");
  const proxyTarget = env.VM_SYSTEM_API_PROXY_TARGET || "http://127.0.0.1:3210";

  return {
    plugins,
    resolve: {
      alias: {
        "@": path.resolve(import.meta.dirname, "client", "src"),
        "@shared": path.resolve(import.meta.dirname, "shared"),
        "@assets": path.resolve(import.meta.dirname, "attached_assets"),
      },
    },
    envDir: path.resolve(import.meta.dirname),
    root: path.resolve(import.meta.dirname, "client"),
    build: {
      outDir: path.resolve(import.meta.dirname, "dist/public"),
      emptyOutDir: true,
    },
    server: {
      port: 3000,
      strictPort: false, // Will find next available port if 3000 is busy
      host: true,
      allowedHosts: ["localhost", "127.0.0.1"],
      proxy: {
        "/api/v1": {
          target: proxyTarget,
          changeOrigin: true,
        },
      },
      fs: {
        strict: true,
        deny: ["**/.*"],
      },
    },
  };
});
