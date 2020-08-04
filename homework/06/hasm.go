/*
Package Test implements the Hack machine
Compiler
*/
package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
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

func (c *Compiler) Write() {
	//write to another file
	outname := c.filename[:len(c.filename)-3] + "hack"
	log.Println("write to file:", outname)

	fo, err := os.Create(outname)
	defer fo.Close()
	if err != nil {
		log.Fatalln("Compiler failed to create file:", err)
	}
	fWriter := bufio.NewWriter(fo)

	for i, line := range c.stream {
		line = "line " + strconv.Itoa(i) + ":" + line + "\n"
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
	compiler.Write()

	log.Println("Compiler finished success!")
}
