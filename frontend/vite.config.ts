import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

export defailt defineConfig({
  plugins: [react(), tailwindcss()],
});
