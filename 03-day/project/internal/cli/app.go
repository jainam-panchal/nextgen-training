package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"jainam-panchal/nextgen-training/text-editor/internal/editor"
)

func Run(ctx context.Context) error {
	ed := editor.NewEditor()
	lineCh, errCh := startInputReader()

	fmt.Print("> ")

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-errCh:
			if err == nil {
				return nil
			}
			return err
		case line, ok := <-lineCh:
			if !ok {
				return nil
			}

			line = strings.TrimSpace(line)
			if line == "" {
				fmt.Print("> ")
				continue
			}

			shouldExit, err := handleCommand(ed, line)
			if err != nil {
				fmt.Println("error:", err)
				fmt.Print("> ")
				continue
			}

			if shouldExit {
				return nil
			}

			fmt.Print("> ")
		}
	}
}

func startInputReader() (<-chan string, <-chan error) {
	lineCh := make(chan string)
	errCh := make(chan error, 1)

	go func() {
		defer close(lineCh)
		defer close(errCh)

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			lineCh <- scanner.Text()
		}

		errCh <- scanner.Err()
	}()

	return lineCh, errCh
}

func handleCommand(ed *editor.Editor, line string) (bool, error) {
	fields := strings.Fields(line)
	switch fields[0] {
	case "exit":
		return true, nil
	case "undo":
		return false, handleUndo(ed)
	case "redo":
		return false, handleRedo(ed)
	case "print":
		fmt.Println(ed.DocumentString())
		return false, nil
	case "history":
		fmt.Println(ed.HistoryString())
		return false, nil
	case "help":
		printHelp()
		return false, nil
	case "insert":
		return false, handleInsert(ed, line)
	case "delete":
		return false, handleDelete(ed, fields)
	case "replace":
		return false, handleReplace(ed, line)
	case "format":
		return false, handleFormat(ed, fields)
	default:
		return false, fmt.Errorf("unknown command, type help")
	}
}

func handleUndo(ed *editor.Editor) error {
	if err := ed.Undo(); err != nil {
		return err
	}

	printState(ed)
	return nil
}

func handleRedo(ed *editor.Editor) error {
	if err := ed.Redo(); err != nil {
		return err
	}

	printState(ed)
	return nil
}

func handleInsert(ed *editor.Editor, line string) error {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return fmt.Errorf("usage: insert <pos> <text>")
	}

	pos, err := parseInt(fields[1])
	if err != nil {
		return err
	}

	text, err := parseInsertText(line)
	if err != nil {
		return err
	}

	return applyAndPrint(ed, &editor.InsertOperation{
		Position: pos,
		Text:     text,
	})
}

func handleDelete(ed *editor.Editor, fields []string) error {
	if len(fields) != 3 {
		return fmt.Errorf("usage: delete <pos> <length>")
	}

	pos, err := parseInt(fields[1])
	if err != nil {
		return err
	}

	length, err := parseInt(fields[2])
	if err != nil {
		return err
	}

	return applyAndPrint(ed, &editor.DeleteOperation{
		Position: pos,
		Length:   length,
	})
}

func handleReplace(ed *editor.Editor, line string) error {
	pos, oldText, newText, err := parseReplaceInput(line)
	if err != nil {
		return err
	}

	return applyAndPrint(ed, &editor.ReplaceOperation{
		Position: pos,
		OldText:  oldText,
		NewText:  newText,
	})
}

func handleFormat(ed *editor.Editor, fields []string) error {
	if len(fields) != 5 {
		return fmt.Errorf("usage: format <pos> <length> <bold> <italic>")
	}

	pos, err := parseInt(fields[1])
	if err != nil {
		return err
	}

	length, err := parseInt(fields[2])
	if err != nil {
		return err
	}

	bold, err := strconv.ParseBool(fields[3])
	if err != nil {
		return err
	}

	italic, err := strconv.ParseBool(fields[4])
	if err != nil {
		return err
	}

	return applyAndPrint(ed, &editor.FormatOperation{
		Position: pos,
		Length:   length,
		Bold:     bold,
		Italic:   italic,
	})
}

func parseInt(value string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func parseInsertText(line string) (string, error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("usage: insert <pos> <text>")
	}

	text := strings.TrimSpace(parts[2])
	if len(text) >= 2 && text[0] == '"' && text[len(text)-1] == '"' {
		text = text[1 : len(text)-1]
	}

	return text, nil
}

func parseReplaceInput(line string) (int, string, string, error) {
	parts := strings.SplitN(line, " ", 3)
	if len(parts) < 3 {
		return 0, "", "", fmt.Errorf("usage: replace <pos> <old> <new>")
	}

	pos, err := parseInt(parts[1])
	if err != nil {
		return 0, "", "", err
	}

	args := strings.TrimSpace(parts[2])
	if args == "" {
		return 0, "", "", fmt.Errorf("usage: replace <pos> <old> <new>")
	}

	if args[0] == '"' {
		oldText, rest, ok := extractQuoted(args)
		if !ok {
			return 0, "", "", fmt.Errorf("usage: replace <pos> <old> <new>")
		}

		newText := strings.TrimSpace(rest)
		if len(newText) >= 2 && newText[0] == '"' && newText[len(newText)-1] == '"' {
			newText = newText[1 : len(newText)-1]
		}
		if newText == "" {
			return 0, "", "", fmt.Errorf("usage: replace <pos> <old> <new>")
		}

		return pos, oldText, newText, nil
	}

	fields := strings.Fields(args)
	if len(fields) < 2 {
		return 0, "", "", fmt.Errorf("usage: replace <pos> <old> <new>")
	}

	oldText := fields[0]
	newText := strings.Join(fields[1:], " ")
	return pos, oldText, newText, nil
}

func extractQuoted(value string) (string, string, bool) {
	if len(value) < 2 || value[0] != '"' {
		return "", "", false
	}

	for i := 1; i < len(value); i++ {
		if value[i] == '"' {
			return value[1:i], value[i+1:], true
		}
	}

	return "", "", false
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  insert <pos> <text>")
	fmt.Println("  delete <pos> <length>")
	fmt.Println("  replace <pos> <old> <new>")
	fmt.Println("  format <pos> <length> <bold> <italic>")
	fmt.Println("  undo")
	fmt.Println("  redo")
	fmt.Println("  print")
	fmt.Println("  history")
	fmt.Println("  exit")
}

func applyAndPrint(ed *editor.Editor, op editor.EditOperation) error {
	if err := ed.ApplyOperation(op); err != nil {
		return err
	}

	printState(ed)
	return nil
}

func printState(ed *editor.Editor) {
	fmt.Println(ed.DocumentString())
	fmt.Println(ed.HistoryString())
}
