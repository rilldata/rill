<script>
    import Base from "./Base.svelte";
    export let isNull = false;
    export let inTable = false;
</script>
<Base {isNull} classes={inTable && 'text-left'}>
    <slot />
</Base>

