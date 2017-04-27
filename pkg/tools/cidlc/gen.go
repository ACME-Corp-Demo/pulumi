// Copyright 2017 Pulumi, Inc. All rights reserved.

package cidlc

import (
	"bufio"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
)

// writefmt wraps the bufio.Writer.WriteString function, but also performs fmt.Sprintf-style formatting.
func writefmt(w *bufio.Writer, msg string, args ...interface{}) {
	w.WriteString(fmt.Sprintf(msg, args...))
}

// writefmtln wraps the bufio.Writer.WriteString function, performing fmt.Sprintf-style formatting and appending \n.
func writefmtln(w *bufio.Writer, msg string, args ...interface{}) {
	writefmt(w, msg+"\n", args...)
}

// emitHeaderWarning emits the standard "WARNING" into a generated file.
func emitHeaderWarning(w *bufio.Writer) {
	writefmtln(w, "// *** WARNING: this file was generated by the Coconut IDL Compiler (CIDLC).  ***")
	writefmtln(w, "// *** Do not edit by hand unless you are taking matters into your own hands! ***")
	writefmtln(w, "")
}

// mirrorDirLayout ensures a target output directory contains the same layout as the input package.
func mirrorDirLayout(pkg *Package, out string) error {
	for relpath := range pkg.Files {
		// Make the target file by concatening the output with the relative path, and ensure the directory exists.
		path := filepath.Join(out, relpath)
		if err := ensurePath(path); err != nil {
			return err
		}
	}
	return nil
}

// ensurePath ensures that a target filepath exists (like `mkdir -p`), returning a non-nil error if any problem occurs.
func ensurePath(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, 0755)
}

func forEachField(t TypeMember, action func(*types.Var, PropertyOptions)) int {
	return forEachStructField(t.Struct(), t.PropertyOptions(), action)
}

func forEachStructField(s *types.Struct, opts []PropertyOptions, action func(*types.Var, PropertyOptions)) int {
	n := 0
	for i, j := 0, 0; i < s.NumFields(); i++ {
		fld := s.Field(i)
		if fld.Anonymous() {
			// For anonymous types, recurse.
			named := fld.Type().(*types.Named)
			embedded := named.Underlying().(*types.Struct)
			k := forEachStructField(embedded, opts[j:], action)
			j += k
			n += k
		} else {
			// For actual fields, invoke the action, and bump the counters.
			if action != nil {
				action(s.Field(i), opts[j])
			}
			j++
			n++
		}
	}
	return n
}
