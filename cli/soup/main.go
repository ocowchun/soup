package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/ocowchun/soup/evaluator"
	"github.com/ocowchun/soup/lexer"
	"github.com/ocowchun/soup/parser"
	"golang.org/x/term"
)

func main() {
	fmt.Println("welcome to soup")

	if len(os.Args) == 1 {
		fmt.Println("repl")
		err := repl()
		if err != nil {
			printError(err)
			//fmt.Println("error:", err)
			os.Exit(65)
		}

	} else if len(os.Args) == 2 {
		f := os.Args[1]
		fmt.Println("file", f)
		err := runFile(f)
		if err != nil {
			printError(err)
			//}
			os.Exit(65)
		}

	} else {
		fmt.Println("unsupport args")
	}

}

func printError(err error) {
	var parsingError *parser.ParsingError
	if errors.As(err, &parsingError) {
		fmt.Printf("Parsing error at line %d, got token: `%s` type: %s, error: %s\n",
			parsingError.Token.Line, parsingError.Token.Content, parsingError.Token.TokenType,
			parsingError.Message)
		return
	}

	var runtimeError *evaluator.RuntimeError
	if errors.As(err, &runtimeError) {

		fmt.Println(err.Error())
		for _, e := range runtimeError.StackTrace() {
			fmt.Printf("\t at %s (line %d)\n", e.IdentifierName(), e.LineNumber())
		}

		fmt.Printf("\t at main (line %d)\n", runtimeError.LineNumber())
		return
	}

	fmt.Println("panic:", err)
}

func repl() error {

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buf := make([]byte, 3)
	fmt.Print("soup> ")
	lines := make([]string, 0)
	currentLine := ""
	lineIndex := 0
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}

		switch {
		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 68:
			// Left arrow
			fmt.Print("\033[1D")
		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 67:
			// Right arrow
			fmt.Print("\033[1C")
		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 65:
			// Up arrow
			//fmt.Print("\033[A")

			fmt.Print("\033[2K\r")
			fmt.Print("soup> ")
			if lineIndex >= 0 {
				currentLine = lines[lineIndex]
				fmt.Print(currentLine)
				lineIndex--
			}
		case n == 3 && buf[0] == 27 && buf[1] == 91 && buf[2] == 66:
			// Down arrow
			//fmt.Print("\033[B")
			// clean the line
			fmt.Print("\033[2K\r")
			fmt.Print("soup> ")

		case n == 1 && buf[0] == 127:
			// Backspace
			fmt.Print("\033[1D \033[1D")
		case buf[0] == 3: // Ctrl+C
			return nil
		case n == 1 && (buf[0] == '\r' || buf[0] == '\n'):
			lines = append(lines, currentLine)
			fmt.Print("\033[2K\r")
			fmt.Print(currentLine)
			fmt.Print("\n\r") // Move to next line or handle input

			currentLine = ""
			lineIndex = len(lines) - 1
			fmt.Print("soup> ")
			//continue
			// Enter/Return pressed
		default:
			currentLine += string(buf[:n])
			fmt.Print(string(buf[:n]))
		}
	}

	//scanner := bufio.NewScanner(os.Stdin)
	//for {
	//	fmt.Print("soup> ")
	//
	//	if !scanner.Scan() {
	//		break
	//	}
	//
	//	line := scanner.Text()
	//	if line == "exit" {
	//		break
	//	}
	//	fmt.Println(line)
	//
	//	//l := lexer.NewString(input)
	//	//p := parser.New(l)
	//	//
	//	//program, err := p.Parse()
	//	//if err != nil {
	//	//	return err
	//	//}
	//	//
	//	//ev := evaluator.New()
	//	//result, err := ev.Eval(program)
	//	//if err != nil {
	//	//	return err
	//	//}
	//	//printReturnValue(result)
	//}
	return nil
}

func runFile(fileName string) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	l := lexer.New(file)

	p := parser.New(l)

	program, err := p.Parse()
	if err != nil {
		return err
	}

	ev := evaluator.New()
	result, err := ev.Eval(program)
	if err != nil {
		return err
	}
	printReturnValue(result)

	return nil
}

func printReturnValue(ret *evaluator.ReturnValue) {
	fmt.Printf("Result: %s\n", ret.String())
}
