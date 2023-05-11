<script lang="ts">
  export let variant: "standard" | "bold" = "standard";

  function getVariantClass(variant: string) {
    switch (variant) {
      case "standard":
        return "text-xl font-normal";
      case "bold":
        return "text-lg font-semibold";
    }
  }
</script>

<h1 class="{getVariantClass(variant)} text-gray-800">
  <slot />
</h1>
