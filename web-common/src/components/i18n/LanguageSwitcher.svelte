<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import {
    getLocale,
    setLocale,
  } from "@rilldata/web-common/paraglide/runtime.js";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";

  export let persistLocale: ((code: string) => Promise<void>) | undefined =
    undefined;

  const LOCALES = [
    { code: "en", label: () => m.language_en() },
    { code: "es", label: () => m.language_es() },
  ] as const;

  const currentLocale = getLocale();

  type LocaleCode = (typeof LOCALES)[number]["code"];

  async function selectLocale(code: LocaleCode) {
    if (code === currentLocale) return;

    if (persistLocale) {
      try {
        await persistLocale(code);
      } catch (e) {
        console.error("Failed to persist language preference", e);
        eventBus.emit("notification", {
          message: m.language_switcher_persist_error(),
          type: "error",
        });
        return;
      }
    }

    setLocale(code);
  }
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger
    >{m.language_switcher_label()}</DropdownMenu.SubTrigger
  >
  <DropdownMenu.SubContent>
    {#each LOCALES as loc}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={currentLocale === loc.code}
        onclick={() => selectLocale(loc.code)}
      >
        {loc.label()}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>
