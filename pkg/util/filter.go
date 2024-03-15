package util

import (
	"embed"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

//go:embed gfwlist_domain.txt
var f embed.FS
var Set sets.Set[string]

func init() {
	file, err := f.ReadFile("gfwlist_domain.txt")
	if err != nil {
		log.Fatal("Error reading file:", err)
	}
	split := strings.Split(string(file), "\n")
	Set = sets.New(split...)
}
