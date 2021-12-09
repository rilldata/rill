<script>
export let height = "calc(100vh - 2rem)"
</script>

<div style="height:{height};">
    <slot />
</div>

<style>
div {
    overflow-y: auto;
    overflow-x: hidden;
}
</style>