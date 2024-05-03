package types

import (
	"fmt"
)

type Specialization = map[Type]Type

func Instantiate(t Type, ctx Specialization, mod *Module) Type {

	var out Type

	// Non Generic Type, No Specialization Needed
	if !IsGeneric(t) {
		fmt.Println(t)
		return t
	}

	if ctx == nil {
		ctx = make(Specialization)
	}

	switch t := t.(type) {

	case *TypeParam:
		// Check If specialized
		typ, ok := ctx[t]

		// Should always be specialized otherwise there is a problem elsewhere
		if !ok {
			panic("should be specialized")
		}

		// return specialized type
		return typ
	case *DefinedType:
		// Specialize underlying type
		typ := NewSpecializedType(t, ctx, mod)

		return typ
	case *SpecializedType:
		typ := NewSpecializedType(t.InstanceOf, apply(t.Spec, ctx), mod)
		return typ
	case *FunctionSignature:
		typ := NewSpecializedFunctionSignature(t, ctx, mod)
		return typ
	case *SpecializedFunctionSignature:
		typ := NewSpecializedFunctionSignature(t.Signature, apply(t.Spec, ctx), mod)
		return typ
	case *Alias:
		return Instantiate(t.RHS, ctx, mod)
	case *Pointer:
		cT := t.PointerTo               // Type Pointing To
		uT := Instantiate(cT, ctx, mod) // Instantiate Type with Specialization Map
		out = NewPointer(uT)            // Create new pointer with specialized type
		return out                      // return updated pointer
	default:
		// unimplemented instantiation
		panic(fmt.Sprintf("cannot instantiate type %s", t))
	}
}

// TODO: Find better way
func cloneWithSpecialization(t Type, ctx Specialization, mod *Module) Type {
	switch parent := t.(type) {
	case *Basic:
		return t // Basic Types cannot be specialized
	case *Struct:
		// Collect Fields
		fields := []*Var{}
		for _, field := range parent.Fields {
			s := Instantiate(field.Type(), ctx, mod)
			spec := NewVar(field.Name(), s)
			fields = append(fields, spec)
		}
		return NewStruct(fields)
	case *Enum:
		variants := EnumVariants{}

		for _, variant := range parent.Variants {
			if len(variant.Fields) == 0 {
				variants = append(variants, NewEnumVariant(variant.Name, variant.Discriminant, nil))
				continue
			}

			fields := []*Var{}

			for _, f := range variant.Fields {
				v := NewVar(f.Name(), nil)
				s := Instantiate(f.Type(), ctx, mod)
				v.SetType(s)
				fields = append(fields, v)
			}

			variants = append(variants, NewEnumVariant(variant.Name, variant.Discriminant, fields))
		}

		return NewEnum(parent.Name, variants)
	default:
		panic(fmt.Sprintf("unhandled case %T", parent))
	}

}

func apply(s1, s2 Specialization) Specialization {
	s3 := make(Specialization)
	for k, v := range s1 {
		v2, ok := s2[v]
		if !ok {
			s3[k] = v
		} else {
			s3[k] = v2
		}
	}

	for k, v := range s2 {
		if _, ok := s3[k]; !ok {
			s3[k] = v // Add substitution from s2 if not already defined in s3
		}
	}
	return s3
}

func SpecializedSymbolName(symbol Symbol, args TypeList) string {
	o := symbol.SymbolName()
	for _, typ := range args {
		o += "::_G::" + typ.String()
	}
	return o
}
