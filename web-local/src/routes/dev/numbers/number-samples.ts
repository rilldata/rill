import random from "@stdlib/random/base";

const N = 22;
// const rand = random.randu.factory({ seed: 128 });
// for (let i = 0; i < 100; i++) {
//   console.log(rand());
// }

const range = new Array(N).fill(0).map((_, i) => i);

const tDist = random.t.factory({ seed: 1228 });
const randu = random.randu.factory({ seed: 1228 });
const uniform = random.uniform.factory({ seed: 1228 });

const randDiscrete = random.discreteUniform.factory({ seed: 1228 });
// for (let i = 0; i < 100; i++) {
//   console.log(t(1) ** 5);
// }

console.log(range.map((x) => tDist(1) ** 5));

export const numberLists = [
  {
    desc: "pathological for humanizer",
    sample: [
      -0.01277434195, 3.27535562058, -178.59557627756, 1000.2606552, 0,
      0.00000004063831624, -39.47453665617, -33.29703734674, -0.00291292193,
      12.94930626255, -0.07578137641, -22.59459041752, -0, 477.50966127074,
      0.00000000580283174, -0.0000000335513, -154.64489886467, 0,
      -27.2133474649, -0.02432294641, -0.000000000039737053,
    ],
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
    desc: "uniform(0,1)",
    sample: range.map((x) => randu()),
  },

  {
    desc: "in (0,1), ragged, with exact zeros",
    sample: range
      .map((x) => randu() * 1.1 - 0.2)
      .map((x) => (x > 0 ? x : 0))
      .map((x) => +x.toPrecision(randDiscrete(1, 5))),
  },

  {
    desc: "uniform(0,1e12)",
    sample: range.map((x) => randu() * 1e12),
  },

  {
    desc: "uniform(0,1e-14)",
    sample: range.map((x) => randu() * 1e-14),
  },

  {
    desc: "uniform(-300,700)",
    sample: range.map((x) => uniform(-300, 1000)),
  },
  {
    desc: "uniform(-300,700) with O(1e7) outlier",
    sample: range.map((x, i) =>
      i === 7 ? randu() * 1e7 : uniform(-300, 1000)
    ),
  },
];
