package com.rilldata.calcite.visitors;

import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlIdentifier;
import org.apache.calcite.sql.SqlNode;
import org.apache.calcite.sql.util.SqlBasicVisitor;

/**
 * This expander will be called on a node if the column name or alias specified in the query matches a measure name
 * or alias specified in the METRICS VIEW modeling query.
 *
 * This will rewrite the column with actual definition of that measure. Example -
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
 *        M_DIST -> COUNT(DISTINCT DIM1) AS M_DIST
 *        M_DIST AS M1 -> COUNT(DISTINCT DIM1) AS M1
 * </pre>
 * */
public class MeasureExpander extends SqlBasicVisitor<SqlNode>
{
  // This is the actual measure definition from the model
  private final SqlCall measure;

  public MeasureExpander(SqlNode measure)
  {
    this.measure = (SqlCall) measure;
  }

  /**
   * CASE - M_DIST AS M1 -> COUNT(DISTINCT DIM1) AS M1
   * */
  @Override public SqlNode visit(SqlCall call)
  {
    call.setOperand(0, measure.operand(0));
    return call;
  }

  /**
   * CASE - M_DIST -> COUNT(DISTINCT DIM1) AS M_DIST
   * */
  @Override public SqlNode visit(SqlIdentifier id)
  {
    return measure;
  }
}
