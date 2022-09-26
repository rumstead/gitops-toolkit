package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/rumstead/gitops-toolkit/pkg/config/v1alpha1"
)

func main() {
	generateV1Alpha1()
}

func generateV1Alpha1() {
	r := new(jsonschema.Reflector)
	r.ExpandedStruct = true
	s := r.Reflect(&v1alpha1.RequestClusters{})
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile("./pkg/config/v1alpha1/schema.json", data, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
