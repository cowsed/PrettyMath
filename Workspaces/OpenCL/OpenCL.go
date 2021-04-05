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
	"log"
	"os"
	//"strconv"
	"strings"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/jgillich/go-opencl/cl"
	"io/ioutil"
)

func loadFile(fname string) string {
	content, err := ioutil.ReadFile(fname)
	check(err)
	return string(content)
}

var baseProgram = loadFile("Workspaces/OpenCL/mandelbrot.cl")

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

	//Parameters
	pos   [2]float32
	scale float32

	//CL Stuff
	contextCL      *cl.Context
	queueCL        *cl.CommandQueue
	deviceCL       *cl.Device
	kernelCL       *cl.Kernel
	errorsCL       string
	programCurrent bool
	buffersCurrent bool

	args []ClDataHolder

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
		width:         800,
		height:        600,
		args:          []ClDataHolder{},
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

	ws.PrepareBuffers()
	ws.buffersCurrent = true

	ws.makeParameters()

	return ws
}

func (ws *Workspace) makeParameters() {
	names, types, err := findNamesAndTypes(ws.programName, ws.programSource)
	if err != nil {
		return
	}
	if len(names) != len(types) {
		return
	}
	oldArgs := ws.args[:]
	ws.args = make([]ClDataHolder, 0)
	bufferI := 0
	for i, t := range types {
		//Clean up stuff
		parts := strings.Split(strings.TrimSpace(t), " ")
		fmt.Println("Parts: ", parts)
		qualifier := parts[0]
		fmt.Println("qualified as", qualifier)
		actualType := strings.Join(parts[1:], " ")
		actualType = strings.TrimSpace(actualType)

		oldArgAvailable := i < len(oldArgs)

		switch actualType {
		case "unsigned int":
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLUint32Input{0, names[i], 0,4,actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)
		case "int":
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLInt32Input{0, names[i], 0,4,actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)

		case "float":
			fmt.Println("Found: const float")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatInput{0,-100,100,0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)
		case "float3":
			fmt.Println("Found: const float3")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatVecInput{[]float32{0.0,0.0,0.0},3,-100,100,0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)
		case "float2":
			fmt.Println("Found: const float2")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatVecInput{[]float32{0.0,0.0},2,-100,100,0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)

		case "image2d_t":
			log.Println("UNDEFINED BEHAVIOUR sorta")
			//If there are not enough buffers
			if bufferI >= len(ws.imageBuffers) {
				fmt.Println(" p r o b le m a t ic")
				return
				//return nil, fmt.Errorf("Mismath between desired and found buffers")
			}
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLImageInput{"imageThing", bufferI, names[i], false, actualType}
			}
			newArg.setName(names[i])
			ws.args = append(ws.args, newArg)
			bufferI++

		}

	}
	fmt.Println("there are", len(ws.args), "args made")
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

func (ws *Workspace) BuildProgram() {
	ws.programSource = ws.editor.GetText()
	//Reset errors
	ws.errorsCL = ""
	ws.errMarkers = imgui.NewErrorMarkers()
	ws.editor.SetErrorMarkers(ws.errMarkers)

	program, err := ws.contextCL.CreateProgramWithSource([]string{ws.programSource})
	if err != nil {
		fmt.Println("Failling hard")
		fmt.Errorf("Failed Building Program somehow err: %v", err.Error())
	}
	fmt.Println("Succeed hard")

	err = program.BuildProgram(nil, "")
	if err != nil {
		fmt.Println("Failed toBuild program")
		ws.errorsCL = err.Error()

		ws.errMarkers = parseCLErrors(ws.errorsCL)
		ws.editor.SetErrorMarkers(ws.errMarkers)
		ws.releaseOnFinish()

		log.Println(err.Error())
		return
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
		ws.releaseOnFinish()
		log.Fatalf(err.Error())
		return
	}

	ws.kernelCL = kernel
	ws.programCurrent=true

}

func (ws *Workspace) Run() {
	if ws.queueCL == nil || ws.kernelCL == nil {
		fmt.Println("Program is not built/initialized")
		return
	}
	if !ws.buffersCurrent {
		//Reset then prepare buffers
		ws.imageBuffers = []*cl.MemObject{}
		ws.images = []*image.RGBA{}

		ws.PrepareBuffers()
		ws.buffersCurrent = true

	}

	for i, d := range ws.args {
		err := d.SetArg(i, ws.kernelCL, ws)
		if err != nil {
			log.Println(i, "th arguement error", err.Error())
			ws.releaseOnFinish()
			break
		}
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
		check(err)
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

var data = []float32{1.0,2.0,3.0}
func (ws *Workspace) Build() {
	if !ws.amOpen {
		//Release Everything
		ws.onClose()
		for _, f := range ws.selfOnCloses {
			fmt.Println("Releasing")
			f()
		}
		ws.releaseOnFinish()
	}
	
	buildStatus:=" "
	if !ws.programCurrent{
		buildStatus="*"
	}
	
	giu.TabItem("OpenCL Pipeline").Layout(

		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 700,
			giu.Group().Layout(
				giu.Line(
					giu.Button("Build"+buildStatus).OnClick(ws.BuildProgram),
					giu.Button("Run").OnClick(ws.Run),
					giu.Button("Save Image").OnClick(ws.Save),
				),
				giu.Custom(ws.buildParameterInputs),
				giu.Custom(func() {
					ws.editor.Render("OpenCl", imgui.Vec2{0, 0}, true)
					if ws.editor.IsTextChanged() {
						ws.programSource = ws.editor.GetText()
						ws.programCurrent=false

						fmt.Println("Update Parameters")
						ws.makeParameters()
					}
				},
				),
			),
			giu.Custom(func() {
				size := imgui.ContentRegionAvail()
				aspectRatio := float32(ws.width) / float32(ws.height)

				giu.Group().Layout(
					giu.Image(ws.outputTex).Size(size.X, size.X/aspectRatio),
					giu.Label("Its possible your image is just transparent"),
				).Build()
			}),
		),
	).IsOpen(&ws.amOpen).Build()
}
func (ws *Workspace) buildParameterInputs() {
	//fmt.Println("Making",len(ws.args), "args")
	open := imgui.TreeNodeV("Paramaters", imgui.TreeNodeFlagsFramed)
	if open {
		imgui.Text("Right-click for more information")

		for _, dh := range ws.args {

			dh.Build(ws)

		}
		imgui.TreePop()
	}

}
