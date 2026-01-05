package handlers

import "fmt"

const (
	ErrEmptyImage  = "Image field is empty"
	ErrEmptyAction = "Action isn't set"
)

func toWrapHandlersErrors(img string) error {
	return fmt.Errorf("%s", img)
}
