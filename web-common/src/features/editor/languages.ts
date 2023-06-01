import { SQLDialect } from "@codemirror/lang-sql";

export const DuckDBSQL: SQLDialect = SQLDialect.define({
  keywords:
    "select from where group by all having order limit sample unnest with window qualify values filter exclude replace like ilike glob as case when then end in cast left join on not desc asc sum union",
});
