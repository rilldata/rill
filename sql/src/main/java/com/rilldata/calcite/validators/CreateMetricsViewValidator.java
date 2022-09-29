package com.rilldata.calcite.validators;

import com.rilldata.calcite.models.SqlCreateMetricsView;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlLiteral;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.SqlNodeList;
import org.apache.calcite.sql.SqlSelect;
import org.apache.calcite.sql.parser.SqlParseException;
import org.apache.calcite.sql.parser.SqlParserPos;
import org.apache.calcite.tools.Planner;
import org.apache.calcite.tools.ValidationException;

public class CreateMetricsViewValidator
{
  /**
   * Validates create metrics view query by parsing and validating group by queries
   * created from the dimensions and measures specified in the modeling query
   */
  public static String validateModelingQuery(SqlCreateMetricsView sqlCreateMetricsView, SqlDialect sqlDialect,
      Planner planner
  ) throws ValidationException
  {
    SqlNodeList dimensions = sqlCreateMetricsView.dimensions;
    SqlNodeList groupByList = new SqlNodeList(SqlParserPos.ZERO);
    for (int i = 1; i <= dimensions.size(); i++) {
      groupByList.add(SqlLiteral.createExactNumeric(i + "", SqlParserPos.ZERO));
    }
    for (SqlNode measure : sqlCreateMetricsView.measures.getList()) {
      SqlNodeList selectList = new SqlNodeList(dimensions, SqlParserPos.ZERO);
      selectList.add(measure);
      SqlSelect groupBy = new SqlSelect(
          SqlParserPos.ZERO,
          SqlNodeList.EMPTY,
          selectList,
          sqlCreateMetricsView.from,
          null,
          groupByList,
          null,
          null,
          null,
          null,
          null,
          null
      );
      String sqlString = groupBy.toSqlString(sqlDialect).toString();
      SqlNode parsed = null;
      try {
        parsed = planner.parse(sqlString);
      } catch (SqlParseException e) {
        throw new RuntimeException(String.format("Parsing failed for query we created %s", sqlString));
      }
      SqlNode validated = planner.validate(parsed);
      planner.close();
    }
    // dimensions, measures and table are validated, return sql string to be stored in db
    return sqlCreateMetricsView.toSqlString(sqlDialect).toString();
  }
}
