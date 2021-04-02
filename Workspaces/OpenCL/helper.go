package opencl_renderer

//Helper file for the OpenCL Workspace

import (
	"log"
	//	"errors"
	"fmt"
	"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
	re "regexp"
	//"strings"
)
var lineFinder = re.MustCompile(`\d+:\d+:\d+: (note: |error: )`)


func parseCLErrors(errs string) imgui.ErrorMarkers {
	errMarkers := imgui.NewErrorMarkers()

	if len(errs) == 0 {
		return errMarkers
	}


	if errs[:17] == "cl: build error (" {
		errs = errs[17:]
	} else {
		return errMarkers
	}
	errorInds:=lineFinder.FindAllIndex([]byte(errs),-1)
	for i, is:= range(errorInds){
		var something, line, col int
		var str string
		_, err:=fmt.Sscanf(errs[is[0]:is[1]], "%d:%d:%d: %s", &something, &line, &col, &str)
		if err!=nil{
			return errMarkers
		}
		second:=0
		if i==len(errorInds)-1{
			second=len(errs)
		} else {
			second=errorInds[i+1][0]
		}
		message:=errs[is[1]:second]
		
		errMarkers.Insert(line,message)
	}
	
	

	return errMarkers
}


func makeKernel(kernelName string, kernelSource string, context *cl.Context) *cl.Kernel {
	log.Println("Creating Kernel", kernelName)
	program, err := context.CreateProgramWithSource([]string{kernelSource})
	check(err)
	err = program.BuildProgram(nil, "")
	check(err)
	kernel, err := program.CreateKernel(kernelName)
	check(err)

	numArgs, err := kernel.NumArgs()
	check(err)
	log.Printf("%v arguements detected", numArgs)

	for i := 0; i < numArgs; i++ {
		name, err := kernel.ArgName(i)
		if err == cl.ErrUnsupported {
			break
		} else if err != nil {
			log.Printf("GetKernelArgInfo for name:%v failed: %+v", name, err)
			break
		} else {
			log.Printf("Kernel arg %d: %s \n", i, name)
		}
	}

	return kernel
}

func makeCLContext() (*cl.Platform, *cl.Device, *cl.Context, *cl.CommandQueue) {
	//Load Platform
	platforms, err := cl.GetPlatforms()
	check(err)
	platform := platforms[0]
	log.Printf("Loaded Platform: %v\n", platform.Name())

	//Setup opencl
	devices, err := platform.GetDevices(cl.DeviceTypeGPU)
	check(err)
	device := devices[0]
	log.Printf("Loaded Device: %v", device.Name())

	context, err := cl.CreateContext([]*cl.Device{device})
	check(err)

	queue, err := context.CreateCommandQueue(device, 0)
	check(err)
	return platform, device, context, queue
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
