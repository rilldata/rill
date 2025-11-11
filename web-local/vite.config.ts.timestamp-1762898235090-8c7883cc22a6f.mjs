// vite.config.ts
import { sveltekit } from "file:///Users/royendo/Desktop/GitHub/rill/node_modules/@sveltejs/kit/src/exports/vite/index.js";
import dns from "dns";
import { defineConfig } from "file:///Users/royendo/Desktop/GitHub/rill/node_modules/vitest/dist/config.js";
dns.setDefaultResultOrder("verbatim");
var config = defineConfig({
  build: {
    rollupOptions: {
      // This ensures that the web-admin package is not bundled into the web-local package.
      // This is necessary because the Scheduled Reports dialog lives in `web-common` and imports the admin-client.
      external: (id) => id.startsWith("@rilldata/web-admin/")
    }
  },
  resolve: {
    alias: {
      src: "/src",
      // trick to get absolute imports to work
      "@rilldata/web-local": "/src",
      "@rilldata/web-common": "/../web-common/src",
      "@rilldata/web-admin": "/../web-admin/src"
    }
  },
  server: {
    strictPort: true,
    fs: {
      allow: ["."]
    }
  },
  define: {
    "import.meta.env.VITE_PLAYWRIGHT_TEST": process.env.PLAYWRIGHT_TEST,
    "import.meta.env.VITE_PLAYWRIGHT_CLOUD_TEST": process.env.PLAYWRIGHT_CLOUD_TEST
  },
  plugins: [sveltekit()],
  envDir: "../",
  envPrefix: "RILL_UI_PUBLIC_"
});
var vite_config_default = config;
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcudHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCIvVXNlcnMvcm95ZW5kby9EZXNrdG9wL0dpdEh1Yi9yaWxsL3dlYi1sb2NhbFwiO2NvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9maWxlbmFtZSA9IFwiL1VzZXJzL3JveWVuZG8vRGVza3RvcC9HaXRIdWIvcmlsbC93ZWItbG9jYWwvdml0ZS5jb25maWcudHNcIjtjb25zdCBfX3ZpdGVfaW5qZWN0ZWRfb3JpZ2luYWxfaW1wb3J0X21ldGFfdXJsID0gXCJmaWxlOi8vL1VzZXJzL3JveWVuZG8vRGVza3RvcC9HaXRIdWIvcmlsbC93ZWItbG9jYWwvdml0ZS5jb25maWcudHNcIjtpbXBvcnQgeyBzdmVsdGVraXQgfSBmcm9tIFwiQHN2ZWx0ZWpzL2tpdC92aXRlXCI7XG5pbXBvcnQgZG5zIGZyb20gXCJkbnNcIjtcbmltcG9ydCB7IGRlZmluZUNvbmZpZyB9IGZyb20gXCJ2aXRlc3QvY29uZmlnXCI7XG5cbi8vIHByaW50IGRldiBzZXJ2ZXIgYXMgYGxvY2FsaG9zdGAgbm90IGAxMjcuMC4wLjFgXG5kbnMuc2V0RGVmYXVsdFJlc3VsdE9yZGVyKFwidmVyYmF0aW1cIik7XG5cbmNvbnN0IGNvbmZpZyA9IGRlZmluZUNvbmZpZyh7XG4gIGJ1aWxkOiB7XG4gICAgcm9sbHVwT3B0aW9uczoge1xuICAgICAgLy8gVGhpcyBlbnN1cmVzIHRoYXQgdGhlIHdlYi1hZG1pbiBwYWNrYWdlIGlzIG5vdCBidW5kbGVkIGludG8gdGhlIHdlYi1sb2NhbCBwYWNrYWdlLlxuICAgICAgLy8gVGhpcyBpcyBuZWNlc3NhcnkgYmVjYXVzZSB0aGUgU2NoZWR1bGVkIFJlcG9ydHMgZGlhbG9nIGxpdmVzIGluIGB3ZWItY29tbW9uYCBhbmQgaW1wb3J0cyB0aGUgYWRtaW4tY2xpZW50LlxuICAgICAgZXh0ZXJuYWw6IChpZCkgPT4gaWQuc3RhcnRzV2l0aChcIkByaWxsZGF0YS93ZWItYWRtaW4vXCIpLFxuICAgIH0sXG4gIH0sXG4gIHJlc29sdmU6IHtcbiAgICBhbGlhczoge1xuICAgICAgc3JjOiBcIi9zcmNcIiwgLy8gdHJpY2sgdG8gZ2V0IGFic29sdXRlIGltcG9ydHMgdG8gd29ya1xuICAgICAgXCJAcmlsbGRhdGEvd2ViLWxvY2FsXCI6IFwiL3NyY1wiLFxuICAgICAgXCJAcmlsbGRhdGEvd2ViLWNvbW1vblwiOiBcIi8uLi93ZWItY29tbW9uL3NyY1wiLFxuICAgICAgXCJAcmlsbGRhdGEvd2ViLWFkbWluXCI6IFwiLy4uL3dlYi1hZG1pbi9zcmNcIixcbiAgICB9LFxuICB9LFxuICBzZXJ2ZXI6IHtcbiAgICBzdHJpY3RQb3J0OiB0cnVlLFxuICAgIGZzOiB7XG4gICAgICBhbGxvdzogW1wiLlwiXSxcbiAgICB9LFxuICB9LFxuICBkZWZpbmU6IHtcbiAgICBcImltcG9ydC5tZXRhLmVudi5WSVRFX1BMQVlXUklHSFRfVEVTVFwiOiBwcm9jZXNzLmVudi5QTEFZV1JJR0hUX1RFU1QsXG4gICAgXCJpbXBvcnQubWV0YS5lbnYuVklURV9QTEFZV1JJR0hUX0NMT1VEX1RFU1RcIjpcbiAgICAgIHByb2Nlc3MuZW52LlBMQVlXUklHSFRfQ0xPVURfVEVTVCxcbiAgfSxcbiAgcGx1Z2luczogW3N2ZWx0ZWtpdCgpXSxcbiAgZW52RGlyOiBcIi4uL1wiLFxuICBlbnZQcmVmaXg6IFwiUklMTF9VSV9QVUJMSUNfXCIsXG59KTtcblxuZXhwb3J0IGRlZmF1bHQgY29uZmlnO1xuIl0sCiAgIm1hcHBpbmdzIjogIjtBQUFzVCxTQUFTLGlCQUFpQjtBQUNoVixPQUFPLFNBQVM7QUFDaEIsU0FBUyxvQkFBb0I7QUFHN0IsSUFBSSxzQkFBc0IsVUFBVTtBQUVwQyxJQUFNLFNBQVMsYUFBYTtBQUFBLEVBQzFCLE9BQU87QUFBQSxJQUNMLGVBQWU7QUFBQTtBQUFBO0FBQUEsTUFHYixVQUFVLENBQUMsT0FBTyxHQUFHLFdBQVcsc0JBQXNCO0FBQUEsSUFDeEQ7QUFBQSxFQUNGO0FBQUEsRUFDQSxTQUFTO0FBQUEsSUFDUCxPQUFPO0FBQUEsTUFDTCxLQUFLO0FBQUE7QUFBQSxNQUNMLHVCQUF1QjtBQUFBLE1BQ3ZCLHdCQUF3QjtBQUFBLE1BQ3hCLHVCQUF1QjtBQUFBLElBQ3pCO0FBQUEsRUFDRjtBQUFBLEVBQ0EsUUFBUTtBQUFBLElBQ04sWUFBWTtBQUFBLElBQ1osSUFBSTtBQUFBLE1BQ0YsT0FBTyxDQUFDLEdBQUc7QUFBQSxJQUNiO0FBQUEsRUFDRjtBQUFBLEVBQ0EsUUFBUTtBQUFBLElBQ04sd0NBQXdDLFFBQVEsSUFBSTtBQUFBLElBQ3BELDhDQUNFLFFBQVEsSUFBSTtBQUFBLEVBQ2hCO0FBQUEsRUFDQSxTQUFTLENBQUMsVUFBVSxDQUFDO0FBQUEsRUFDckIsUUFBUTtBQUFBLEVBQ1IsV0FBVztBQUNiLENBQUM7QUFFRCxJQUFPLHNCQUFROyIsCiAgIm5hbWVzIjogW10KfQo=
