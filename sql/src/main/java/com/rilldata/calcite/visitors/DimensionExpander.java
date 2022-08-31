package com.rilldata.calcite.visitors;

import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlKind;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.util.SqlBasicVisitor;

/**
 * This expander will be called on a node if the column name or alias specified in the query matches a dimension name
 * or alias specified in the METRICS VIEW modeling query.
 *
 * This will rewrite the column with actual definition of that dimension. Example -
 * <pre>
 * For Modeling query -
 *         CREATE METRICS VIEW Test1
 *         DIMENSIONS
 *         DIM1, ceil("MET1") AS MET_C, DIM2
 *         MEASURES
 *         COUNT(DISTINCT DIM1) AS M_DIST,
 *         AVG(DISTINCT MET1) AS M_AVG
 *         FROM MAIN.TEST
 * Replacement to columns specified in query -
 *        DIM1 -> DIM1
 *        DIM1 AS D -> DIM1 AS D
 *        MET_C -> ceil("MET1") AS MET_C
 *        MET_C AS D -> ceil("MET1") AS D
 * </pre>
 * */
public class DimensionExpander extends SqlBasicVisitor<SqlNode>
{
  // This is the actual dimension definition from the model
  private final SqlNode dimension;

  public DimensionExpander(SqlNode dimension)
  {
    this.dimension = dimension;
  }

  /**
   * CASE - DIM1 AS D -> DIM1 AS D
   *        MET_C AS D -> ceil("MET1") AS D
   * */
  @Override public SqlNode visit(SqlCall call)
  {
    if (dimension.getKind().equals(SqlKind.IDENTIFIER)) {
      call.setOperand(0, dimension);
    } else if (dimension.getKind().equals(SqlKind.AS)) {
      call.setOperand(0, ((SqlCall) dimension).operand(0));
    }
    return call;
  }

  /**
   * CASE - DIM1 -> DIM1
   *        MET_C -> ceil("MET1") AS MET_C
   * */
  @Override public SqlNode visit(SqlIdentifier id)
  {
    return dimension;
  }
}
