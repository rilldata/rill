<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import {
    getLocale,
    setLocale,
  } from "@rilldata/web-common/paraglide/runtime.js";
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";

  const LOCALES = [
    { code: "en", label: () => m.language_en() },
    { code: "es", label: () => m.language_es() },
  ] as const;

  const currentLocale = getLocale();

  type LocaleCode = (typeof LOCALES)[number]["code"];

  function selectLocale(code: LocaleCode) {
    if (code === currentLocale) return;
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
