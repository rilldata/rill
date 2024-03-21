// vite.config.ts
import { sveltekit } from "file:///Users/burg/code/rill/web-local/node_modules/@sveltejs/kit/src/exports/vite/index.js";
import dns from "dns";
import { defineConfig } from "file:///Users/burg/code/rill/node_modules/vite/dist/node/index.js";
dns.setDefaultResultOrder("verbatim");
var runtimeUrl = "";
try {
  runtimeUrl = process.env.RILL_DEV ? "http://localhost:9009" : "";
} catch (e) {
  console.error(e);
}
var config = defineConfig(({ mode }) => ({
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
      // Adding $img alias to fix Vite build warnings due to static assets referenced in CSS
      // See: https://stackoverflow.com/questions/75843825/sveltekit-dev-build-and-path-problems-with-static-assets-referenced-in-css
      $img: mode === "production" ? "/../web-common/static/img" : "../img"
    }
  },
  server: {
    port: 3e3,
    strictPort: true,
    fs: {
      allow: ["."]
    }
  },
  define: {
    RILL_RUNTIME_URL: `"${runtimeUrl}"`,
    "import.meta.env.VITE_PLAYWRIGHT_TEST": process.env.PLAYWRIGHT_TEST
  },
  plugins: [sveltekit()]
}));
var vite_config_default = config;
export {
  vite_config_default as default
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsidml0ZS5jb25maWcudHMiXSwKICAic291cmNlc0NvbnRlbnQiOiBbImNvbnN0IF9fdml0ZV9pbmplY3RlZF9vcmlnaW5hbF9kaXJuYW1lID0gXCIvVXNlcnMvYnVyZy9jb2RlL3JpbGwvd2ViLWxvY2FsXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ZpbGVuYW1lID0gXCIvVXNlcnMvYnVyZy9jb2RlL3JpbGwvd2ViLWxvY2FsL3ZpdGUuY29uZmlnLnRzXCI7Y29uc3QgX192aXRlX2luamVjdGVkX29yaWdpbmFsX2ltcG9ydF9tZXRhX3VybCA9IFwiZmlsZTovLy9Vc2Vycy9idXJnL2NvZGUvcmlsbC93ZWItbG9jYWwvdml0ZS5jb25maWcudHNcIjtpbXBvcnQgeyBzdmVsdGVraXQgfSBmcm9tIFwiQHN2ZWx0ZWpzL2tpdC92aXRlXCI7XG5pbXBvcnQgZG5zIGZyb20gXCJkbnNcIjtcbmltcG9ydCB7IGRlZmluZUNvbmZpZyB9IGZyb20gXCJ2aXRlXCI7XG5cbi8vIHByaW50IGRldiBzZXJ2ZXIgYXMgYGxvY2FsaG9zdGAgbm90IGAxMjcuMC4wLjFgXG5kbnMuc2V0RGVmYXVsdFJlc3VsdE9yZGVyKFwidmVyYmF0aW1cIik7XG5cbmxldCBydW50aW1lVXJsID0gXCJcIjtcbnRyeSB7XG4gIHJ1bnRpbWVVcmwgPSBwcm9jZXNzLmVudi5SSUxMX0RFViA/IFwiaHR0cDovL2xvY2FsaG9zdDo5MDA5XCIgOiBcIlwiO1xufSBjYXRjaCAoZSkge1xuICBjb25zb2xlLmVycm9yKGUpO1xufVxuXG5jb25zdCBjb25maWcgPSBkZWZpbmVDb25maWcoKHsgbW9kZSB9KSA9PiAoe1xuICBidWlsZDoge1xuICAgIHJvbGx1cE9wdGlvbnM6IHtcbiAgICAgIC8vIFRoaXMgZW5zdXJlcyB0aGF0IHRoZSB3ZWItYWRtaW4gcGFja2FnZSBpcyBub3QgYnVuZGxlZCBpbnRvIHRoZSB3ZWItbG9jYWwgcGFja2FnZS5cbiAgICAgIC8vIFRoaXMgaXMgbmVjZXNzYXJ5IGJlY2F1c2UgdGhlIFNjaGVkdWxlZCBSZXBvcnRzIGRpYWxvZyBsaXZlcyBpbiBgd2ViLWNvbW1vbmAgYW5kIGltcG9ydHMgdGhlIGFkbWluLWNsaWVudC5cbiAgICAgIGV4dGVybmFsOiAoaWQpID0+IGlkLnN0YXJ0c1dpdGgoXCJAcmlsbGRhdGEvd2ViLWFkbWluL1wiKSxcbiAgICB9LFxuICB9LFxuICByZXNvbHZlOiB7XG4gICAgYWxpYXM6IHtcbiAgICAgIHNyYzogXCIvc3JjXCIsIC8vIHRyaWNrIHRvIGdldCBhYnNvbHV0ZSBpbXBvcnRzIHRvIHdvcmtcbiAgICAgIFwiQHJpbGxkYXRhL3dlYi1sb2NhbFwiOiBcIi9zcmNcIixcbiAgICAgIFwiQHJpbGxkYXRhL3dlYi1jb21tb25cIjogXCIvLi4vd2ViLWNvbW1vbi9zcmNcIixcbiAgICAgIC8vIEFkZGluZyAkaW1nIGFsaWFzIHRvIGZpeCBWaXRlIGJ1aWxkIHdhcm5pbmdzIGR1ZSB0byBzdGF0aWMgYXNzZXRzIHJlZmVyZW5jZWQgaW4gQ1NTXG4gICAgICAvLyBTZWU6IGh0dHBzOi8vc3RhY2tvdmVyZmxvdy5jb20vcXVlc3Rpb25zLzc1ODQzODI1L3N2ZWx0ZWtpdC1kZXYtYnVpbGQtYW5kLXBhdGgtcHJvYmxlbXMtd2l0aC1zdGF0aWMtYXNzZXRzLXJlZmVyZW5jZWQtaW4tY3NzXG4gICAgICAkaW1nOiBtb2RlID09PSBcInByb2R1Y3Rpb25cIiA/IFwiLy4uL3dlYi1jb21tb24vc3RhdGljL2ltZ1wiIDogXCIuLi9pbWdcIixcbiAgICB9LFxuICB9LFxuICBzZXJ2ZXI6IHtcbiAgICBwb3J0OiAzMDAwLFxuICAgIHN0cmljdFBvcnQ6IHRydWUsXG4gICAgZnM6IHtcbiAgICAgIGFsbG93OiBbXCIuXCJdLFxuICAgIH0sXG4gIH0sXG4gIGRlZmluZToge1xuICAgIFJJTExfUlVOVElNRV9VUkw6IGBcIiR7cnVudGltZVVybH1cImAsXG4gICAgXCJpbXBvcnQubWV0YS5lbnYuVklURV9QTEFZV1JJR0hUX1RFU1RcIjogcHJvY2Vzcy5lbnYuUExBWVdSSUdIVF9URVNULFxuICB9LFxuICBwbHVnaW5zOiBbc3ZlbHRla2l0KCldLFxufSkpO1xuXG5leHBvcnQgZGVmYXVsdCBjb25maWc7XG4iXSwKICAibWFwcGluZ3MiOiAiO0FBQStRLFNBQVMsaUJBQWlCO0FBQ3pTLE9BQU8sU0FBUztBQUNoQixTQUFTLG9CQUFvQjtBQUc3QixJQUFJLHNCQUFzQixVQUFVO0FBRXBDLElBQUksYUFBYTtBQUNqQixJQUFJO0FBQ0YsZUFBYSxRQUFRLElBQUksV0FBVywwQkFBMEI7QUFDaEUsU0FBUyxHQUFHO0FBQ1YsVUFBUSxNQUFNLENBQUM7QUFDakI7QUFFQSxJQUFNLFNBQVMsYUFBYSxDQUFDLEVBQUUsS0FBSyxPQUFPO0FBQUEsRUFDekMsT0FBTztBQUFBLElBQ0wsZUFBZTtBQUFBO0FBQUE7QUFBQSxNQUdiLFVBQVUsQ0FBQyxPQUFPLEdBQUcsV0FBVyxzQkFBc0I7QUFBQSxJQUN4RDtBQUFBLEVBQ0Y7QUFBQSxFQUNBLFNBQVM7QUFBQSxJQUNQLE9BQU87QUFBQSxNQUNMLEtBQUs7QUFBQTtBQUFBLE1BQ0wsdUJBQXVCO0FBQUEsTUFDdkIsd0JBQXdCO0FBQUE7QUFBQTtBQUFBLE1BR3hCLE1BQU0sU0FBUyxlQUFlLDhCQUE4QjtBQUFBLElBQzlEO0FBQUEsRUFDRjtBQUFBLEVBQ0EsUUFBUTtBQUFBLElBQ04sTUFBTTtBQUFBLElBQ04sWUFBWTtBQUFBLElBQ1osSUFBSTtBQUFBLE1BQ0YsT0FBTyxDQUFDLEdBQUc7QUFBQSxJQUNiO0FBQUEsRUFDRjtBQUFBLEVBQ0EsUUFBUTtBQUFBLElBQ04sa0JBQWtCLElBQUksVUFBVTtBQUFBLElBQ2hDLHdDQUF3QyxRQUFRLElBQUk7QUFBQSxFQUN0RDtBQUFBLEVBQ0EsU0FBUyxDQUFDLFVBQVUsQ0FBQztBQUN2QixFQUFFO0FBRUYsSUFBTyxzQkFBUTsiLAogICJuYW1lcyI6IFtdCn0K
