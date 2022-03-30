<script>
import { getContext } from "svelte";
import transientBooleanStore from "$lib/util/transient-boolean-store";
const callbacks = getContext('rill:app:ui:shift-click-action-callbacks');
let shiftClicked = transientBooleanStore();

// if a parent component upstream triggers the shift-click action,
// let's flip our transientBooleanStore to create the animation.
callbacks.addCallback(() => {
    shiftClicked.flip();
})

</script>

<span class="inline-block shiftable" class:shiftClicked={$shiftClicked}><slot /></span>

<style>

.shiftable {
    padding-left: 2px;
    margin-right: -2px;
    transform: translateY(0px) translateX(-2px);
    transition: transform 200ms;
}

.shiftClicked {
    animation: pulse 250ms;
    border-radius: 2px;
    position: relative;
    mix-blend-mode: screen;
    background-blend-mode: screen;
}

@keyframes pulse {
    0%, 100% {
        transform: translateY(0px) translateX(-2px);
    }
    50% {
        transform: translateY(2px) translateX(2px);
        box-shadow: -1px -1px 0px rgba(100,100,100,1),
                    -2px -2px 0px rgba(75,75,75,1),
                    -3px -3px 0px rgba(50,50,50,1);
    }
}

</style>