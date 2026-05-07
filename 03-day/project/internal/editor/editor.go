package editor

import (
	"fmt"
	"strings"

	"jainam-panchal/nextgen-training/text-editor/internal/container"
)

type TextFormat struct {
	Position int
	Length   int
	Bold     bool
	Italic   bool
}

type EditOperation interface {
	Apply(doc *Document) error
	Reverse(doc *Document) error
	String() string
}

type Document struct {
	Text    string
	Formats []TextFormat
}

type Editor struct {
	Doc     Document
	History container.DoublyLinkedList[EditOperation]
	Current *container.Node[EditOperation]
}

func NewEditor() *Editor {
	return &Editor{
		Doc: Document{
			Text:    "",
			Formats: make([]TextFormat, 0),
		},
	}
}

func (e *Editor) ApplyOperation(op EditOperation) error {
	if err := op.Apply(&e.Doc); err != nil {
		return err
	}

	// if we've gone back with undo and then applied the operation
	if e.Current != e.History.Tail {
		e.History.RemoveAfter(e.Current)
	}

	e.Current = e.History.Append(op)
	return nil
}

func (e *Editor) Undo() error {
	if e.Current == nil {
		return fmt.Errorf("no operation to undo")
	}

	if err := e.Current.Data.Reverse(&e.Doc); err != nil {
		return err
	}

	e.Current = e.Current.Prev
	return nil
}

func (e *Editor) Redo() error {
	var next *container.Node[EditOperation]

	if e.Current == nil {
		next = e.History.Head
	} else {
		next = e.Current.Next
	}

	if next == nil {
		return fmt.Errorf("no operation to redo")
	}

	if err := next.Data.Apply(&e.Doc); err != nil {
		return err
	}

	e.Current = next
	return nil
}

func (e *Editor) DocumentString() string {
	return fmt.Sprintf("Text: %q | Formats: %v", e.Doc.Text, e.Doc.Formats)
}

func (e *Editor) HistoryString() string {
	if e.History.Head == nil {
		return "[]"
	}

	var builder strings.Builder
	builder.WriteString("[")

	for node := e.History.Head; node != nil; node = node.Next {
		if node == e.Current {
			builder.WriteString("*")
		}

		builder.WriteString(fmt.Sprintf("%v", node.Data))

		if node.Next != nil {
			builder.WriteString(" <-> ")
		}
	}

	if e.Current == nil {
		builder.WriteString(" *")
	}

	builder.WriteString("]")
	return builder.String()
}
