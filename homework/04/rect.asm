//index = 0
@index
M=0
@SCREEN
D=A
@addr
M=D

(LOOP)
//if ram[0] - index <=0
//  jump END
@0
D=M
@index
D=D-M
@END
D;JLE

//ram[addr]=-1
@addr
A=M
M=-1
//addr += 32
@32
D=A
@addr
M=M+D
//index++
@index
M=M+1

////jump loop
@LOOP

0;JMP
//end
(END)
@END
0;JMP
