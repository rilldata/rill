<!-- Shows a persistent banner when browsing as another user (assumed state).
     Subscribes to the storage event so every tab shows or hides the banner
     when any tab assumes or unassumes. -->
<script lang="ts">
  import { onMount } from "svelte";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
  import {
    STORAGE_KEY,
    assumedUser,
  } from "@rilldata/web-admin/features/superuser/users/assume-state";

  const BANNER_ID = "representing-user";

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
          onClick: () => assumedUser.unassume(),
        },
      },
    });
  }

  onMount(() => {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored) {
      showBanner(stored);
    }

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
