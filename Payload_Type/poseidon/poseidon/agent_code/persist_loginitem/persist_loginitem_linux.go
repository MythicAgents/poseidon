//go:build linux

package persist_loginitem

type PersistLoginItemLinux struct {
	Message string
}

func runCommand(name string, path string, global bool, list bool, remove bool) PersistLoginItemLinux {
	n := PersistLoginItemLinux{}
	n.Message = "Not Implemented"
	return n
}
