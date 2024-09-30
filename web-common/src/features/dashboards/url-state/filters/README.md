# URL param for filters

Contains the parser for url param for metrics view filter.
We support simple dimension and measure filters.

Example dimension filter,
```
country IN ('US','IN') AND state = 'ABC'
country NOT IN ('US','IN') AND (state = 'ABC' OR lat >= 12.56)
```

Example measure filter,
```
country NIN ('US','IN') and state having (lat >= 12.56)
```

## Updating the grammar

`expression.ne` has the [nearley](https://nearley.js.org/) parser.

Run `npm run build-filter-grammar -w web-common` to generate the compiled grammer.
