// Package editor defines interfaces and operations
package editor

import "fmt"

type InsertOperation struct {
	Position int
	Text     string
}

type DeleteOperation struct {
	Position    int
	Length      int
	DeletedText string
}

type ReplaceOperation struct {
	Position int
	OldText  string
	NewText  string
}

type FormatOperation struct {
	Position int
	Length   int
	Bold     bool
	Italic   bool
}

func (op *FormatOperation) Apply(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid format position: %d", op.Position)
	}

	if op.Length <= 0 {
		return fmt.Errorf("invalid format length: %d", op.Length)
	}

	end := op.Position + op.Length
	if end > len(doc.Text) {
		return fmt.Errorf("format range out of bounds: position=%d length=%d", op.Position, op.Length)
	}

	doc.Formats = append(doc.Formats, TextFormat{
		Position: op.Position,
		Length:   op.Length,
		Bold:     op.Bold,
		Italic:   op.Italic,
	})
	return nil
}

func (op *FormatOperation) Reverse(doc *Document) error {
	for i := len(doc.Formats) - 1; i >= 0; i-- {
		format := doc.Formats[i]
		if format.Position == op.Position &&
			format.Length == op.Length &&
			format.Bold == op.Bold &&
			format.Italic == op.Italic {
			doc.Formats = append(doc.Formats[:i], doc.Formats[i+1:]...)
			return nil
		}
	}

	return nil
}

func (op *FormatOperation) String() string {
	return fmt.Sprintf(
		`Format(pos=%d, len=%d, bold=%t, italic=%t)`,
		op.Position,
		op.Length,
		op.Bold,
		op.Italic,
	)
}

func (op *ReplaceOperation) Apply(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid replace position: %d", op.Position)
	}

	if op.OldText == "" {
		return fmt.Errorf("invalid replace old text: empty")
	}

	end := op.Position + len(op.OldText)
	if end > len(doc.Text) {
		return fmt.Errorf("replace range out of bounds: position=%d length=%d", op.Position, len(op.OldText))
	}

	if doc.Text[op.Position:end] != op.OldText {
		return fmt.Errorf("replace text mismatch at position %d", op.Position)
	}

	doc.Text = doc.Text[:op.Position] + op.NewText + doc.Text[end:]
	return nil
}

func (op *ReplaceOperation) Reverse(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid replace position: %d", op.Position)
	}

	end := op.Position + len(op.NewText)
	if end > len(doc.Text) {
		return fmt.Errorf("replace reverse out of bounds: position=%d length=%d", op.Position, len(op.NewText))
	}

	if doc.Text[op.Position:end] != op.NewText {
		return fmt.Errorf("replace reverse text mismatch at position %d", op.Position)
	}

	doc.Text = doc.Text[:op.Position] + op.OldText + doc.Text[end:]
	return nil
}

func (op *ReplaceOperation) String() string {
	return fmt.Sprintf(
		`Replace(pos=%d, old=%q, new=%q)`,
		op.Position,
		op.OldText,
		op.NewText,
	)
}

func (op *DeleteOperation) Apply(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid delete position: %d", op.Position)
	}

	if op.Length <= 0 {
		return fmt.Errorf("invalid delete length: %d", op.Length)
	}

	end := op.Position + op.Length

	if end > len(doc.Text) {
		return fmt.Errorf("delete range out of bounds: position=%d length=%d", op.Position, op.Length)
	}

	op.DeletedText = doc.Text[op.Position:end]
	doc.Text = doc.Text[:op.Position] + doc.Text[end:]
	return nil
}

func (op *DeleteOperation) Reverse(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid delete position: %d", op.Position)
	}

	doc.Text = doc.Text[:op.Position] + op.DeletedText + doc.Text[op.Position:]
	return nil
}

func (op *DeleteOperation) String() string {
	return fmt.Sprintf(`Delete(pos=%d, len=%d, text=%q)`, op.Position, op.Length,
		op.DeletedText)
}

func (op *InsertOperation) Apply(doc *Document) error {
	if op.Position < 0 || op.Position > len(doc.Text) {
		return fmt.Errorf("invalid insert position: %d", op.Position)
	}

	doc.Text = doc.Text[:op.Position] + op.Text + doc.Text[op.Position:]
	return nil
}

func (op *InsertOperation) Reverse(doc *Document) error {
	end := op.Position + len(op.Text)

	if op.Position < 0 || end > len(doc.Text) {
		return fmt.Errorf("invalid insert position: %d", op.Position)
	}

	doc.Text = doc.Text[:op.Position] + doc.Text[end:]
	return nil
}

func (op *InsertOperation) String() string {
	return fmt.Sprintf(`Insert(pos=%d, text=%q)`, op.Position, op.Text)
}
