<script lang="ts">
  export let fontSizeClass = "text-xl";
  export let fontWeightClass = "font-normal";
</script>

<h1 class="{fontSizeClass} {fontWeightClass} text-gray-800">
  <slot />
</h1>
