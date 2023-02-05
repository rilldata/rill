import type { Reference } from ".";
interface ReferenceTestCase {
  query: string;
  references: Reference[];
  embeddedReferences: Reference[];
}

export const singleFrom: ReferenceTestCase = {
  query: `select * FROM tbl`,
  references: [
    {
      reference: "tbl",
      type: "from",
      index: 9,
      referenceIndex: 14,
    },
  ],
  embeddedReferences: [],
};

export const bigGapInFrom: ReferenceTestCase = {
  query: `select * FROM 




tbl`,
  references: [
    {
      reference: "tbl",
      type: "from",
      index: 9,
      referenceIndex: 19,
    },
  ],
  embeddedReferences: [],
};

/** subqueries don't count as a reference, exactly, but the FROM statement
 * in a subquery does.
 */
export const subqueryFrom: ReferenceTestCase = {
  query: `select * from (select * from tbl)`,
  references: [
    {
      reference: "tbl",
      type: "from",
      index: 24,
      referenceIndex: 29,
    },
  ],
  embeddedReferences: [],
};

/** unfinished queries shouldn't have any references */
export const unfinishedFrom: ReferenceTestCase = {
  query: `select * from `,
  references: [],
  embeddedReferences: [],
};

/** simple join. Also tests line breaks. */
export const simpleJoin: ReferenceTestCase = {
  query: `select * from tbl JOIN 
  
  x ON tbl.id = x.id`,
  references: [
    {
      reference: "tbl",
      type: "from",
      index: 9,
      referenceIndex: 14,
    },
    {
      reference: "x",
      type: "join",
      index: 18,
      referenceIndex: 29,
    },
  ],
  embeddedReferences: [],
};

export const remoteSourceFrom: ReferenceTestCase = {
  query: `FROM "s3://path/to/bucket.parquet"`,
  references: [
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "from",
      index: 0,
      referenceIndex: 5,
    },
  ],
  embeddedReferences: [
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "from",
      index: 0,
      referenceIndex: 5,
    },
  ],
};

/** do not capture aliases (for now) */
export const remoteSourceFromAlias: ReferenceTestCase = {
  query: `FROM "s3://path/to/bucket.parquet" as tbl`,
  references: [
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "from",
      index: 0,
      referenceIndex: 5,
    },
  ],
  embeddedReferences: [
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "from",
      index: 0,
      referenceIndex: 5,
    },
  ],
};

export const remoteSourceJoin: ReferenceTestCase = {
  query: `   FROM tbl JOIN "s3://path/to/bucket.parquet" as tbl2 ON tbl2.id = tbl.id`,
  references: [
    {
      reference: "tbl",
      type: "from",
      index: 3,
      referenceIndex: 8,
    },
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "join",
      index: 12,
      referenceIndex: 17,
    },
  ],
  embeddedReferences: [
    {
      reference: '"s3://path/to/bucket.parquet"',
      type: "join",
      index: 12,
      referenceIndex: 17,
    },
  ],
};

export const fromInCTE: ReferenceTestCase = {
  query: `WITH x as (select * from tbl) FROM x`,
  references: [
    { reference: "tbl", type: "from", index: 20, referenceIndex: 25 },
    { reference: "x", type: "from", index: 30, referenceIndex: 35 },
  ],
  embeddedReferences: [],
};

export const tests = [
  singleFrom,
  bigGapInFrom,
  subqueryFrom,
  unfinishedFrom,
  simpleJoin,
  remoteSourceFrom,
  remoteSourceFromAlias,
  remoteSourceJoin,
  fromInCTE,
];
