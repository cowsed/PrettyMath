package opencl_renderer

//Eventually have a pipeline where each has a program
//Each program has a parameter section that math the opencl program or else causes an error
//images are defined by their name and the same names are used on the same ReadWriteBuffer
//Have a global section where you can define all buffers (and maybe shared parameters)
//For per pipe parameter section have a way of showing it inherits from the global parameters via the opencl paramater name
//For those that dont match have the corresponding input variety from imgui

import (
	"fmt"
	"image"
	//"image/color"

	"log"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
)

var baseProgram = `__kernel void blur(
  __write_only image2d_t image,
  const unsigned int Width,
  const unsigned int Height
) {

  int id = get_global_id(0);
  int idx = id % Width;
  int idy = id / Width;

  bool mask=(((idx/16)%2) +  ((idy/16)%2))%2==0;
  
  float4 col = (float4)(1,1,1,1)*mask;
  write_imagef(image, (int2)(idx,idy), col);

}
`

type Workspace struct {
	amOpen  bool
	onClose func()

	selfOnCloses []func() //Functions to run when the workspace is closed
	onFinish     []func() //Releases/Mem Management to run when the rendering of frames is over

	programSource string
	programName   string

	editor     imgui.TextEditor
	errMarkers imgui.ErrorMarkers

	//Output Image Stuff
	width, height int32
	outputTex     *giu.Texture

	//CL Stuff
	contextCL *cl.Context
	queueCL   *cl.CommandQueue
	deviceCL  *cl.Device
	kernelCL  *cl.Kernel
	errorsCL  string

	//Parameters
	imageBuffers []*cl.MemObject
	images       []*image.RGBA
}

//Init initializes a new OpenCL workspace
func Init(onCloseFunc func()) Workspace {

	ws := Workspace{
		amOpen:        true,
		onClose:       onCloseFunc,
		programSource: baseProgram,
		programName:   "blur",
		width:         400,
		height:        300,
	}
	_, ws.deviceCL, ws.contextCL, ws.queueCL = makeCLContext()
	ws.selfOnCloses = []func(){ws.contextCL.Release, ws.queueCL.Release}

	//Setup Text Editor
	ws.errMarkers = imgui.NewErrorMarkers()

	ws.editor = imgui.NewTextEditor()
	ws.editor.SetErrorMarkers(ws.errMarkers)

	ws.editor.SetLanguageDefinitionC()
	ws.editor.SetText(ws.programSource)
	ws.editor.SetShowWhitespaces(false)
	ws.editor.SetTabSize(2)

	return ws
}

func (ws *Workspace) BuildProgram() (*cl.Kernel, error) {
	//Reset errors
	ws.errorsCL = ""
	ws.errMarkers = imgui.NewErrorMarkers()
	ws.editor.SetErrorMarkers(ws.errMarkers)

	program, err := ws.contextCL.CreateProgramWithSource([]string{ws.programSource})
	if err != nil {
		fmt.Println("Failling hard")
		return nil, fmt.Errorf("Failed Building Program somehow")
	}
	fmt.Println("Succeed hard")

	err = program.BuildProgram(nil, "")
	if err != nil {
		fmt.Println("Failed toBuild program")
		ws.errorsCL = err.Error()

		ws.errMarkers = parseCLErrors(ws.errorsCL)
		ws.editor.SetErrorMarkers(ws.errMarkers)
		return nil, err
	}
	fmt.Println("Built program")

	kernel, err := program.CreateKernel(ws.programName)
	check(err)
	if err != nil {
		fmt.Println("Kernel Error", err.Error())
	}
	numArgs, err := kernel.NumArgs()
	check(err)
	log.Printf("%v arguements detected", numArgs)

	var name string
	for i := 0; i < numArgs; i++ {
		name, err = kernel.ArgName(i)
		if err == cl.ErrUnsupported {
			break
		} else if err != nil {
			log.Printf("GetKernelArgInfo for name:%v failed: %+v", name, err)
			break
		} else {
			log.Printf("Kernel arg %d: %s \n", i, name)
		}
	}
	if err != nil {
		return nil, err
	}
	return kernel, nil

}

func (ws *Workspace) PrepareBuffers() {
	rect := image.Rectangle{image.Point{0, 0}, image.Point{int(ws.width), int(ws.height)}}
	image1 := image.NewRGBA(rect)
	fmt.Println("Image Length", len(image1.Pix))
	ws.images = append(ws.images, image1)

	image1Buffer, err := ws.contextCL.CreateImageSimple(cl.MemReadWrite|cl.MemUseHostPtr, int(ws.width), int(ws.height), cl.ChannelOrderRGBA, cl.ChannelDataTypeUNormInt8, image1.Pix)
	ws.onFinish = append(ws.onFinish, image1Buffer.Release)
	check(err)

	ws.imageBuffers = append(ws.imageBuffers, image1Buffer)
}

func (ws *Workspace) Run() {

	ws.programSource = ws.editor.GetText()
	fmt.Println(ws.programSource)
	var err error
	ws.kernelCL, err = ws.BuildProgram()
	if err != nil {
		ws.releaseOnFinish()
		return
	}
	//Reset then prepare buffers
	ws.imageBuffers = []*cl.MemObject{}
	ws.images = []*image.RGBA{}

	ws.PrepareBuffers()

	err = ws.kernelCL.SetArgs(ws.imageBuffers[0], ws.width, ws.height)
	if err != nil {
		ws.releaseOnFinish()
		fmt.Println("imageBuffers", len(ws.imageBuffers), "[0]", ws.imageBuffers[0])
		fmt.Println("FailedArg set: ", err.Error())
		return
	}

	local, err := ws.kernelCL.WorkGroupSize(ws.deviceCL)
	check(err)

	global := int(ws.width * ws.height)
	d := int(ws.width) * int(ws.height) % local
	if d != 0 {
		global += local - d
	}

	e, err := ws.queueCL.EnqueueNDRangeKernel(ws.kernelCL, nil, []int{global}, []int{local}, nil)
	e.Release()

	//Read from buffer
	e, err = ws.queueCL.EnqueueReadImage(ws.imageBuffers[0], true, [3]int{0, 0, 0}, [3]int{int(ws.width), int(ws.height), 1}, int(ws.width)*4, 0, ws.images[0].Pix, nil)
	check(err)
	e.Release()

	go func() {

		ws.outputTex, err = giu.NewTextureFromRgba(ws.images[0])
		fmt.Println("Finished making tex")
		check(err)
	}()

	ws.releaseOnFinish()
	giu.Update()
}

func (ws *Workspace) releaseOnFinish() {
	//Release all necessary things
	for _, f := range ws.onFinish {
		f()
	}
}

func (ws *Workspace) Build() {
	if !ws.amOpen {
		//Release Everything
		ws.onClose()
		for _, f := range ws.selfOnCloses {
			fmt.Println("Releasing")
			f()
		}
	}


	giu.TabItem("OpenCL Pipeline").Layout(
		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 700,
			giu.Group().Layout(
				giu.Button("Run").OnClick(ws.Run),
				giu.Custom(func() {
					ws.editor.Render("OpenCl", imgui.Vec2{0, 0}, true)
				},
				),
			),
			giu.Custom(func() {
				size := imgui.ContentRegionAvail()
				aspectRatio := float32(ws.width) / float32(ws.height)
				fmt.Println("newSize", size.X, size.X/aspectRatio)
				giu.Group().Layout(
					giu.Image(ws.outputTex).Size(size.X, size.X/aspectRatio),
					giu.Label("Its possible your image is just transparent"),
				).Build()
			}),
		),
	).IsOpen(&ws.amOpen).Build()
}
