package main
/*
	A series of fields should be taken as "structure contents", and *not*
	just the single 'field' elements. Namely, 'fields' subsumes 'field'
	elements.

	More particularly, 'fields' corresponds to list, in order, of any of the
	follow elements: pad, field, list, localfield, exprfield, valueparm
	and switch.

	Thus, the 'Field' type must contain the union of information corresponding
	to all aforementioned fields.

	This would ideally be a better job for interfaces, but I could not figure
	out how to make them jive with Go's XML package. (And I don't really feel
	up to type translation.)
*/

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
)

type Fields []*Field

type Field struct {
	XMLName xml.Name

	// For 'pad' element
	Bytes int `xml:"bytes,attr"`

	// For 'field', 'list', 'localfield', 'exprfield' and 'switch' elements.
	Name Name `xml:"name,attr"`

	// For 'field', 'list', 'localfield', and 'exprfield' elements.
	Type Type `xml:"type,attr"`

	// For 'list', 'exprfield' and 'switch' elements.
	Expr *Expression `xml:",any"`

	// For 'valueparm' element.
	ValueMaskType Type `xml:"value-mask-type,attr"`
	ValueMaskName Name `xml:"value-mask-name,attr"`
	ValueListName Name `xml:"value-list-name,attr"`

	// For 'switch' element.
	Bitcases Bitcases `xml:"bitcase"`

	// I don't know which elements these are for. The documentation is vague.
	// They also seem to be completely optional.
	OptEnum Type `xml:"enum,attr"`
	OptMask Type `xml:"mask,attr"`
	OptAltEnum Type `xml:"altenum,attr"`
}

// String is for debugging purposes.
func (f *Field) String() string {
	switch f.XMLName.Local {
	case "pad":
		return fmt.Sprintf("pad (%d bytes)", f.Bytes)
	case "field":
		return fmt.Sprintf("field (type = '%s', name = '%s')", f.Type, f.Name)
	case "list":
		return fmt.Sprintf("list (type = '%s', name = '%s', length = '%s')",
			f.Type, f.Name, f.Expr)
	case "localfield":
		return fmt.Sprintf("localfield (type = '%s', name = '%s')",
			f.Type, f.Name)
	case "exprfield":
		return fmt.Sprintf("exprfield (type = '%s', name = '%s', expr = '%s')",
			f.Type, f.Name, f.Expr)
	case "valueparam":
		return fmt.Sprintf("valueparam (type = '%s', name = '%s', list = '%s')",
			f.ValueMaskType, f.ValueMaskName, f.ValueListName)
	case "switch":
		bitcases := make([]string, len(f.Bitcases))
		for i, bitcase := range f.Bitcases {
			bitcases[i] = bitcase.StringPrefix("\t")
		}
		return fmt.Sprintf("switch (name = '%s', expr = '%s')\n\t%s",
			f.Name, f.Expr, strings.Join(bitcases, "\n\t"))
	default:
		log.Panicf("Unrecognized field element: %s", f.XMLName.Local)
	}

	panic("unreachable")
}

type Bitcases []*Bitcase

// Bitcase represents a single expression followed by any number of fields.
// Namely, if the switch's expression (all bitcases are inside a switch),
// and'd with the bitcase's expression is equal to the bitcase expression,
// then the fields should be included in its parent structure.
// Note that since a bitcase is unique in that expressions and fields are
// siblings, we must exhaustively search for one of them. Essentially,
// it's the closest thing to a Union I can get to in Go without interfaces.
// Would an '<expression>' tag have been too much to ask? :-(
type Bitcase struct {
	Fields Fields `xml:",any"`

	// All the different expressions.
	// When it comes time to choose one, use the 'Expr' method.
	ExprOp *Expression `xml:"op"`
	ExprUnOp *Expression `xml:"unop"`
	ExprField *Expression `xml:"fieldref"`
	ExprValue *Expression `xml:"value"`
	ExprBit *Expression `xml:"bit"`
	ExprEnum *Expression `xml:"enumref"`
	ExprSum *Expression `xml:"sumof"`
	ExprPop *Expression `xml:"popcount"`
}

// StringPrefix is for debugging purposes only.
// StringPrefix takes a string to prefix to every extra line for formatting.
func (b *Bitcase) StringPrefix(prefix string) string {
	fields := make([]string, len(b.Fields))
	for i, field := range b.Fields {
		fields[i] = fmt.Sprintf("%s%s", prefix, field)
	}
	return fmt.Sprintf("%s\n\t%s%s", b.Expr(), prefix,
		strings.Join(fields, "\n\t"))
}

// Expr chooses the only non-nil Expr* field from Bitcase.
// Panic if there is more than one non-nil expression.
func (b *Bitcase) Expr() *Expression {
	choices := []*Expression{
		b.ExprOp, b.ExprUnOp, b.ExprField, b.ExprValue,
		b.ExprBit, b.ExprEnum, b.ExprSum, b.ExprPop,
	}

	var choice *Expression = nil
	numNonNil := 0
	for _, c := range choices {
		if c != nil {
			numNonNil++
			choice = c
		}
	}

	if choice == nil {
		log.Panicf("No top level expression found in a bitcase.")
	}
	if numNonNil > 1 {
		log.Panicf("More than one top-level expression was found in a bitcase.")
	}
	return choice
}