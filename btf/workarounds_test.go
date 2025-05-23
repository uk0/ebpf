package btf

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cilium/ebpf/internal"
	"github.com/cilium/ebpf/internal/testutils"

	"github.com/go-quicktest/qt"
)

func TestDatasecResolveWorkaround(t *testing.T) {
	testutils.SkipOnOldKernel(t, "5.2", "BTF_KIND_DATASEC")

	i := &Int{Size: 1}

	for _, typ := range []Type{
		&Typedef{"foo", i, nil},
		&Volatile{i},
		&Const{i},
		&Restrict{i},
		&TypeTag{i, "foo"},
	} {
		t.Run(fmt.Sprint(typ), func(t *testing.T) {
			if _, ok := typ.(*TypeTag); ok {
				testutils.SkipOnOldKernel(t, "5.17", "BTF_KIND_TYPE_TAG")
			}

			ds := &Datasec{
				Name: "a",
				Size: 2,
				Vars: []VarSecinfo{
					{
						Size:   1,
						Offset: 0,
						// struct, union, pointer, array will trigger the bug.
						Type: &Var{Name: "a", Type: &Pointer{i}},
					},
					{
						Size:   1,
						Offset: 1,
						Type: &Var{
							Name: "b",
							Type: typ,
						},
					},
				},
			}

			b, err := NewBuilder([]Type{ds})
			if err != nil {
				t.Fatal(err)
			}

			h, err := NewHandle(b)
			testutils.SkipIfNotSupportedOnOS(t, err)
			var ve *internal.VerifierError
			if errors.As(err, &ve) {
				t.Fatalf("%+v\n", ve)
			}
			if err != nil {
				t.Fatal(err)
			}
			h.Close()
		})
	}
}

func TestEmptyBTFWithStringTableWorkaround(t *testing.T) {
	var b Builder

	_, err := b.addString("foo")
	qt.Assert(t, qt.IsNil(err))

	h, err := NewHandle(&b)
	testutils.SkipIfNotSupported(t, err)
	qt.Assert(t, qt.IsNil(err))
	qt.Assert(t, qt.IsNil(h.Close()))
}
