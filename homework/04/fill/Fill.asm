// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/04/Fill.asm

// Runs an infinite loop that listens to the keyboard input.
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel;
// the screen should remain fully black as long as the key is pressed. 
// When no key is pressed, the program clears the screen, i.e. writes
// "white" in every pixel;
// the screen should remain fully clear as long as no key is pressed.

// Put your code here.
(LOOP)
@i
M=0
//flag=0, white, -1, black
@flag
M=0
//addr=screen
@SCREEN
D=A
@addr
M=D

//listen to keyboard
@KBD
D=M
@BLACK
D;JNE
//set flag to white
@flag
M=0
@FILL
0;JMP
//black flag
(BLACK)
@flag
M=-1

//fill screen
(FILL)
@flag
D=M
@addr
A=M
M=D
//addr++
@addr
M=M+1
@8192
D=A
//i++
@i
M=M+1
D=D-M
//finish screen buffer
@LOOP
D;JLE
@FILL
0;JMP

