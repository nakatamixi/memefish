package parser

type prec int

const (
	precLit prec = iota
	precSelector
	precUnary
	precMulDiv
	precAddSub
	precBitShift
	precBitAnd
	precBitXor
	precBitOr
	precComparison
	precNot
	precAndOr
)

func exprPrec(e Expr) prec {
	switch e := e.(type) {
	case *CallExpr, *CountStarExpr, *CastExpr, *ExtractExpr, *CaseExpr, *ParenExpr, *ScalarSubQuery, *ArraySubQuery, *ExistsSubQuery, *Param, *Ident, *Path, *ArrayLiteral, *StructLiteral, *NullLiteral, *BoolLiteral, *IntLiteral, *FloatLiteral, *StringLiteral, *BytesLiteral, *DateLiteral, *TimestampLiteral:
		return precLit
	case *IndexExpr, *SelectorExpr:
		return precSelector
	case *InExpr, *IsNullExpr, *IsBoolExpr, *BetweenExpr:
		return precComparison
	case *BinaryExpr:
		switch e.Op {
		case OpMul, OpDiv:
			return precMulDiv
		case OpAdd, OpSub:
			return precAddSub
		case OpBitLeftShift, OpBitRightShift:
			return precBitShift
		case OpBitAnd:
			return precBitAnd
		case OpBitXor:
			return precBitXor
		case OpBitOr:
			return precBitOr
		case OpEqual, OpNotEqual, OpLess, OpLessEqual, OpGreater, OpGreaterEqual, OpLike, OpNotLike:
			return precComparison
		case OpAnd, OpOr:
			return precAndOr
		}
	case *UnaryExpr:
		switch e.Op {
		case OpPlus, OpMinus, OpBitNot:
			return precUnary
		case OpNot:
			return precNot
		}
	}

	panic("exprPrec: unexpected")
}

func paren(p prec, e Expr) string {
	ep := exprPrec(e)
	if ep <= p {
		return e.SQL()
	} else {
		return "(" + e.SQL() + ")"
	}
}

// ================================================================================
//
// SELECT
//
// ================================================================================

func (q *QueryStatement) SQL() string {
	var sql string
	if q.Hint != nil {
		sql += q.Hint.SQL() + " "
	}
	sql += q.Query.SQL()
	return sql
}

func (h *Hint) SQL() string {
	sql := "@{" + h.Records[0].SQL()
	for _, r := range h.Records[1:] {
		sql += ", " + r.SQL()
	}
	sql += "}"
	return sql
}

func (h *HintRecord) SQL() string {
	return h.Key.SQL() + "=" + h.Value.SQL()
}

func (s *Select) SQL() string {
	sql := "SELECT "
	if s.Distinct {
		sql += "DISTINCT "
	}
	if s.AsStruct {
		sql += "AS STRUCT "
	}
	sql += s.Results[0].SQL()
	for _, r := range s.Results[1:] {
		sql += ", " + r.SQL()
	}
	if s.From != nil {
		sql += " " + s.From.SQL()
	}
	if s.Where != nil {
		sql += " " + s.Where.SQL()
	}
	if s.GroupBy != nil {
		sql += " " + s.GroupBy.SQL()
	}
	if s.Having != nil {
		sql += " " + s.Having.SQL()
	}
	if s.OrderBy != nil {
		sql += " " + s.OrderBy.SQL()
	}
	if s.Limit != nil {
		sql += " " + s.Limit.SQL()
	}
	return sql
}

func (c *CompoundQuery) SQL() string {
	op := string(c.Op)
	if c.Distinct {
		op += " DISTINCT"
	} else {
		op += " ALL"
	}

	sql := c.Queries[0].SQL()
	for _, q := range c.Queries[1:] {
		sql += " " + op + " " + q.SQL()
	}
	if c.OrderBy != nil {
		sql += " " + c.OrderBy.SQL()
	}
	if c.Limit != nil {
		sql += " " + c.Limit.SQL()
	}
	return sql
}

