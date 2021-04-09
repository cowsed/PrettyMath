package tools

import(
	"fmt"
	"github.com/AllenDang/giu/imgui"
)

func DragFloatN(label string, vec []float32, speed, min, max float32, format string) bool {
	value_changed := false
	//imgui.BeginGroup()
	//imgui.PushID(label)
	size := imgui.CalcItemWidth() / float32(len(vec)+1)
	for i, _ := range vec {
		imgui.PushItemWidth(size)
		id := fmt.Sprintf("%s-%d\n", label, i)
		imgui.PushID(id)
		if i > 0 {
			imgui.SameLine()
		}
		changed := imgui.DragFloatV("", &vec[i], speed, min, max, format, 0)
		value_changed = value_changed || changed
		imgui.PopID()
		imgui.PopItemWidth()
	}
	//imgui.PopID()

	imgui.SameLine()
	imgui.Text(label)

	//imgui.EndGroup()
	return value_changed
}
