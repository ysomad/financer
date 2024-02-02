package guid

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

func New(prefix string) string {
	s := uuid.New().String()
	s = strings.ReplaceAll(s, "-", "")
	return fmt.Sprintf("%s_%s", prefix, s)
}
