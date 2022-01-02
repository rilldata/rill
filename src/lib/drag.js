export function drag(node, params) {
    let minSize_ = params?.minSize || 300;
    let maxSize_ = params?.maxSize || 800;
    
    let side_ = params?.side || 'right';
    let property = `--${side_}-sidebar-width`; //params?.property || '--left-sidebar-width';
    let moving = false;
    let xSpace = minSize_;

    node.style.cursor = "move";
    node.style.userSelect = "none";

    function mousedown() {
        moving = true;
    }

    function mousemove(e) {
        if (moving) {
        const size = side_ === 'right' ? innerWidth - e.pageX : e.pageX;
        if (size > minSize_ && size < maxSize_) {
            xSpace = size;
        }

        document.body.style.setProperty(property, `${xSpace}px`)
        }
    }

    function mouseup() {
        moving = false;
    }

    node.addEventListener("mousedown", mousedown);
    window.addEventListener("mousemove", mousemove);
    window.addEventListener("mouseup", mouseup);
    return {
        update() {
        moving = false;
        },
    };
}
