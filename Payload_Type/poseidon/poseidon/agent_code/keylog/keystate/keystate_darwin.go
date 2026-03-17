//go:build darwin && (keylog || debug)

package keystate

import "errors"

func keyLogger() error {
	return errors.New("Not implemented.")
}
