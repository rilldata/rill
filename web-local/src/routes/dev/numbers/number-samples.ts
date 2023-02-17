import random from "@stdlib/random/base";
import shuffle from "@stdlib/random/shuffle";
// import { number } from "@stdlib/stdlib/docs/types";

const N = 20;

const range = new Array(N).fill(0).map((_, i) => i);

let tDist = random.t.factory({ seed: 1228 });
let randu = random.randu.factory({ seed: 1228 });
let uniform = random.uniform.factory({ seed: 1228 });

let randDiscrete = random.discreteUniform.factory({ seed: 1228 });
type shuffleFn = (x: number[]) => number[];

let shuffler = shuffle.factory({
  seed: 239,
}) as shuffleFn;

type NumericSampleGen = {
  desc: string;
  sampleFn: () => number[];
};

type NumericSample = {
  desc: string;
  sample: number[];
};

const numberListsGen: NumericSampleGen[] = [
  {
    desc: "pathological for humanizer",
    sampleFn: () =>
      [
        -0.01277434195, 3.27535562058, -178.59557627756, 1000.2606552, 0,
        0.00000004063831624, -39.47453665617, -33.29703734674, -0.00291292193,
        12.94930626255, -0.07578137641, -22.59459041752, -0, 477.50966127074,
        0.00000000580283174, -0.0000000335513, -154.64489886467, 0,
        -27.2133474649, -0.02432294641, -0.000000000039737053,
      ].slice(0, N),
  },

  {
    desc: "magnitudes (e-15, e-12) (Kasper's case)",
    sampleFn: () => range.map((x) => 10 ** uniform(-15, -12)),
  },

  {
    desc: "uniform (-1000, 1000)",
    sampleFn: () => range.map((x) => uniform(-1000, 1000)),
  },
  {
    desc: "uniform (-300,700) with O(1e7) outlier",
    sampleFn: () =>
      range.map((x, i) => (i === 7 ? randu() * 1e7 : uniform(-300, 1000))),
  },

  {
    desc: "in (0,1), ragged, with exact zeros",
    sampleFn: () =>
      range
        .map(() => randu() * 1.1 - 0.2)
        .map((x) => (x > 0 ? x : 0))
        .map((x) => +x.toPrecision(randDiscrete(1, 5))),
  },

  {
    desc: "power law-ish (uniform over magnitudes (e-15, e12))",
    sampleFn: () => range.map((x) => 10 ** uniform(-15, 12)),
  },

  {
    desc: "power law-ish (uniform over magnitudes (e-6, e13))",
    sampleFn: () =>
      range
        .map((x) => 10 ** uniform(-6, 13))
        .map((x) => (randu() < 0.2 ? 0 : x)),
  },

  {
    desc: "-(t-dist)^8 (all neg, with exact zeros)",
    sampleFn: () =>
      range.map((x) => -(tDist(1) ** 8)).map((x) => (randu() < 0.1 ? 0 : x)),
  },

  {
    desc: "t-dist to the 5th",
    sampleFn: () => range.map((x) => tDist(1) ** 7),
  },

  {
    desc: "t-dist to the 5th, 2 digits precision",
    sampleFn: () => range.map((x) => +(tDist(1) ** 5).toPrecision(2)),
  },
  {
    desc: "orders of mag e-5 to e5, 2 digits precision, some exact zeros",
    sampleFn: () =>
      range
        .map((x) => -(10 ** uniform(-5, 5)))
        .map((x) => (randu() < 0.3 ? 0 : x))
        .map((x) => x * (randu() < 0.5 ? -1 : 1))
        .map((x) => +x.toPrecision(2)),
  },
  {
    desc: "orders of mag e0 to e5, rounded to ints",
    sampleFn: () =>
      range
        .map((x) => 10 ** uniform(0, 5))
        .map((x) => x * (randu() < 0.5 ? -1 : 1))
        .map(Math.round),
  },

  {
    desc: "t-dist cubed, rounded to int",
    sampleFn: () => range.map((x) => Math.round(tDist(1) ** 3)),
  },
  {
    desc: "all negative, power law-ish, zero inflated",
    sampleFn: () =>
      range
        .map((x) => -(10 ** uniform(-3, 6)))
        .map((x) => (randu() < 0.3 ? 0 : x)),
  },

  {
    desc: "pos & neg, power law-ish, zero inflated 2",
    sampleFn: () =>
      shuffler(
        range.map((i) => (i % 2 === 0 ? -1 : 1) * randu() * 10 ** (i * 0.6 - 3))
      ),
    // .map((x) => (randu() < 0.3 ? 0 : x)),
  },

  {
    desc: "uniform (0,1)",
    sampleFn: () => range.map((x) => randu()),
  },

  {
    desc: "uniform(0,1e12)",
    sampleFn: () => range.map((x) => randu() * 1e12),
  },

  {
    desc: "power law-ish (uniform over magnitudes (e-200, e200))",
    sampleFn: () => range.map((x) => 10 ** uniform(-200, 200)),
  },
];

export const numberLists: NumericSample[] = numberListsGen.map((g) => {
  // reset all seeds
  tDist = random.t.factory({ seed: 1228 });
  randu = random.randu.factory({ seed: 1228 });
  uniform = random.uniform.factory({ seed: 1228 });
  randDiscrete = random.discreteUniform.factory({ seed: 1228 });
  shuffler = shuffle.factory({
    seed: 239,
  }) as shuffleFn;

  return { desc: g.desc, sample: g.sampleFn() };
});
