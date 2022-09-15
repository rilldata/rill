package com.rilldata.calcite.utility;

import org.apache.calcite.rel.type.RelDataType;
import org.apache.calcite.rel.type.RelDataTypeFactory;
import org.apache.calcite.sql.SqlOperatorBinding;
import org.apache.calcite.sql.type.SqlReturnTypeInference;
import org.apache.calcite.sql.type.SqlTypeName;
import org.apache.druid.sql.calcite.planner.Calcites;

/**
 * Methods resued from https://github.com/apache/druid/blob/master/sql/src/main/java/org/apache/druid/sql/calcite/planner/Calcites.java
 * */
public class CalciteUtils
{
  public static final SqlReturnTypeInference
      ARG1_NULLABLE_RETURN_TYPE_INFERENCE = new Arg1NullableTypeInference();

  public static class Arg1NullableTypeInference implements SqlReturnTypeInference
  {
    @Override
    public RelDataType inferReturnType(SqlOperatorBinding opBinding)
    {
      RelDataType type = opBinding.getOperandType(1);
      return Calcites.createSqlTypeWithNullability(
          opBinding.getTypeFactory(),
          type.getSqlTypeName(),
          true
      );
    }
  }

  public static final SqlReturnTypeInference TYPE_INFERENCE =
      opBinding -> {
        final RelDataTypeFactory typeFactory = opBinding.getTypeFactory();

        final int n = opBinding.getOperandCount();
        if (n == 0) {
          return typeFactory.createSqlType(SqlTypeName.NULL);
        }

        SqlTypeName returnSqlTypeName = SqlTypeName.NULL;
        boolean hasDouble = false;
        boolean isString = false;
        for (int i = 0; i < n; i++) {
          RelDataType type = opBinding.getOperandType(i);
          SqlTypeName sqlTypeName = type.getSqlTypeName();

          // Return types are listed in order of preference:
          if (type.getSqlTypeName() != null) {
            if (SqlTypeName.CHAR_TYPES.contains(type.getSqlTypeName())) {
              returnSqlTypeName = sqlTypeName;
              isString = true;
              break;
            } else if (Calcites.isDoubleType(sqlTypeName)) {
              returnSqlTypeName = SqlTypeName.DOUBLE;
              hasDouble = true;
            } else if (Calcites.isLongType(sqlTypeName) && !hasDouble) {
              returnSqlTypeName = SqlTypeName.BIGINT;
            }
          } else if (sqlTypeName != SqlTypeName.NULL) {
            throw new RuntimeException(String.format("Argument %d has invalid type: %s", i, sqlTypeName));
          }
        }

        if (isString) {
          // String can be null in both modes
          return typeFactory.createTypeWithNullability(typeFactory.createSqlType(returnSqlTypeName), true);
        } else {
          return typeFactory.createSqlType(returnSqlTypeName);
        }
      };
}
