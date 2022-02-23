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
  x = 0,
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

  // OUR FIRST JOB IS REFLECTION ALONG THE LOCATION AXIS.
  // If the location of the element would have it drawn off the screen,
  // we will need to reflect it.

  if (location === "bottom") {

    if (parentBottom + elementHeight + distance + pad > windowHeight) {
      top = parentTop - elementHeight - distance;
    } else {
      top = parentBottom + distance;
    }

    
  } else if (location === "top") {
    if ((parentTop) - elementHeight - distance - pad < 0) {
      top = parentBottom + distance;
    } else {
      top = parentTop - elementHeight - distance;
    }
  } else if (location === "left") {
    // FIXME: is this the left / right default?

    if (parentLeft - distance - elementWidth - pad < 0) {
      // reflect
      left = parentRight + distance;
    } else {
      left = parentLeft - elementWidth - distance;
    }

  } else if (location === 'right') {
    // reflect the left to the other side if this won't work.
    if (parentRight + x + distance + elementWidth + pad > windowWidth) {
      // reflect
      left = parentLeft - elementWidth - distance + x;
    } else {
      left = parentRight + x + distance;
    }
    
  }

  // OUR SECOND JOB IS RE-ALIGNMENT ALONG THE ALIGNMENT ACTION.
  // if the alignment causes the elemnet to go off the page,
  // we'll nudge it.
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
    if (top + elementHeight > windowHeight + y - pad) {
      // top = parentBottom - elementHeight;
      top = Math.max(pad, windowHeight + y - elementHeight - pad);
    }
  } else if (alignment === "bottom") {
    top = parentBottom - elementHeight;
    if (top < 0) {
      top = parentTop;
    }
  } else if (location === "left" || location === "right") {
    // TOP-BOTTOM alignment with left / right location.
    top = minmax(
      parentTop - (elementHeight - parentHeight) / 2,
      distance + scrollY,
      windowHeight + scrollY - distance - elementHeight
    );
  } else {
    // align center + location is top or bottom
    // location is top or bottom
    left = minmax(
      parentLeft - (elementWidth - parentWidth) / 2,
      pad,
      windowWidth - elementWidth - pad
    );
  }

  // I think we want some very controlled stuff here.

  return [left, top];
}