func (s *SubQuery) SQL() string {
	sql := "(" + s.Query.SQL() + ")"
	if s.OrderBy != nil {
		sql += " " + s.OrderBy.SQL()
	}
	if s.Limit != nil {
		sql += " " + s.Limit.SQL()
	}
	return sql
}

func (s *Star) SQL() string {
	return "*"
}

func (s *StarPath) SQL() string {
	return s.Expr.SQL() + ".*"
}

func (a *Alias) SQL() string {
	return a.Expr.SQL() + " " + a.As.SQL()
}

func (a *AsAlias) SQL() string {
	return "AS " + a.Alias.SQL()
}

func (e *ExprSelectItem) SQL() string {
	return e.Expr.SQL()
}

func (f *From) SQL() string {
	return "FROM " + f.Source.SQL()
}

func (w *Where) SQL() string {
	return "WHERE " + w.Expr.SQL()
}

func (g *GroupBy) SQL() string {
	sql := "GROUP BY " + g.Exprs[0].SQL()
	for _, e := range g.Exprs[1:] {
		sql += ", " + e.SQL()
	}
	return sql
}

func (h *Having) SQL() string {
	return "HAVING " + h.Expr.SQL()
}

func (o *OrderBy) SQL() string {
	sql := "ORDER BY " + o.Items[0].SQL()
	for _, item := range o.Items[1:] {
		sql += ", " + item.SQL()
	}
	return sql
}

func (o *OrderByItem) SQL() string {
	sql := o.Expr.SQL()
	if o.Collate != nil {
		sql += " " + o.Collate.SQL()
	}
	if o.Dir != "" {
		sql += " " + string(o.Dir)
	}
	return sql
}

func (c *Collate) SQL() string {
	return "COLLATE " + c.Value.SQL()
}

func (l *Limit) SQL() string {
	sql := "LIMIT " + l.Count.SQL()
	if l.Offset != nil {
		sql += " " + l.Offset.SQL()
	}
	return sql
}

func (o *Offset) SQL() string {
	return "OFFSET " + o.Value.SQL()
}

// ================================================================================
//
// JOIN
//
// ================================================================================

func (u *Unnest) SQL() string {
	var sql string
	if u.Implicit {
		sql += u.Expr.SQL()
	} else {
		sql += "UNNEST(" + u.Expr.SQL() + ")"
	}
	if u.Hint != nil {
		sql += " " + u.Hint.SQL()
	}
	if u.As != nil {
		sql += " " + u.As.SQL()
	}
	if u.WithOffset != nil {
		sql += " " + u.WithOffset.SQL()
	}
	if u.Sample != nil {
		sql += " " + u.Sample.SQL()
	}
	return sql
}

func (w *WithOffset) SQL() string {
	sql := "WITH OFFSET"
	if w.As != nil {
		sql += " " + w.As.SQL()
	}
	return sql
}

func (t *TableName) SQL() string {
	sql := t.Table.SQL()
	if t.Hint != nil {
		sql += " " + t.Hint.SQL()
	}
	if t.As != nil {
		sql += " " + t.As.SQL()
	}
	if t.Sample != nil {
		sql += " " + t.Sample.SQL()
	}
	return sql
}

func (s *SubQueryJoinExpr) SQL() string {
	sql := "(" + s.Query.SQL() + ")"
	if s.As != nil {
		sql += " " + s.As.SQL()
	}
	if s.Sample != nil {
		sql += " " + s.Sample.SQL()
	}
	return sql
}

func (p *ParenJoinExpr) SQL() string {
	sql := "(" + p.Source.SQL() + ")"
	if p.Sample != nil {
		sql += " " + p.Sample.SQL()
	}
	return sql
}

func (j *Join) SQL() string {
	sql := j.Left.SQL()
	if j.Op != CommaJoin {
		sql += " "
	}
	sql += string(j.Op) + " "
	if j.Hint != nil {
		sql += j.Hint.SQL() + " "
	}
	sql += j.Right.SQL()
	if j.Cond != nil {
		sql += " " + j.Cond.SQL()
	}
	return sql
}

func (o *On) SQL() string {
	return "ON " + o.Expr.SQL()
}

