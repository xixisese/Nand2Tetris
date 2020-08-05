/*
Package Test implements the Hack machine
Compiler
*/
package main

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	varBase = 16
)

var symbolTable = map[string]int{
	//build in symbol
	"R0":     0,
	"R1":     1,
	"R2":     2,
	"R3":     3,
	"R4":     4,
	"R5":     5,
	"R6":     6,
	"R7":     7,
	"R8":     8,
	"R9":     9,
	"R10":    10,
	"R11":    11,
	"R12":    12,
	"R13":    13,
	"R14":    14,
	"R15":    15,
	"SP":     0,
	"LCL":    1,
	"ARG":    2,
	"THIS":   3,
	"THAT":   4,
	"SCREEN": 16384,
	"KBD":    24576,
}

type Compiler struct {
	filename string
	stream   []string
}

func newCompiler(name string) *Compiler {
	return &Compiler{filename: name}
}

func (c *Compiler) Open() {
	file, err := os.Open(c.filename)
	defer file.Close()
	if err != nil {
		log.Fatalln("failed to open file:", err)
	}
	log.Println("file open success:", c.filename)

	//read file lines
	fScanner := bufio.NewScanner(file)
	fScanner.Split(bufio.ScanLines)

	for fScanner.Scan() {
		c.stream = append(c.stream, fScanner.Text())
	}
}
func (c *Compiler) newFile(suffix string) *os.File {
	pathIndex := strings.LastIndexAny(c.filename, "/\\")
	path := "."
	if pathIndex != -1 {
		path = c.filename[:pathIndex]
	}
	buildPath := path + "/build/"
	buildFile := c.filename[pathIndex+1:]
	outname := buildPath + buildFile[:len(buildFile)-3] + suffix
	log.Println("Generate new file:", outname)

	if err := os.MkdirAll(filepath.Dir(outname), 0770); err != nil {
		log.Fatalln("Compiler failed to create file directory", err)
		return nil
	}
	f, err := os.Create(outname)
	if err != nil {
		log.Fatalln("Compiler failed to create file", err)
	}

	return f
}
func (c *Compiler) Write() {
	//write to another file
	file := c.newFile("hack")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()
}

//PreCompile remove the empty lines and comments, generating .pre file
func (c *Compiler) PreCompile() {
	//open file
	file := c.newFile("pre.1")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	var newStream []string

	//line loop
	for _, line := range c.stream {
		line := strings.TrimSpace(line)

		//ignore empty lines, or comment line with "//" prefix
		if (line == "") || strings.HasPrefix(line, "//") {
			continue
		}

		//remove comment within the lines
		if i := strings.Index(line, "//"); i != -1 {
			line = strings.TrimSpace(line[:i])
		}

		newStream = append(newStream, line)
	}

	//replace old stream
	c.stream = newStream

	//flush to output
	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()
}

//GenSymbolTable generate the symbol table .symbol from the .pre file
func (c *Compiler) GenSymbolTable() {
	//open file
	symbolFile := c.newFile("symbol.2")
	defer symbolFile.Close()
	symbolFileWriter := bufio.NewWriter(symbolFile)

	file := c.newFile("noLabel.2")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	var newStream []string
	lineIndex := 0
	varIndex := 0
	//hadle label defination
	for _, line := range c.stream {
		if strings.HasPrefix(line, "(") && strings.HasSuffix(line, ")") {
			line = strings.TrimSpace(line[1 : len(line)-1])
			symbolTable[line] = lineIndex + 1
			continue
		}
		lineIndex++
		newStream = append(newStream, line)
	}

	//handle variable
	for _, line := range newStream {
		if !strings.HasPrefix(line, "@") {
			continue
		}

		v := line[1:]
		if _, err := strconv.ParseInt(v, 10, 32); err == nil {
			continue
		}

		if _, ok := symbolTable[v]; !ok {
			symbolTable[v] = varBase + varIndex
			varIndex++
			log.Println("Add new variable:", v)
		}
	}

	c.stream = newStream

	//flush to output
	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()

	//fush symbol file
	for k, v := range symbolTable {
		l := k + ":" + strconv.Itoa(v) + "\n"
		if _, err := symbolFileWriter.WriteString(l); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	symbolFileWriter.Flush()
}

//ReplaceSymbol replace the symbols by value and generate .nosymbol file
func (c *Compiler) ReplaceSymbol() {
	//open file
	file := c.newFile("nosymbol.3")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	var newStream []string

	//line loop
	for _, line := range c.stream {
		//ignore the C instructions
		if !strings.HasPrefix(line, "@") {
			newStream = append(newStream, line)
			continue
		}
		k := line[1:]

		//ignore the ones already a number
		if _, err := strconv.ParseInt(k, 10, 32); err == nil {
			newStream = append(newStream, line)
			continue
		}

		//replace the symbol
		v, ok := symbolTable[k]
		if !ok {
			log.Fatalln("Unrecognized symbol:", k)
			return
		}

		line = "@" + strconv.Itoa(v)

		newStream = append(newStream, line)
	}

	//replace old stream
	c.stream = newStream

	//flush to output
	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()
}
func (c *Compiler) ParseSyntax() {
	//open file
	file := c.newFile("syntax.4")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	var newStream []string

	//line loop
	for _, line := range c.stream {
		line := strings.TrimSpace(line)
		newStream = append(newStream, line)
	}

	//replace old stream
	c.stream = newStream

	//flush to output
	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()
}
func (c *Compiler) Assemble() {
	//open file
	file := c.newFile("hack")
	defer file.Close()
	fWriter := bufio.NewWriter(file)

	var newStream []string

	//line loop
	for _, line := range c.stream {
		line := strings.TrimSpace(line)
		newStream = append(newStream, line)
	}

	//replace old stream
	c.stream = newStream

	//flush to output
	for _, line := range c.stream {
		line = line + "\n"
		if _, err := fWriter.WriteString(line); err != nil {
			log.Fatalln("Compiler failed to write file:", err)
			break
		}
	}
	fWriter.Flush()
}

/*
	main operation
*/
func main() {
	//parameter check
	if len(os.Args) != 2 {
		log.Fatalln("Please enter the filename as parameter")
		return
	}

	filename := os.Args[1]
	compiler := newCompiler(filename)
	compiler.Open()
	compiler.PreCompile()
	compiler.GenSymbolTable()
	compiler.ReplaceSymbol()
	compiler.ParseSyntax()
	compiler.Assemble()

	log.Println("Compiler finished success!")
}
