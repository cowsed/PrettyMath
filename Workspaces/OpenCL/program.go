package opencl_renderer

//Holds all the information needed for a program/kernel
import (
	"fmt"
	"strings"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"

	"github.com/jgillich/go-opencl/cl"
)

type CLProgram struct {
	current bool

	programSource string
	programName   string

	programArgs []ClDataHolder

	editor     imgui.TextEditor
	errMarkers imgui.ErrorMarkers

	errorsCL string
	kernelCL *cl.Kernel

	//ws *Workspace
}

func (p *CLProgram) Build(ws *Workspace) {
	giu.TabItem(p.programName+" Program").Layout(
		giu.Custom(func() { p.buildParameterInputs(ws) }),
		giu.Custom(func() {
			p.editor.Render("OpenCl", imgui.Vec2{X: 0, Y: 0}, true)
			if p.editor.IsTextChanged() {
				p.programSource = p.editor.GetText()
				p.current = false
				ws.programsCurrent = false
				//p.ws.checkPrograms()
				fmt.Println("Update Parameters")
				p.makeParameters(ws)
			}
		}),
	).Build()
}
func (p *CLProgram) getWorkGroupSizes(ws *Workspace) ([]int, []int) {
	local, err := p.kernelCL.WorkGroupSize(ws.deviceCL)
	check(err)

	global := int(ws.width * ws.height)
	d := int(ws.width) * int(ws.height) % local
	if d != 0 {
		global += local - d
	}

	return []int{global}, []int{local}
}
func (p *CLProgram) setArgs(ws *Workspace) {
	fmt.Println("Setting Args")
	for i, dh := range p.programArgs {
		err := dh.SetArg(i, p.kernelCL, ws)
		if err != nil {
			fmt.Println(i, "th arguement error", err.Error())
			ws.releaseOnFinish()
			break
		}
	}
}

func (p *CLProgram) BuildProgram(ws *Workspace) {

	p.programSource = p.editor.GetText()
	//Reset errors
	p.errorsCL = ""
	p.errMarkers = imgui.NewErrorMarkers()
	p.editor.SetErrorMarkers(p.errMarkers)

	program, err := ws.contextCL.CreateProgramWithSource([]string{p.programSource})
	if err != nil {
		fmt.Println("Failling hard")
	}
	fmt.Println("Succeed hard")

	err = program.BuildProgram(nil, "")
	if err != nil {
		fmt.Println("Failed toBuild program")
		p.errorsCL = err.Error()

		p.errMarkers = parseCLErrors(p.errorsCL)
		p.editor.SetErrorMarkers(p.errMarkers)
		ws.releaseOnFinish()

		fmt.Println(err.Error())
		return
	}
	fmt.Println("Built program")

	kernel, err := program.CreateKernel(p.programName)
	check(err)
	if err != nil {
		fmt.Println("Kernel Error", err.Error())
	}
	numArgs, err := kernel.NumArgs()
	check(err)
	fmt.Printf("%v arguements detected", numArgs)

	var name string
	for i := 0; i < numArgs; i++ {
		name, err = kernel.ArgName(i)
		if err == cl.ErrUnsupported {
			break
		} else if err != nil {
			fmt.Printf("GetKernelArgInfo for name:%v failed: %+v", name, err)
			break
		} else {
			fmt.Printf("Kernel arg %d: %s \n", i, name)
		}
	}
	if err != nil {
		ws.releaseOnFinish()
		fmt.Println(err.Error())
		panic(err)
		//return
	}

	p.kernelCL = kernel
	p.current = true

}

//Setup Text Editor
func (p *CLProgram) initEditor(ws *Workspace) {

	p.errMarkers = imgui.NewErrorMarkers()

	p.editor = imgui.NewTextEditor()
	p.editor.SetErrorMarkers(p.errMarkers)

	p.editor.SetLanguageDefinitionC()
	p.editor.SetText(p.programSource)
	p.editor.SetShowWhitespaces(false)
	p.editor.SetTabSize(2)
}

func (p *CLProgram) buildParameterInputs(ws *Workspace) {
	open := imgui.TreeNodeV("Paramaters", imgui.TreeNodeFlagsFramed)
	if open {
		imgui.Text("Right-click for more information")

		for _, dh := range p.programArgs {
			dh.Build(ws)
		}
		imgui.TreePop()
	}

}

func (p *CLProgram) makeParameters(ws *Workspace) {
	fmt.Printf("My name is %p\n", p)

	names, types, err := findNamesAndTypes(p.programName, p.programSource)
	fmt.Println("names:", names)
	if err != nil {
		return
	}
	if len(names) != len(types) {
		return
	}
	oldArgs := p.programArgs[:]
	p.programArgs = make([]ClDataHolder, 0)
	bufferI := 0
	for i, t := range types {
		//Clean up stuff
		parts := strings.Split(strings.TrimSpace(t), " ")
		//fmt.Println("Parts: ", parts)
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
				newArg = &CLUint32Input{0, names[i], 0, 100, actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)
		case "int":
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLInt32Input{0, names[i], 0, 100, actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)

		case "float":
			//fmt.Println("Found: const float")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatInput{0, -100, 100, 0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)
		case "float3":
			//fmt.Println("Found: const float3")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatVecInput{[]float32{0.0, 0.0, 0.0}, 3, -100, 100, 0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)
		case "float2":
			//fmt.Println("Found: const float2")
			var newArg ClDataHolder
			if oldArgAvailable && oldArgs[i].getType() == actualType {
				newArg = oldArgs[i]
			} else {
				newArg = &CLFloatVecInput{[]float32{0.0, 0.0}, 2, -100, 100, 0.001, names[i], actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)

		case "image2d_t":
			fmt.Println("UNDEFINED BEHAVIOUR sorta")
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
				newArg = &CLImageInput{"imageThing", bufferI, names[i], false, 800, 600, actualType}
			}
			newArg.setName(names[i])
			p.programArgs = append(p.programArgs, newArg)
			bufferI++

		}

	}
	fmt.Println("there are", len(p.programArgs), "args made")
}
