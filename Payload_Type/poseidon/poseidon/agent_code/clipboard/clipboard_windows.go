//go:build windows

package clipboard

func GetClipboard(readTypes []string) (string, error) {
	return "Not Implemented", nil
}
