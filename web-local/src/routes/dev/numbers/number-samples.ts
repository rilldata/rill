import random from "@stdlib/random/base";
import shuffle from "@stdlib/random/shuffle";
// import { number } from "@stdlib/stdlib/docs/types";

const N = 20;

const range = new Array(N).fill(0).map((_, i) => i);

const tDist = random.t.factory({ seed: 1228 });
const randu = random.randu.factory({ seed: 1228 });
const uniform = random.uniform.factory({ seed: 1228 });

const randDiscrete = random.discreteUniform.factory({ seed: 1228 });
type shuffleFn = (x: number[]) => number[];

const shuffler = shuffle.factory({
  seed: 239,
}) as shuffleFn;

type NumericSample = {
  desc: string;
  sample: number[];
};

export const numberLists: NumericSample[] = [
  {
    desc: "pathological for humanizer",
    sample: [
      -0.01277434195, 3.27535562058, -178.59557627756, 1000.2606552, 0,
      0.00000004063831624, -39.47453665617, -33.29703734674, -0.00291292193,
      12.94930626255, -0.07578137641, -22.59459041752, -0, 477.50966127074,
      0.00000000580283174, -0.0000000335513, -154.64489886467, 0,
      -27.2133474649, -0.02432294641, -0.000000000039737053,
    ].slice(0, N),
  },
  {
    desc: "t dist",
    sample: range.map((x) => tDist(1)),
  },

  {
    desc: "t-dist to the 5th",
    sample: range.map((x) => tDist(1) ** 7),
  },

  {
    desc: "(t-dist)^8 (all positive)",
    sample: range.map((x) => tDist(1) ** 8),
  },

  {
    desc: "t-dist to the 5th, 2 digits precision",
    sample: range.map((x) => +(tDist(1) ** 5).toPrecision(2)),
  },
  {
    desc: "all negative, power law-ish, zero inflated",
    sample: range
      .map((x) => -(10 ** uniform(-3, 6)))
      .map((x) => (randu() < 0.3 ? 0 : x)),
  },

  {
    desc: "pos & neg, power law-ish, zero inflated",
    sample: range
      .map((x) => (randu() < 0.4 ? -1 : 1) * 10 ** uniform(-3, 6))
      .map((x) => (randu() < 0.3 ? 0 : x)),
  },

  {
    desc: "pos & neg, power law-ish, zero inflated 2",
    sample: shuffler(
      range.map((i) => (i % 2 === 0 ? -1 : 1) * randu() * 10 ** (i * 0.6 - 3))
    ),
    // .map((x) => (randu() < 0.3 ? 0 : x)),
  },

  {
    desc: "uniform (0,1)",
    sample: range.map((x) => randu()),
  },

  {
    desc: "in (0,1), ragged, with exact zeros",
    sample: range
      .map(() => randu() * 1.1 - 0.2)
      .map((x) => (x > 0 ? x : 0))
      .map((x) => +x.toPrecision(randDiscrete(1, 5))),
  },

  {
    desc: "uniform(0,1e12)",
    sample: range.map((x) => randu() * 1e12),
  },

  {
    desc: "power law-ish (uniform over magnitudes (e-15, e12))",
    sample: range.map((x) => 10 ** uniform(-15, 12)),
  },

  {
    desc: "uniform over magnitudes (e-15, e-12)",
    sample: range.map((x) => 10 ** uniform(-15, -12)),
  },

  {
    desc: "uniform (-1000, 1000)",
    sample: range.map((x) => uniform(-1000, 1000)),
  },
  {
    desc: "uniform (-300,700) with O(1e7) outlier",
    sample: range.map((x, i) =>
      i === 7 ? randu() * 1e7 : uniform(-300, 1000)
    ),
  },
];
