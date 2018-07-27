package main

import (
	"github.com/freddy33/qsm-go/space_gl"
	"os"
	"github.com/freddy33/qsm-go/space_nk"
	"github.com/freddy33/qsm-go/space_glot"
	"fmt"
	"runtime"
	"github.com/KyleBanks/conways-gol/cgol"
)

func main() {
	runtime.LockOSThread()
	c := "gl_cube"
	if len(os.Args) > 1 {
		c = os.Args[1]
	}
	fmt.Println("Executing", c)
	switch c {
	case "nk":
		space_nk.DisplayNkScatter()
	case "gl_cube":
		space_gl.DisplayCube()
	case "glot_ex":
		space_glot.DisplayExample3D()
	case "glot_cube":
		space_glot.DisplayCube()
	case "cgol":
		cgol.RunConwayGOL()
	default:
		fmt.Println("The param",c,"unknown")
	}
	fmt.Println("Finished Executing", c)
}
