import tailwindcss from "@tailwindcss/vite";
import { tanstackStart } from "@tanstack/react-start/plugin/vite";
import viteReact from "@vitejs/plugin-react";
import { nitro } from "nitro/vite";
import { defineConfig } from "vite";
import tsconfigPaths from "vite-tsconfig-paths";

export default defineConfig({
  resolve: {
    dedupe: ["react", "react-dom", "@tanstack/react-router", "@tanstack/react-query"],
  },
  plugins: [
    tanstackStart({
      // Custom SSR entry wraps TanStack Start's handler for error pages (src/server.ts).
      server: { entry: "server" },
    }),
    viteReact(),
    tailwindcss(),
    tsconfigPaths(),
    nitro(),
  ],
});
