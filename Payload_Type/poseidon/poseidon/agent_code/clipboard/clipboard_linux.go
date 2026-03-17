//go:build linux && (clipboard || debug)

package clipboard

func GetClipboard(readTypes []string) (string, error) {
	return "Not Implemented", nil
}
