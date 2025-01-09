//go:build windows

package persist_loginitem

type PersistLoginItemWindows struct {
	Message string
}

func runCommand(name string, path string, global bool, list bool, remove bool) PersistLoginItemWindows {
	n := PersistLoginItemWindows{}
	n.Message = "Not Implemented"
	return n
}
