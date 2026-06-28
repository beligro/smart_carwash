// @lovable.dev/vite-tanstack-config already includes the following — do NOT add them manually
// or the app will break with duplicate plugins:
//   - tanstackStart, viteReact, tailwindcss, tsConfigPaths, nitro (build-only using cloudflare as a default target),
//     componentTagger (dev-only), VITE_* env injection, @ path alias, React/TanStack dedupe,
//     error logger plugins, and sandbox detection (port/host/strictPort).
// You can pass additional config via defineConfig({ vite: { ... }, etc... }) if needed.
import { defineConfig } from "@lovable.dev/vite-tanstack-config";

export default defineConfig({
  // Drop the Nitro deploy layer entirely. With Nitro the SSR bundle is
  // redirected to .output/server/index.mjs, but TanStack Start's prerenderer
  // expects a runnable server at dist/server/server.js. Without Nitro, the
  // start plugin builds client -> dist/client, server -> dist/server/server.js,
  // then prerenders the route into dist/client/index.html. We ship dist/client
  // as a fully static bundle; dist/server is unused at runtime.
  nitro: false,
  vite: { base: "/" },
  tanstackStart: {
    // Redirect TanStack Start's bundled server entry to src/server.ts (our SSR error wrapper).
    server: { entry: "server" },
    // Prerender the marketing route to static HTML.
    pages: [{ path: "/" }],
    prerender: {
      enabled: true,
      crawlLinks: false,
      failOnError: false,
    },
  },
});
