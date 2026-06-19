<script lang="ts">
  import * as DropdownMenu from "@rilldata/web-common/components/dropdown-menu";
  import {
    getLocale,
    locales,
    setLocale,
    type Locale,
  } from "@rilldata/web-common/lib/i18n/gen/runtime";

  const localeLabels: Record<Locale, string> = {
    en: "English",
    de: "Deutsch",
  };

  // setLocale reloads the page, so reading once on mount is sufficient.
  const currentLocale = getLocale();
</script>

<DropdownMenu.Sub>
  <DropdownMenu.SubTrigger>Language</DropdownMenu.SubTrigger>
  <DropdownMenu.SubContent>
    {#each locales as locale (locale)}
      <DropdownMenu.CheckboxItem
        checkRight
        checked={locale === currentLocale}
        onclick={() => setLocale(locale)}
      >
        {localeLabels[locale] ?? locale}
      </DropdownMenu.CheckboxItem>
    {/each}
  </DropdownMenu.SubContent>
</DropdownMenu.Sub>
