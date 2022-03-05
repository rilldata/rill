<script>
import { tweened } from "svelte/motion";
import { cubicOut as easing } from "svelte/easing";
export let size = "1em";
export let color='currentColor';

export let mode = 'hamburger';

let TOP = 25
let MIDDLE = 50
let BOTTOM = 75
let LEFT = 30
let RIGHT = 70

let H_LEFT = 15;
let H_RIGHT = 85;

let STROKE_WIDTH = 12;

const defaults = {
    topLeft: TOP,
    topRight: TOP,
    midLeft: MIDDLE,
    midRight: MIDDLE,
    bottomLeft:BOTTOM,
    bottomRight: BOTTOM,

    leftTop: LEFT,
    leftMid: LEFT,
    leftBottom: LEFT,
    rightTop: RIGHT,
    rightMid: RIGHT,
    rightBottom: RIGHT
}

const params = tweened(defaults, { duration: 200, easing, delay: 350 });

$: if (mode === 'hamburger') {
    params.set({
        ...defaults,
        topRight: TOP,
        topLeft: TOP,
        bottomLeft: BOTTOM,
        bottomRight: BOTTOM,

        leftTop: H_LEFT,
        leftMid: H_LEFT,
        leftBottom: H_LEFT,
        rightTop: H_RIGHT,
        rightMid: H_RIGHT,
        rightBottom: H_RIGHT
    });
} else if (mode === 'left') {
    params.set({
        ...defaults,
        topRight: MIDDLE,
        bottomRight: MIDDLE,
        leftMid: RIGHT,
    });
}
else if (mode === 'right') {
    params.set({
        ...defaults,
        topLeft: MIDDLE,
        bottomLeft: MIDDLE,
        rightMid: LEFT
    });
}

</script>


<svg width={size} height={size} viewBox="0 0 100 100">
    <line x1={$params.leftTop} x2={$params.rightTop} y1={$params.topLeft} y2={$params.topRight} stroke={color} stroke-width={STROKE_WIDTH} stroke-linecap=round />
    <line x1={$params.leftMid} x2={$params.rightMid} y1={$params.midLeft} y2={$params.midRight} stroke={color} stroke-width={STROKE_WIDTH} stroke-linecap=round />
    <line x1={$params.leftBottom} x2={$params.rightBottom} y1={$params.bottomLeft} y2={$params.bottomRight} stroke={color} stroke-width={STROKE_WIDTH} stroke-linecap=round />
</svg>