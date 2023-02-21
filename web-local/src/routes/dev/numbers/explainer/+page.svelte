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

  onMount(() => {
    console.time("charMeasuringDiv");
    numFormattingWidthLookupKeys.forEach((str) => {
      charMeasuringDiv.innerHTML = str;
      let rect = charMeasuringDiv.getBoundingClientRect();
      numFormattingWidthLookup[str] = rect.right - rect.left;
    });

    console.timeEnd("charMeasuringDiv");

    pxWidthLookupFn = (str: string) => {
      return str
        .split("")
        .map((char) => numFormattingWidthLookup[char])
        .reduce((a, b) => a + b, 0);
    };
  });
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
      blurb: `Our number representations must be fully general -- positive and negative, huge and tiny, we need to be able to display any valid f64 number, because invariably our users will end up trying to display numbers that upset our expectations of what is reasonable. Indeed, for one of our first real world use cases, Kasper was trying to build a dashboard with currency amounts on the order of 1e-14. As we target TAM expansion, we should expect even more unusual requirements.
      
      <br/><br/>
      In addition to removing a lot of detail, the legacy humanizer also misleadingly renders several numbers (exact 0 as "<0", several negative values as "<0.1k"). Not shown in this example, but the legacy humanizer also struggles with huge numbers and infinitesimals`,
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
      blurb: `One of the principles we've previously tried to hold to for our number approximations is that we should limit the amount of precision that show users. Typically only a few digits of precision should be enough to convey most of the meaningful information in a number, and reducing the number of digits shown can increase the information density of a display, and may in many cases improve the scannability and interpretabilty of a set of numbers.

      <br/><br/>
      
      However, there is of course a trade off: in the general case, imposing a digit limit may require the use of order of magnitude suffixes for number that are large or small. This may make the numbers harder to scan and may impose additional cognitive overhead when attempting to umderstand and interpret a list of numbers.
`,
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
      blurb: `With a digit limit imposed, the question becomes: what is the most effective way to show order of magnitude suffixes?

      <br/><br/>
      There is unfortunately no obviously correct answer. The strategy we have used in the legacy humanizer is to express all of the numbers in a set in terms of the OoM of the largest number in that set. This has advantages in terms of scannability, since all numbers can be directlty compared visually without having to look at a suffix, but it can mean that a lof of detail is lost, especially in the (fairly common) case in which a set of numbers has an outlier, which can sometimes be the least meaningful number in a set of numbers, but which will dominate the interpretation if only one OoM is shown.

      <br/><br/>
      
      Alternatively, we might choose to show several orders of magnitude, appending an order of magnitude suffix whenever required in order to provide a meaningful representation within the digit limit we have set. Within this strategy we have several degrees of freedom. One desirable property within this overall strategy might be to try to show as many numbers as possible without an order of magnitude suffix, for which we might adopt a number of rounding and truncation strategies--in particular we might be willing to truncate  infinitesimals in order to show them withouta suffix.

      

      <br/><br/>
      
      Note that in all cases, I have chosen to use short scale letter suffixes when possible (k, M, B, etc) to follow the convention of the legacy humanizer, and the legacy dashboard.Additionally,when those commonly known letter suffixes are not available, I have chosen to use "engineering style" multiple of three order of magnitude groupings (e.g. e-12, e-9, e21, e24, etc) since these follow the multiple of three order of magnitude convention used by the short scale letter suffixes, and since this convention allows at least a bit of scannability by number shape within these three order of magnitude ranges.


`,
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
      blurb: `Aligning columns of numbers to the decimal point is standard practice in print, and greatly improves scannability.

      <br/><br/>
      However, as as can be seen from the examples above, we once again encounter a trade off around alignment: if we impose a fixed total digit limit for each number, and we prioritize showing numbers without a suffix, then aligning by decimal point can cause a column of numbers to print raggedly, as well as causing the column to take up more horizontal space than it would without alignment.

      <br/><br/>

      These latter issues are really only a problem with the multiple magnitudes strategy; when using the largest magnitude strategy, the numbers are decimal aligned by default.

      <br/><br/>

     The examples below show the improvement in scannability that is allowed when numbers are aligned by decimal point -- at the cost of increased width and ragged/non-existent right alignment, the decimal aligned columns is scannable by shape, whereas the non-aligned columns are more of a block of text. 

<br/><br/>
     
     Additionally, we might or might not choose to align suffixes vertically. That option is shown as well

`,
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

{#if explainerDefs}
  {#each explainerDefs as { title, blurb, formatterCols, sample }}
    <h1>{title}</h1>
    <div style="width: 500px;">{@html blurb}</div>

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
