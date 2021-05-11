package opencl_renderer

//Eventually have a pipeline where each has a program
//Each program has a parameter section that math the opencl program or else causes an error
//images are defined by their name and the same names are used on the same ReadWriteBuffer
//Have a global section where you can define all buffers (and maybe shared parameters)
//For per pipe parameter section have a way of showing it inherits from the global parameters via the opencl paramater name
//For those that dont match have the corresponding input variety from imgui
//For any buffer if there is any parameter that matches something like parameter_(something), put something in the section that requires a rebuild
//(Variables that are declared to be of type image2d_t or image3d_t refer to image memory objects. These can only be specified as arguments to a function. Elements of an image cannot be directly accessed.)
import (
	"fmt"
	"image"
	"image/png"
	//"log"
	"os"
	//"strconv"
	//"strings"

	"github.com/AllenDang/giu"
	//"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
)

var baseProgram = loadFile("Workspaces/OpenCL/mandelbrot.cl")

type Workspace struct {
	amOpen  bool
	onClose func()

	selfOnCloses []func() //Functions to run when the workspace is closed
	onFinish     []func() //Releases/Mem Management to run when the rendering of frames is over

	programs        []*CLProgram
	programsCurrent bool

	//Output Image Stuff
	desiredFrames int32
	currentFrame int
	width, height int32
	outputTex     *giu.Texture

	//CL Stuff
	contextCL      *cl.Context
	queueCL        *cl.CommandQueue
	deviceCL       *cl.Device
	errorsCL       string
	buffersCurrent bool

	args []ClDataHolder

	//Parameters
	imageBuffers []*cl.MemObject
	images       []*image.RGBA
}

//Init initializes a new OpenCL workspace
func Init(onCloseFunc func()) Workspace {
	prog := CLProgram{
		programSource: baseProgram,
		programName:   "fractal",
	}
	ws := Workspace{
		amOpen:   true,
		onClose:  onCloseFunc,
		programs: []*CLProgram{&prog},
		width:    800,
		height:   600,
		args:     []ClDataHolder{},
	}
	_, ws.deviceCL, ws.contextCL, ws.queueCL = makeCLContext()
	ws.selfOnCloses = []func(){ws.contextCL.Release, ws.queueCL.Release}

	ws.PrepareBuffers()
	ws.buffersCurrent = true
	fmt.Println("image buffers", ws.imageBuffers)
	fmt.Println("images", ws.images)

	for i := range ws.programs {
		//ws.programs[i].ws = &ws
		ws.programs[i].initEditor(&ws)
		ws.programs[i].makeParameters(&ws)
	}

	return ws
}

func (ws *Workspace) makeAnim(){
	for ws.currentFrame=0; ws.currentFrame<int(ws.desiredFrames); ws.currentFrame++{
		fmt.Println("Making new#",	ws.currentFrame)
		ws.Run()
		fname:=fmt.Sprintf("Workspaces/OpenCL/Frames/Frame%04d",ws.currentFrame)
		f,_:=os.Create(fname)
		png.Encode(f,ws.images[0])
		giu.Update()
	}
}

func (ws *Workspace) PrepareBuffers() {

	fmt.Println("Preparing buffers")
	rect := image.Rectangle{image.Point{0, 0}, image.Point{int(ws.width), int(ws.height)}}
	image1 := image.NewRGBA(rect)
	ws.images = append(ws.images, image1)

	image1Buffer, err := ws.contextCL.CreateImageSimple(cl.MemReadWrite|cl.MemUseHostPtr, int(ws.width), int(ws.height), cl.ChannelOrderRGBA, cl.ChannelDataTypeUNormInt8, image1.Pix)
	ws.selfOnCloses = append(ws.selfOnCloses, image1Buffer.Release)
	check(err)

	ws.imageBuffers = append(ws.imageBuffers, image1Buffer)
}

func (ws *Workspace) BuildPrograms() {
	ws.programsCurrent = true
	for i := range ws.programs {
		ws.programs[i].BuildProgram(ws)
		if ws.programs[i].current == false {
			//There was probably an error
			ws.programsCurrent = false
			return
		}
	}
}
func (ws *Workspace) checkPrograms() {
	//ws.programsCurrent = true
	for i := range ws.programs {
		if ws.programs[i].current == false {
			fmt.Println("Found not current program")
			ws.programsCurrent = false
		}
	}
}

func (ws *Workspace) Run() {
	fmt.Println("Running\n\n\n====")
	kernelsReady := true
	//ws.programsCurrent = true

	for i := range ws.programs {
		fmt.Println("Pre Program is ",&(ws.programs[i]))
		if ws.programs[i].kernelCL == nil {
			kernelsReady = false
		}
		if ws.programs[i].current == false {
			ws.programsCurrent = false
		}

	}

	if ws.queueCL == nil || !kernelsReady {
		fmt.Println("Program(s) are not built/initialized")
		return
	}
	if !ws.buffersCurrent {
		fmt.Println("Preparing Buffers")
		//Reset then prepare buffers
		ws.imageBuffers = []*cl.MemObject{}
		ws.images = []*image.RGBA{}

		ws.PrepareBuffers()
		ws.buffersCurrent = true

	}

	for i := range ws.programs {
		ws.programs[i].setArgs(ws)
		global, local := ws.programs[i].getWorkGroupSizes(ws)
		e, err := ws.queueCL.EnqueueNDRangeKernel(ws.programs[i].kernelCL, nil, global, local, nil)
		e.Release()
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	//Read from buffer
	e, err := ws.queueCL.EnqueueReadImage(ws.imageBuffers[0], true, [3]int{0, 0, 0}, [3]int{int(ws.width), int(ws.height), 1}, int(ws.width)*4, 0, ws.images[0].Pix, nil)
	check(err)
	e.Release()

	fmt.Println("Images At end:len:", len(ws.images))
	fmt.Println("Images ", ws.images[0].Pix[:10])
	go func() {
		before := ws.outputTex

		ws.outputTex, err = giu.NewTextureFromRgba(ws.images[0])
		fmt.Println("Made tex. err:", err)
		fmt.Println("Before", before, "After:", ws.outputTex)
		check(err)
		giu.Update()
	}()

	ws.releaseOnFinish()
	giu.Update()
}

//releases appropriate cl memory objects to avoid memory leaks
func (ws *Workspace) releaseOnFinish() {
	//Release all necessary things
	for _, f := range ws.onFinish {
		f()
	}
}

func (ws *Workspace) Save() {
	f, _ := os.Create("out.png")
	png.Encode(f, ws.images[0])
}
