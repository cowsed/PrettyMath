package tools

import (
	"fmt"
	"image"
	"os"
	"os/exec"
	"image/png"
)

//Copies an image to the clipboard
//only works on my machine
func CopyImageRGBA(img *image.RGBA) {
	fname := "/tmp/out.png"
	f,_ :=os.Create(fname)
	png.Encode(f,img)

	arg0:="-selection"
	arg1:="clipboard "
	arg2:="-target"
	arg3:="image/png"
	arg4:="-i"
	
	
	app := "xclip"
	

	cmd := exec.Command(app,arg0,arg1,arg2,arg3,arg4,fname)
	fmt.Println("Commanded")
	
	cmd.Run()
	
	stdout, err := cmd.Output()
	fmt.Println("Otputted")

	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(stdout)
		return
	}

}
