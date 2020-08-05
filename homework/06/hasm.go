/*
Package Test implements the Hack machine
Compiler
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	varBase = 16

	AIns = "0"
	CIns = "1"
	FS   = ","
)

var (
	symbolTable = map[string]int{
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

	cmpTable = map[string]string{
		"0":   "0101010",
		"1":   "0111111",
		"-1":  "0111010",
		"D":   "0001100",
		"A":   "0110000",
		"M":   "1110000",
		"!D":  "0001101",
		"!A":  "0110001",
		"!M":  "1110001",
		"-D":  "0001111",
		"-A":  "0110011",
		"-M":  "1110011",
		"D+1": "0011111",
		"1+D": "0011111",
		"A+1": "0110111",
		"1+A": "0110111",
		"M+1": "1110111",
		"1+M": "1110111",
		"D-1": "0001110",
		"A-1": "0110010",
		"M-1": "1110010",
		"D+A": "0000010",
		"A+D": "0000010",
		"D+M": "1000010",
		"M+D": "1000010",
		"D-A": "0010011",
		"D-M": "1010011",
		"A-D": "0000111",
		"M-D": "1000111",
		"D&A": "0000000",
		"A&D": "0000000",
		"D&M": "1000000",
		"M&D": "1000000",
		"D|A": "0010101",
		"A|D": "0010101",
		"D|M": "1010101",
		"M|D": "1010101",
	}

	desTable = map[string]string{
		"null": "000",
		"M":    "001",
		"D":    "010",
		"MD":   "011",
		"A":    "100",
		"AM":   "101",
		"AD":   "110",
		"AMD":  "111",
	}

	jmpTable = map[string]string{
		"null": "000",
		"JGT":  "001",
		"JEQ":  "010",
		"JGE":  "011",
		"JLT":  "100",
		"JNE":  "101",
		"JLE":  "110",
		"JMP":  "111",
	}
)

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
			symbolTable[line] = lineIndex
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
		if strings.HasPrefix(line, "@") {
			//output the parsed A instruction as 0, data
			line = AIns + FS + line[1:]
			newStream = append(newStream, line)
			continue
		}

		//output the parsed C instruction as 1, des, comp, jmp
		var des, comp, jmp string
		//get des fields
		ret := strings.Split(line, "=")

		if len(ret) > 1 {
			des = ret[0]
			ret = ret[1:]
		}
		ret = strings.Split(ret[0], ";")
		comp = ret[0]
		if len(ret) > 1 {
			jmp = ret[1]
		}

		line = CIns + FS + des + FS + comp + FS + jmp
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
		a := strings.Split(line, FS)
		//hadle A instruction
		if a[0] == AIns {
			v, err := strconv.ParseInt(a[1], 10, 32)
			if err != nil {
				log.Fatalln("Unrecognized integer value in A instruction:", a[1], a)
			}
			line = fmt.Sprintf("%s%015b", AIns, v)
		} else {
			//handle C Instruction
			des := a[1]
			cmp := a[2]
			jmp := a[3]

			line = CIns + "11"

			//transalte cmp
			if cmp == "" {
				log.Fatalln("Des field is empty")
			}
			v, ok := cmpTable[cmp]
			if !ok {
				log.Fatalln("Unrecognized cmp symbol:", cmp)
			}
			line += v

			//translate des
			if des == "" {
				des = "null"
			}

			v, ok = desTable[des]
			if !ok {
				log.Fatalln("Unrecognized des symbol:", des)
			}

			line += v

			//translate jmp
			if jmp == "" {
				jmp = "null"
			}
			v, ok = jmpTable[jmp]
			if !ok {
				log.Fatalln("Unrecognized jmp symbol:", jmp)
			}
			line += v
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
