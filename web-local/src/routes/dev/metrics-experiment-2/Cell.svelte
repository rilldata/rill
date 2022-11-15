<script lang="ts">
  import { level as levelTokens } from "@rilldata/web-local/lib/application-config";
  export let grow = false;
  export let sticky = false;
  export let level: "info" | "warning" = undefined;

  let bg;
  let text;
  let border;
  $: if (level && level in levelTokens) {
    let theseLevels = levelTokens[level];
    bg = theseLevels.bgSecondary;
    text = theseLevels.text;
    border = theseLevels.borderSecondary;
    console.log(bg, text, border);
  } else {
    bg = undefined;
    text = undefined;
    border = undefined;
  }
</script>

<div
  class:sticky
  class:left-0={sticky}
  class:grow
  class="border flex items-center {bg} {text} {border}"
>
  <slot />
</div>
