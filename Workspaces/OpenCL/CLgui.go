package opencl_renderer

import (
	"fmt"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/cowsed/PrettyMath/Tools"
)

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

	buildStatus := " "
	if !ws.programsCurrent {
		buildStatus = "*"
	}

	fmt.Printf("program arglength is %d\n", len(ws.programs[0].programArgs))

	giu.TabItem("OpenCL Pipeline").Layout(

		giu.SplitLayout("MainSplit", giu.DirectionHorizontal, true, 700,
			giu.Group().Layout(
				giu.Line(
					giu.Button("Build"+buildStatus).OnClick(ws.BuildPrograms),
					giu.Button("Run").OnClick(ws.Run),
					giu.Button("Save Image").OnClick(ws.Save),
				),
				giu.TabBar("Programs").Layout(
					giu.Custom(func(){
						for i:=range(ws.programs){
							ws.programs[i].Build(ws)
						}
					}),
				),
			),
			giu.Custom(func() {
				size := imgui.ContentRegionAvail()
				aspectRatio := float32(ws.width) / float32(ws.height)

				giu.Group().Layout(
					giu.Image(ws.outputTex).Size(size.X, size.X/aspectRatio),
					giu.Label("Its possible your image is just transparent"),
					giu.Button("Copy to clipboard").OnClick(ws.copyImageBuffer),
				).Build()
			}),
		),
	).IsOpen(&ws.amOpen).Build()
}

func (ws *Workspace) copyImageBuffer() {
	tools.CopyImageRGBA(ws.images[0])
}

