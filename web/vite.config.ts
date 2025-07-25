import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import tailwindcss from "@tailwindcss/vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  server: {
    port: 8080,
    proxy: {
      "/api": "http://127.0.0.1:8080/",
    },
  },
});
