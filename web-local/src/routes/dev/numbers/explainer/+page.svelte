<script lang="ts">
  import TableExampleWidget from "../table-example-widget.svelte";
  import SpanMeasurer from "../span-measurer.svelte";
  import FormattersInColums from "./formatters-in-colums.svelte";
  import { numberLists as numberListsUnprocessed } from "../number-samples";

  import {
    formatterFactories,
    NumberFormatter,
    NumPartPxWidthLookupFn,
    RichFormatNumber,
  } from "../number-to-string-formatters";
  import type { FormatterOptionsV1 } from "../formatter-options";
  import { onMount } from "svelte";
  // ======== pxWidthLookupFn machinery =========
  let pxWidthLookupFn: NumPartPxWidthLookupFn;
  let numFormattingWidthLookupKeys = [
    ".",
    "-",
    "$",
    "%",
    "k",
    "M",
    "B",
    "T",
    "Q",
    "e",
    "E",
  ];
  for (let i = 0; i <= 9; i++) {
    numFormattingWidthLookupKeys.push(i + "");
  }
  let numFormattingWidthLookup: { [key: string]: number } = {};
  let charMeasuringDiv: HTMLDivElement;

  const setUpPxWidthLookupFn = () => {
    console.time("setUpPxWidthLookupFn");

    numFormattingWidthLookupKeys.forEach((str) => {
      charMeasuringDiv.innerHTML = str;
      let rect = charMeasuringDiv.getBoundingClientRect();
      numFormattingWidthLookup[str] = rect.right - rect.left;
    });

    console.timeEnd("setUpPxWidthLookupFn");

    pxWidthLookupFn = (str: string) => {
      return str
        .split("")
        .map((char) => numFormattingWidthLookup[char])
        .reduce((a, b) => a + b, 0);
    };
  };

  onMount(() => {
    // when fonts are done loading,measure the character sizes
    if (document.fonts.check("12px Inter")) {
      setUpPxWidthLookupFn();
    } else {
      document.fonts.onloadingdone = setUpPxWidthLookupFn;
    }
  });
  // onMount(() => {
  //   console.time("charMeasuringDiv");
  //   numFormattingWidthLookupKeys.forEach((str) => {
  //     charMeasuringDiv.innerHTML = str;
  //     let rect = charMeasuringDiv.getBoundingClientRect();
  //     numFormattingWidthLookup[str] = rect.right - rect.left;
  //   });

  //   console.timeEnd("charMeasuringDiv");

  //   pxWidthLookupFn = (str: string) => {
  //     return str
  //       .split("")
  //       .map((char) => numFormattingWidthLookup[char])
  //       .reduce((a, b) => a + b, 0);
  //   };
  // });
  // END ======== pxWidthLookupFn machinery =========

  type FormatterColumnRecipe = [
    formatterName: string,
    colHeading: string,
    options: Partial<FormatterOptionsV1>
  ];

  type ExplainerStoryRecipe = {
    title: string;
    sampleName?: string;
    blurb?: string;
    formatterColRecipes?: FormatterColumnRecipe[];
  };

  type FormatterDescriptionAndOptions = [
    formatter: NumberFormatter,
    description: string,
    options: FormatterOptionsV1,
    pxWidth: number
  ];

  type ExplainerStoryOut = {
    title: string;
    blurb: string;
    formatterCols: FormatterDescriptionAndOptions[];
    sample: number[];
  };

  // let formatterRecipes: [string, string, Partial<FormatterOptionsV1>][] = [
  //   ["JS `toString()`", "full raw numbers", {}],
  //   ["new humanizer", "humanized", {}],
  // ];

  let formattersDescriptionsAndOptions: [
    formatter: NumberFormatter,
    description: string,
    options: FormatterOptionsV1,
    pxWidth: number
  ][];

  let baseOptions: FormatterOptionsV1 = {
    magnitudeStrategy: "unlimitedDigitTarget",
    digitTarget: 5,
    digitTargetPadWithInsignificantZeros: false,
    usePlainNumsForThousands: false,
    usePlainNumsForThousandsOneDecimal: false,
    usePlainNumForThousandths: false,
    usePlainNumForThousandthsPadZeros: false,
    truncateThousandths: false,
    truncateTinyOrdersIfBigOrderExists: false,
    zeroHandling: "exactZero",
    maxTotalDigits: 6,
    maxDigitsLeft: 3,
    maxDigitsRight: 3,
    minDigitsNonzero: 1,
    nonIntegerHandling: "trailingDot",
    formattingUnits: "none",
    specialDecimalHandling: "noSpecial",
    alignDecimalPoints: true,
    alignSuffixes: true,
    suffixPadding: 2,
    lowerCaseEForEng: true,
    showMagSuffixForZero: false,
  };

  let explainerRecipes: ExplainerStoryRecipe[] = [
    {
      title: "Generality",

      sampleName: "pathological for humanizer",
      formatterColRecipes: [
        ["JS `toString()`", "raw-ish numbers (JS `toString()`)", {}],
        ["humanizeGroupValues (current humanizer)", "legacy humanizer", {}],
        ["new humanizer", "new humanizer, multiple magnitudes", {}],
        [
          "new humanizer",
          "new humanizer, largest magnitude",
          { magnitudeStrategy: "largestWithDigitTarget" },
        ],
      ],
    },
    {
      title: "Digit limit vs. need for suffixes",

      sampleName: "power law-ish (uniform over magnitudes (e-15, e12))",
      formatterColRecipes: [
        ["JS `toString()`", "raw-ish numbers (JS `toString()`)", {}],
        ["humanizeGroupValues (current humanizer)", "legacy humanizer", {}],
        ["new humanizer", "new humanizer, multiple magnitudes", {}],
        [
          "new humanizer",
          "new humanizer, largest magnitude",
          { magnitudeStrategy: "largestWithDigitTarget" },
        ],
      ],
    },

    {
      title: "Order of magnitude suffix strategies",

      sampleName: "pos & neg, power law-ish",
      formatterColRecipes: [
        ["JS `toString()`", "raw-ish numbers (JS `toString()`)", {}],
        ["humanizeGroupValues (current humanizer)", "legacy humanizer", {}],
        [
          "new humanizer",
          "new humanizer, largest magnitude",
          { magnitudeStrategy: "largestWithDigitTarget" },
        ],
        [
          "new humanizer",
          "multiple magnitudes (always show suffix, except e0)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 3,
            maxDigitsRight: 3,
            minDigitsNonzero: 3,
          },
        ],
        [
          "new humanizer",
          "multiple magnitudes (try to show as e0 if at least one digit of precision remains)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 3,
            maxDigitsRight: 5,
            minDigitsNonzero: 1,
          },
        ],
        [
          "new humanizer",
          "multiple magnitudes (allow truncation of infintesimals)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 3,
            maxDigitsRight: 5,
            minDigitsNonzero: 0,
          },
        ],
        [
          "new humanizer",
          "multiple magnitudes (truncate infintesimals)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 5,
            maxDigitsRight: 5,
            minDigitsNonzero: 0,
          },
        ],
      ],
    },

    {
      title: "Decimal alignment",

      sampleName: "power law-ish (uniform over magnitudes (e1, e8))",
      formatterColRecipes: [
        ["JS `toString()`", "raw-ish numbers (JS `toString()`)", {}],
        [
          "new humanizer",
          "new humanizer, largest magnitude",
          { magnitudeStrategy: "largestWithDigitTarget" },
        ],
        [
          "new humanizer",
          "multiple magnitudes (aligned)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 5,
            maxDigitsRight: 5,
            minDigitsNonzero: 0,

            alignDecimalPoints: true,
            alignSuffixes: true,
          },
        ],

        [
          "new humanizer",
          "multiple magnitudes (not aligned)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 5,
            maxDigitsRight: 5,
            minDigitsNonzero: 0,

            alignDecimalPoints: false,
            alignSuffixes: true,
          },
        ],

        [
          "new humanizer",
          "multiple magnitudes (suffixes not aligned)",
          {
            maxTotalDigits: 6,
            maxDigitsLeft: 5,
            maxDigitsRight: 5,
            minDigitsNonzero: 0,

            alignDecimalPoints: false,
            alignSuffixes: false,
          },
        ],
      ],
    },

    { title: "Indications of approximation" },
  ];

  let explainerDefs: ExplainerStoryOut[];

  $: {
    if (pxWidthLookupFn !== undefined) {
      // console.log({ explainerRecipes });
      explainerDefs = explainerRecipes
        .filter((recipe) => recipe.formatterColRecipes && recipe.sampleName)
        .map((recipe) => {
          console.log({ recipe });
          const sample = numberListsUnprocessed.find(
            (nl) => nl.desc === recipe.sampleName
          ).sample;

          // console.log({ formatterRecipes });

          formattersDescriptionsAndOptions = recipe.formatterColRecipes.map(
            ([ffName, colHeader, options]) => {
              const finalOptions: FormatterOptionsV1 = {
                ...baseOptions,
                ...options,
              };
              const formatterFactory: NumberFormatter = formatterFactories
                .find((ff) => ff.desc === ffName)
                .fn(sample, pxWidthLookupFn, finalOptions);

              const maxPxWidths = sample
                .map(formatterFactory)
                .map((rn) => rn.maxPxWidth)
                .reduce(
                  (a, b) => ({
                    int: Math.max(a.int, b.int),
                    dot: Math.max(a.dot, b.dot),
                    frac: Math.max(a.frac, b.frac),
                    suffix: Math.max(a.suffix, b.suffix),
                  }),
                  { int: 0, dot: 0, frac: 0, suffix: 0 }
                );

              const totalPxWidth = Object.values(maxPxWidths).reduce(
                (a, b) => a + b,
                0
              );

              return [
                formatterFactory,
                colHeader,
                finalOptions,
                totalPxWidth + finalOptions.suffixPadding,
              ];
            }
          );

          return {
            title: recipe.title,
            blurb: recipe.blurb,
            formatterCols: formattersDescriptionsAndOptions,
            sample,
          };
        });

      console.log({ explainerDefs });
    }
  }
  let tableGutterWidth = 30;
</script>

<div class="outer">
  <div class="inner ui-copy-number" bind:this={charMeasuringDiv}>CONTENT</div>
</div>

<h1 style="font-size: 20px;">
  Visit <a
    href="https://www.notion.so/rilldata/humanizer-v2-explainer-ecfa5daf565644d3ad7a95ac464d0972"
    >Notion page</a
  > for description and discussion
</h1>

<br />

{#if explainerDefs}
  {#each explainerDefs as { title, blurb, formatterCols, sample }}
    <h1>{title}</h1>

    {#if formatterCols && sample}
      <FormattersInColums
        formattersDescriptionsAndOptions={formatterCols}
        {sample}
        {tableGutterWidth}
      />
    {/if}
  {/each}
{/if}

<style>
  .outer {
    overflow: hidden;
    position: relative;
  }
  .inner {
    position: absolute;
    right: -50px;
    top: 50px;
    width: fit-content;
  }
</style>
