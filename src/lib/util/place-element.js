/**
 * float-element
 */

// alignment: left, right, bottom, top
// location: bottom top, left, right

function minmax(v, min, max) {
    return Math.max(min, Math.min(v, max));
  }
  
  export function mouseLocationToBoundingRect({ x, y, width = 0, height = 0 }) {
    return {
      parentPosition: {
        width,
        height,
        left: x,
        right: x + width,
        top: y,
        bottom: y + height,
      },
      elementPosition: {
        width,
        height,
        left: x,
        right: x + width,
        top: y,
        bottom: y + height,
      },
    };
  }
  
  export function placeElement({
    location,
    alignment,
    parentPosition, // using getBoundingClientRect // DOMRect
    elementPosition, // using getBoundingClientRect // DOMRect
    distance = 0,
    y = 0,
    windowWidth = window.innerWidth,
    windowHeight = window.innerHeight,
    pad = 16 * 2,
  }) {
    let left;
    let top;
  
    const elementWidth = elementPosition.width;
    const elementHeight = elementPosition.height;
  
    const parentRight = parentPosition.right;
    const parentLeft = parentPosition.left;
    const parentTop = parentPosition.top + y;
    const parentBottom = parentPosition.bottom + y;
    const parentWidth = parentPosition.width;
    const parentHeight = parentPosition.height;
  
    if (location === "bottom") {
      top = parentBottom + distance;
    } else if (location === "top") {
      top = parentTop - elementHeight - distance;
    } else if (location === "left") {
      // FIXME: is this the left / right default?
      left = parentLeft - elementWidth - distance;
    } else {
      left = parentRight + distance;
    }
    // FIXME: throw warning when location & alignment don't make sense
    if (alignment === "right") {
      left = parentRight - elementWidth;
      // set right if off window
      if (left < 0) {
        left = parentLeft;
      }
    } else if (alignment === "left") {
      // make it alignment="right" if it exceeds windowWith - elementWidth.
      left = parentLeft;
      if (left > windowWidth - elementWidth) {
        left = parentRight - elementWidth;
      }
    } else if (alignment === "top") {
      top = parentTop;
      // if bottom edge of float is below height
      if (top + elementHeight > windowHeight - pad) {
        // top = parentBottom - elementHeight;
        top = Math.max(pad, windowHeight - elementHeight - pad);
      }
    } else if (alignment === "bottom") {
      top = parentBottom - elementHeight;
      if (top < 0) {
        top = parentTop;
      }
    } else if (location === "left" || location === "right") {
      // align center + location is left or right
      // do something here
      top = minmax(
        parentTop - (elementHeight - parentHeight) / 2,
        distance,
        windowHeight - distance
      );
    } else {
      // align center + location is top or bottom
      // location is top or bottom
      left = minmax(
        parentLeft - (elementWidth - parentWidth) / 2,
        distance,
        windowWidth - distance
      );
    }
    return [left, top];
  }