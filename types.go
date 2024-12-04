package main

import (
	"bytes"
	"fmt"
	"go/types"
	"golang.org/x/tools/go/packages"
	"io"
	"runtime/debug"
	"strings"
)

// prints the types in this package
// to invoke, run ./react-sqlite-go-template -t
func PrintType(w io.Writer, structure types.Type, level int) error {
	printf := func(format string, a ...any) {
		if _, err := fmt.Fprintf(w, format, a...); err != nil {
			panic(err)
		}
	}
	indent := func(level int) {
		printf("%s", strings.Repeat("\t", level))
	}

	switch t := structure.(type) {
	case *types.Struct:
		printf("{\n")
		for i := 0; i < t.NumFields(); i++ {
			v := t.Field(i)
			indent(level + 1)
			printf("%s: ", v.Name())
			if err := PrintType(w, v.Type(), level+1); err != nil {
				return err
			}
			printf(",\n")
		}
		indent(level)
		printf("}")
	case *types.Map:
		printf("{\n")
		indent(level + 1)
		if err := PrintType(w, t.Key(), level+1); err != nil {
			return err
		}
		printf(": ")
		if err := PrintType(w, t.Elem(), level+1); err != nil {
			return err
		}
		printf(",\n")
		indent(level)
		printf("}")
	case *types.Slice:
		printf("[\n")
		indent(level + 1)
		if err := PrintType(w, t.Elem(), level+1); err != nil {
			return err
		}
		printf("\n")
		indent(level)
		printf("]")
	case *types.Basic:
		if t.Info()&types.IsNumeric != 0 {
			printf("0")
		} else if t.Info()&types.IsBoolean != 0 {
			printf("false")
		} else if t.Info()&types.IsString != 0 {
			printf("\"\"")
		}
	case *types.Named:
		if t.String() == "time.Time" {
			printf("\"1970-01-01T00:00:00.000Z\"")
		} else {
			if err := PrintType(w, t.Underlying(), level); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown type")
	}
	return nil
}

func PrintStructs(w io.Writer) (n int, err error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		err = fmt.Errorf("no build info")
		return
	}

	conf := packages.Config{Mode: packages.NeedTypes}
	pkgs, err := packages.Load(&conf, info.Path)
	if err != nil {
		return
	}

	if len(pkgs) < 1 {
		err = fmt.Errorf("no results loading package %s", info.Path)
		return
	}

	fmt.Fprintf(w, "const DEFAULTS = {\n")
	buf := bytes.NewBuffer(nil)
	var m int
	for _, pkg := range pkgs {
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			child := scope.Lookup(name)
			structure := child.Type().Underlying()
			m, err = fmt.Fprintf(buf, "\t%s: ", name)
			n += m
			if err != nil {
				return
			}

			if err := PrintType(buf, structure, 1); err != nil {
				// no-op
			} else if _, err = buf.Write([]byte(",\n")); err != nil {
				return n, err
			} else {
				var i int64
				i, err = io.Copy(w, buf)
				n += int(i)
				if err != nil {
					return n, err
				}
			}
			buf.Reset()
		}
	}
	fmt.Fprintf(w, "};\n")
	return

}
