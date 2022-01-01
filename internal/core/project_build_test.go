package core

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
)

func TestVersion(t *testing.T) {
	fmt.Println(version.NewConstraint(">=0.4.22, <0.6.0"))
	fmt.Println(version.NewConstraint(">=0.8.0"))
}
