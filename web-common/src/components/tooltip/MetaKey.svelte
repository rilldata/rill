<script context="module" lang="ts">
  const modifier = isMac() ? "⌘" : "Ctrl";

  function isMac() {
    return window.navigator.userAgent.includes("Macintosh");
  }
</script>

<script lang="ts">
  export let action = "Click";
  export let withAlt = false;
  export let plusses = true;

  $: finalString = makeString(action, withAlt, plusses);

  function makeString(action: string, withAlt: boolean, withPlusSign: boolean) {
    const elements: string[] = [];

    if (withAlt) {
      elements.push(isMac() ? "⌥" : "alt");
    }

    elements.push(modifier);
    elements.push(action);

    return elements.join(withPlusSign ? " + " : "");
  }
</script>

<span>
  {finalString}
</span>
