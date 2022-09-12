package com.rilldata.calcite.dialects;

import org.apache.calcite.avatica.util.Casing;
import org.apache.calcite.sql.SqlCall;
import org.apache.calcite.sql.SqlDialect;
import org.apache.calcite.sql.SqlFunction;
import org.apache.calcite.sql.SqlWriter;
import org.apache.druid.sql.calcite.planner.DruidTypeSystem;

public class DruidDialect extends SqlDialect
{
  public static final Context DEFAULT_CONTEXT = SqlDialect.EMPTY_CONTEXT
      .withDatabaseProduct(DatabaseProduct.UNKNOWN)
      .withIdentifierQuoteString("\"")
      .withUnquotedCasing(Casing.UNCHANGED)
      .withQuotedCasing(Casing.UNCHANGED)
      .withCaseSensitive(true)
      .withDataTypeSystem(DruidTypeSystem.INSTANCE);

  public static final SqlDialect DEFAULT = new DruidDialect(DEFAULT_CONTEXT);

  public DruidDialect(Context context)
  {
    super(context);
  }

  @Override public boolean supportsCharSet()
  {
    return false;
  }

  @Override public boolean supportsGroupByLiteral()
  {
    return false;
  }

  @Override public void unparseCall(SqlWriter writer, SqlCall call, int leftPrec, int rightPrec)
  {
    switch (call.getOperator().getName()) {
      // rewrite XOR to Druid's BITWISE_XOR
    case "XOR":
      SqlFunction xor = (SqlFunction) call.getOperator();
      SqlFunction bitWiseXOR = new SqlFunction(
          "BITWISE_XOR",
          xor.getKind(),
          xor.getReturnTypeInference(),
          xor.getOperandTypeInference(),
          xor.getOperandTypeChecker(), xor.getFunctionType()
      );
      SqlCall bitwiseXorCall = bitWiseXOR.createCall(call.getFunctionQuantifier(), call.getParserPosition(),
          call.getOperandList()
      );
      super.unparseCall(writer, bitwiseXorCall, leftPrec, rightPrec);
      break;
    default:
      super.unparseCall(writer, call, leftPrec, rightPrec);
    }
  }
}
