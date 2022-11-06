package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	"run"
)

func main() {

	if len(os.Args) != 2 {
		color.HiRed("usage:")
		fmt.Println("  exe 1234")
	}
	num_str := os.Args[1]
	num, e := strconv.Atoi(num_str)
	if e != nil {
		panic(e)
	}

	{

		color.HiMagenta("decimal:")
		stdout, _, _ := run.RunCommand("/root/Downloads/src", "ag", "--color", fmt.Sprintf("\\(%d,", num))
		fmt.Println(stdout)
	}
	{
		color.HiMagenta("hex:")
		stdout, _, _ := run.RunCommand("/root/Downloads/src", "ag", "--color", fmt.Sprintf("\\(0x%x,", num))
		fmt.Println(stdout)
	}
}