func (u *Using) SQL() string {
	sql := "USING (" + u.Idents[0].SQL()
	for _, id := range u.Idents[1:] {
		sql += ", " + id.SQL()
	}
	sql += ")"
	return sql
}

func (t *TableSample) SQL() string {
	return "TABLESAMPLE " + string(t.Method) + " " + t.Size.SQL()
}

func (t *TableSampleSize) SQL() string {
	return "(" + t.Value.SQL() + " " + string(t.Unit) + ")"
}

// ================================================================================
//
// Expr
//
// ================================================================================

func (b *BinaryExpr) SQL() string {
	p := exprPrec(b)
	sql := paren(p, b.Left)
	sql += " " + string(b.Op) + " "
	sql += paren(p, b.Right)
	return sql
}

func (u *UnaryExpr) SQL() string {
	p := exprPrec(u)
	if u.Op == OpNot {
		return "NOT " + paren(p, u.Expr)
	}
	return string(u.Op) + paren(p, u.Expr)
}

func (i *InExpr) SQL() string {
	p := exprPrec(i)
	sql := paren(p, i.Left)
	if i.Not {
		sql += " NOT"
	}
	sql += " IN "
	sql += i.Right.SQL()
	return sql
}

func (u *UnnestInCondition) SQL() string {
	return "UNNEST(" + u.Expr.SQL() + ")"
}

func (s *SubQueryInCondition) SQL() string {
	return "(" + s.Query.SQL() + ")"
}

func (v *ValuesInCondition) SQL() string {
	sql := "(" + v.Exprs[0].SQL()
	for _, e := range v.Exprs[1:] {
		sql += ", " + e.SQL()
	}
	sql += ")"
	return sql
}

func (i *IsNullExpr) SQL() string {
	p := exprPrec(i)
	sql := paren(p, i.Left)
	sql += " IS "
	if i.Not {
		sql += "NOT "
	}
	sql += "NULL"
	return sql
}

func (i *IsBoolExpr) SQL() string {
	p := exprPrec(i)
	sql := paren(p, i.Left)
	sql += " IS "
	if i.Not {
		sql += "NOT "
	}
	if i.Right {
		sql += "TRUE"
	} else {
		sql += "FALSE"
	}
	return sql
}

func (b *BetweenExpr) SQL() string {
	p := exprPrec(b)
	sql := paren(p, b.Left)
	if b.Not {
		sql += " NOT"
	}
	return sql + " BETWEEN " + paren(p, b.RightStart) + " AND " + paren(p, b.RightEnd)
}

func (s *SelectorExpr) SQL() string {
	p := exprPrec(s)
	return paren(p, s.Expr) + "." + s.Member.SQL()
}

func (i *IndexExpr) SQL() string {
	p := exprPrec(i)
	sql := paren(p, i.Expr) + "["
	if i.Ordinal {
		sql += "ORDINAL"
	} else {
		sql += "OFFSET"
	}
	sql += "(" + i.Index.SQL() + ")]"
	return sql
}

func (c *CallExpr) SQL() string {
	sql := c.Func.SQL() + "("
	if c.Distinct {
		sql += "DISTINCT "
	}
	for i, a := range c.Args {
		if i != 0 {
			sql += ", "
		}
		sql += a.SQL()
	}
	sql += ")"
	return sql
}

func (a *Arg) SQL() string {
	if a.IntervalUnit != nil {
		return "INTERVAL " + a.Expr.SQL() + " " + a.IntervalUnit.SQL()
	}
	return a.Expr.SQL()
}

func (*CountStarExpr) SQL() string {
	return "COUNT(*)"
}

func (e *ExtractExpr) SQL() string {
	sql := "EXTRACT(" + e.Part.SQL() + " FROM " + e.Expr.SQL()
	if e.AtTimeZone != nil {
		sql += " " + e.AtTimeZone.SQL()
	}
	sql += ")"
	return sql
}

func (a *AtTimeZone) SQL() string {
	return "AT TIME ZONE " + a.Expr.SQL()
}

func (c *CastExpr) SQL() string {
	return "CAST(" + c.Expr.SQL() + " AS " + c.Type.SQL() + ")"
}

