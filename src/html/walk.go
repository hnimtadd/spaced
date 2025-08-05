package html

import (
	"errors"

	"golang.org/x/net/html"
)

type WalkFunc func(node *html.Node) error

var ErrWalkSkip error = errors.New("skip node")

func Walk(node *html.Node, walkFn WalkFunc) error {
	if err := walkFn(node); err != nil {
		if errors.Is(err, ErrWalkSkip) {
			return nil
		}
		return err
	}

	for child := range node.ChildNodes() {
		if err := Walk(child, walkFn); err != nil {
			return err
		}
	}
	return nil
}
