<script>
    import Base from "./Base.svelte";
    export let isNull = false;
    export let inTable = false;
    export const textAlign='text-right';
</script>
<Base {isNull} classes="italic font-semibold text-slate-500 {inTable && 'text-left'}">
    <slot />
</Base>