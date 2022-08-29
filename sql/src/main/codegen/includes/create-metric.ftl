SqlCreate SqlCreateMetric(Span s, boolean replace) :
{
    final SqlIdentifier id;
    final List<SqlNode> dimList;
    final SqlNode fromClause;
    final List<SqlNode> measureList;
    final SqlNode from;
}
{
    <METRICS> <VIEW> id = SimpleIdentifier()
    <DIMENSIONS>
    dimList = SelectList()
    <MEASURES>
    measureList = SelectList()
    <FROM> from = FromClause()
    {
        return new SqlCreateMetric(s.end(this), id, new SqlNodeList(dimList, Span.of(dimList).pos()), new SqlNodeList(measureList, Span.of(measureList).pos()), from);
    }
}
