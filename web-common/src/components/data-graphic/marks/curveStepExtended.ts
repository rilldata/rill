function Step(context, t) {
  this._context = context;
  this._t = t;
}

Step.prototype = {
  areaStart: function () {
    this._line = 0;
  },
  areaEnd: function () {
    this._line = NaN;
  },
  lineStart: function () {
    this._x = this._y = NaN;
    this._point = 0;
  },
  lineEnd: function () {
    if (0 < this._t && this._t < 1 && this._point === 2) {
      const extrapolationPoint = this._x + this.diff;
      const newEnd = this._x * (1 - this._t) + extrapolationPoint * this._t;
      this._context.lineTo(newEnd, this._y);
    }
    if (this._line || (this._line !== 0 && this._point === 1)) {
      this._context.closePath();
    }
    if (this._line >= 0) {
      this._t = 1 - this._t;
      this._line = 1 - this._line;
    }
  },
  point: function (x, y) {
    (x = +x), (y = +y);
    switch (this._point) {
      case 0:
        this._point = 1;
        this._line ? this._context.lineTo(x, y) : this._context.moveTo(x, y);
        break;
      case 1:
        this._point = 2; // falls through
      default: {
        if (this._t <= 0) {
          this._context.lineTo(this._x, y);
          this._context.lineTo(x, y);
        } else {
          const x1 = this._x * (1 - this._t) + x * this._t;
          this.diff = Math.abs(this._x - x);
          this._context.lineTo(x1, this._y);
          this._context.lineTo(x1, y);
        }
        break;
      }
    }
    (this._x = x), (this._y = y);
  },
};

export function curveStepExtended(context) {
  return new Step(context, 0.5);
}
