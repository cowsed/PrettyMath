package opencl_renderer

//Helper file for the OpenCL Workspace

import (
	"log"
	//	"errors"
	"fmt"
	"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
	re "regexp"
	"strings"
)

//Type interface

type ClDataHolder interface{
	Build(ws *Workspace) 
	SetArg(index int, k *cl.Kernel, ws *Workspace) error
}
//Float input
type CLFloatInput struct{
	value float32
	name string
	typeName string
}

func (dh *CLFloatInput) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	if imgui.DragFloatV(dh.name, &dh.value, 0.001, -1000,1000,"%.4f",1){
		ws.Run()
	}
}
func (dh *CLFloatInput) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err:=k.SetArgFloat32(index, dh.value)
	return err
}
//Uint32 input
type CLUint32Input struct{
	value int32
	name string
	typeName string
}
func (dh *CLUint32Input) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	if imgui.DragInt(dh.name, &dh.value) {
		ws.Run()
	}
}
func (dh *CLUint32Input) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err:=k.SetArgUint32(index, uint32(dh.value))
	return err
}

//Image buffer input
type CLImageInput struct{
	identifier string
	index int
	name string
	typeName string
}
func (dh *CLImageInput) Build(ws *Workspace) {
	//Maybe add tooltip that can be used to control secondary parameters
	if imgui.InputText(dh.name, &dh.identifier) {
	ws.Run()
	}
}
func (dh *CLImageInput) SetArg(index int, k *cl.Kernel, ws *Workspace) error {
	err:=k.SetArgBuffer(index, ws.imageBuffers[dh.index])
	return err
}



//findTypes extracts the types from a kernel function definition that matches kernel name
func findNamesAndTypes(kernelName, kernelSource string) ([]string, []string, error) {
	var argFinder = re.MustCompile(`__kernel void `+kernelName+`\(([a-zA-Z_\s0-9,])*`)
	inds:=argFinder.FindAllIndex([]byte(kernelSource), -1)
	if len(inds)==0{
	return nil,nil,fmt.Errorf("No Parameters found. Thats weird")
	}
	header:=kernelSource[inds[0][0]+len("__kernel void "+kernelName+"("):inds[0][1]]

	types:= strings.Split(header, ",")
	names:=make([]string, len(types))
	for i,t :=range types{
		fmt.Printf("%dth type: %s \n", i, t)
		if t=="\n"{
			//continue if an empty line
			continue
		}
		//Get all elements that arent the last (name) and rejoin them
		types[i]=strings.ReplaceAll(t,"\n","")
		parts:=strings.Split(types[i], " ")
		types[i]=strings.Join(parts[:len(parts)-1]," ")
		names[i] = parts[len(parts)-1]
	}
	
	return names, types, nil
	
}

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
