// +build windows

package list_entitlements

type WindowsListEntitlements struct {
	Successful  bool
	Message string
	CodeSign int
}

func listEntitlements(pid int) (WindowsListEntitlements, error) {
	res := WindowsListEntitlements{}
    res.Successful = false
    res.Message = "Not Supported"
	return res, nil
}
func listCodeSign(pid int) (WindowsListEntitlements, error) {
	res := WindowsListEntitlements{}
    res.Successful = false
    res.Message = "Not Supported"
    res.CodeSign = -1;
	return res, nil
}
