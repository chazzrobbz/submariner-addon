/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// lifted from k8s.io/kubernetes so we can add methods to types
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"

	kruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

var (
	functionDest = flag.StringP("func-dest", "f", "-", "Output for swagger functions; '-' means stdout (default)")
	typeSrc      = flag.StringP("type-src", "s", "", "From where we are going to read the types")
	verify       = flag.BoolP("verify", "v", false, "Verifies if the given type-src file has documentation for every type")
)

func main() {
	flag.Parse()

	if *typeSrc == "" {
		klog.Fatalf("Please define -s flag as it is the source file")
	}

	var funcOut io.Writer
	if *functionDest == "-" {
		funcOut = os.Stdout
	} else {
		file, err := os.Create(*functionDest)
		if err != nil {
			klog.Fatalf("Couldn't open %v: %v", *functionDest, err)
		}
		defer file.Close()
		funcOut = file
	}

	var docsForTypes []kruntime.KubeTypes
	if fi, err := os.Stat(*typeSrc); err == nil && !fi.IsDir() {
		docsForTypes = kruntime.ParseDocumentationFrom(*typeSrc)
	} else if err == nil && fi.IsDir() {
		klog.Fatalf("-s must be a valid file or file glob pattern, not a directory")
	} else {
		m, err := filepath.Glob(*typeSrc)
		if err != nil {
			klog.Fatalf("Couldn't search for files matching -s: %v", err)
		}
		if len(m) == 0 {
			klog.Fatalf("-s must be a valid file or file glob pattern")
		}
		for _, file := range m {
			docsForTypes = append(docsForTypes, kruntime.ParseDocumentationFrom(file)...)
		}
	}

	if *verify == true {
		rc, err := kruntime.VerifySwaggerDocsExist(docsForTypes, funcOut)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error in verification process: %s\n", err)
		}

		os.Exit(rc)
	}

	if docsForTypes != nil && len(docsForTypes) > 0 {
		if err := kruntime.WriteSwaggerDocFunc(docsForTypes, funcOut); err != nil {
			fmt.Fprintf(os.Stderr, "Error when writing swagger documentation functions: %s\n", err)
			os.Exit(-1)
		}
	}
}