func (c *CaseExpr) SQL() string {
	sql := "CASE "
	if c.Expr != nil {
		sql += c.Expr.SQL() + " "
	}
	for _, w := range c.Whens {
		sql += w.SQL() + " "
	}
	if c.Else != nil {
		sql += c.Else.SQL() + " "
	}
	sql += "END"
	return sql
}

func (c *CaseWhen) SQL() string {
	return "WHEN " + c.Cond.SQL() + " THEN " + c.Then.SQL()
}

func (c *CaseElse) SQL() string {
	return "ELSE " + c.Expr.SQL()
}

func (p *ParenExpr) SQL() string {
	return "(" + p.Expr.SQL() + ")"
}

func (s *ScalarSubQuery) SQL() string {
	return "(" + s.Query.SQL() + ")"
}

func (a *ArraySubQuery) SQL() string {
	return "ARRAY(" + a.Query.SQL() + ")"
}

func (e *ExistsSubQuery) SQL() string {
	sql := "EXISTS"
	if e.Hint != nil {
		sql += " " + e.Hint.SQL() + " "
	}
	sql += "(" + e.Query.SQL() + ")"
	return sql
}

func (p *Param) SQL() string {
	return "@" + p.Name
}

func (i *Ident) SQL() string {
	return QuoteSQLIdent(i.Name)
}

func (p *Path) SQL() string {
	sql := p.Idents[0].SQL()
	for _, id := range p.Idents[1:] {
		sql += "." + id.SQL()
	}
	return sql
}

func (a *ArrayLiteral) SQL() string {
	sql := "ARRAY"
	if a.Type != nil {
		sql += "<" + a.Type.SQL() + ">"
	}
	sql += "["
	for i, v := range a.Values {
		if i != 0 {
			sql += ", "
		}
		sql += v.SQL()
	}
	sql += "]"
	return sql
}

func (s *StructLiteral) SQL() string {
	sql := "STRUCT"
	if s.Fields != nil {
		sql += "<"
		for i, f := range s.Fields {
			if i != 0 {
				sql += ", "
			}
			sql += f.SQL()
		}
		sql += ">"
	}
	sql += "("
	for i, v := range s.Values {
		if i != 0 {
			sql += ", "
		}
		sql += v.SQL()
	}
	sql += ")"
	return sql
}

func (*NullLiteral) SQL() string {
	return "NULL"
}

func (b *BoolLiteral) SQL() string {
	if b.Value {
		return "TRUE"
	} else {
		return "FALSE"
	}
}

func (i *IntLiteral) SQL() string {
	return i.Value
}

func (f *FloatLiteral) SQL() string {
	return f.Value
}

func (s *StringLiteral) SQL() string {
	return QuoteSQLString(s.Value)
}

func (b *BytesLiteral) SQL() string {
	return QuoteSQLBytes(b.Value)
}

func (d *DateLiteral) SQL() string {
	return "DATE " + QuoteSQLString(d.Value)
}

func (t *TimestampLiteral) SQL() string {
	return "TIMESTAMP " + QuoteSQLString(t.Value)
}

// ================================================================================
//
// Type
//
// ================================================================================

func (s *SimpleType) SQL() string {
	return string(s.Name)
}

func (a *ArrayType) SQL() string {
	return "ARRAY<" + a.Item.SQL() + ">"
}

func (s *StructType) SQL() string {
	sql := "STRUCT<"
	for i, f := range s.Fields {
		if i != 0 {
			sql += ", "
		}
		sql += f.SQL()
	}
	sql += ">"
	return sql
}

func (f *FieldType) SQL() string {
	var sql string
	if f.Member != nil {
		sql += f.Member.SQL() + " "
	}
	sql += f.Type.SQL()
	return sql
}

// ================================================================================
//
// Cast for Special Cases
//
// ================================================================================

func (c *CastIntValue) SQL() string {
	return "CAST(" + c.Expr.SQL() + " AS INT64)"
}

func (c *CastNumValue) SQL() string {
	return "CAST(" + c.Expr.SQL() + " AS " + string(c.Type) + ")"
}
