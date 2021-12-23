//Package workspace holds assorted helper things for all workspaces
package workspace

import (
	"fmt"

	tools "github.com/cowsed/PrettyMath/Tools"
)

//ProgressUpdate holds information for worker threads to communicate
type ProgressUpdate struct {
	Status      string
	Description string
	Amount      float64
}

type WorkspaceMakers struct {
	InitFunc func(onCloseFunc func(), AddProcessComm func() chan ProgressUpdate) tools.Workspace
	Name     string
}

var RegisteredWorkspaces []WorkspaceMakers

func RegisterWorkspace(InitFunc func(onCloseFunc func(), AddProcessComm func() chan ProgressUpdate) tools.Workspace, Name string) {
	ws := WorkspaceMakers{InitFunc: InitFunc, Name: Name}
	RegisteredWorkspaces = append(RegisteredWorkspaces, ws)
	fmt.Println("Registered confirmation of " + Name)
}
