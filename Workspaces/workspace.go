//Package workspace holds assorted helper things for all workspaces
package workspace

//ProgressUpdate holds information for worker threads to communicate
type ProgressUpdate struct {
	Status string
	Description string
	Amount float64
}
