<!-- Shows a persistent banner when browsing as another user (assumed state) -->
<script lang="ts">
  import { onMount } from "svelte";
  import { browser } from "$app/environment";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import { STORAGE_KEY } from "@rilldata/web-admin/features/superuser/users/assume-state";
  import { ADMIN_URL } from "@rilldata/web-admin/client/http-client";

  const BANNER_ID = "representing-user";

  function unassume() {
    if (browser) localStorage.removeItem(STORAGE_KEY);
    eventBus.emit("remove-banner", BANNER_ID);
    // Redirect to login; the auth provider session is the real superuser,
    // so it auto-completes and issues a fresh superuser token.
    const u = new URL("auth/login", ADMIN_URL);
    u.searchParams.set("redirect", window.location.origin);
    window.location.href = u.toString();
  }

  function showBanner(email: string) {
    eventBus.emit("add-banner", {
      id: BANNER_ID,
      priority: 0,
      message: {
        type: "warning",
        message: `Browsing as <strong>${email}</strong>`,
        includesHtml: true,
        iconType: "alert",
        cta: {
          text: "Unassume",
          type: "button",
          onClick: unassume,
        },
      },
    });
  }

  onMount(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      showBanner(stored);
    }

    // Sync banner across tabs: localStorage fires "storage" events in
    // other tabs when the value changes, so every tab shows/hides the
    // banner when a superuser assumes or unassumes in any tab.
    function onStorage(e: StorageEvent) {
      if (e.key !== STORAGE_KEY) return;
      if (e.newValue) {
        showBanner(e.newValue);
      } else {
        eventBus.emit("remove-banner", BANNER_ID);
      }
    }

    window.addEventListener("storage", onStorage);
    return () => window.removeEventListener("storage", onStorage);
  });
</script>
