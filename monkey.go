package monkey

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

// patch is an applied patch
// needed to undo a patch
type patch struct {
	originalBytes      []byte
	replacement        *reflect.Value
	aliasPatchedPos    uintptr
	aliasOriginalBytes []byte
	addr               *uintptr
}

var (
	lock = sync.Mutex{}

	patches = make(map[uintptr]patch)
)

type value struct {
	_   uintptr
	ptr unsafe.Pointer
}

func getPtr(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).ptr
}

type PatchGuard struct {
	target      reflect.Value
	replacement reflect.Value
	alias       *reflect.Value // Use this interface to access the original target
}

func (g *PatchGuard) Unpatch() {
	unpatchValue(g.target)
}

func (g *PatchGuard) Restore() {
	patchValue(g.target, g.replacement, g.alias)
}

func Patch(target, replacement interface{}) *PatchGuard {
	t := reflect.ValueOf(target)
	r := reflect.ValueOf(replacement)
	patchValue(t, r, nil)

	return &PatchGuard{t, r, nil}
}

// Patch replaces a function with another
// alias: A wrapper, to access the original target when patched
func PatchEx(target, alias, replacement interface{}) *PatchGuard {
	t := reflect.ValueOf(target)
	r := reflect.ValueOf(replacement)
	a := reflect.ValueOf(alias)
	patchValue(t, r, &a)

	return &PatchGuard{t, r, &a}
}

// PatchInstanceMethod replaces an instance method methodName for the type target with replacement
// Replacement should expect the receiver (of type target) as the first argument
func PatchInstanceMethod(target reflect.Type, methodName string, replacement interface{}) *PatchGuard {
	m, ok := target.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", methodName))
	}
	r := reflect.ValueOf(replacement)
	patchValue(m.Func, r, nil)

	return &PatchGuard{m.Func, r, nil}
}

// PatchInstanceMethod replaces an instance method methodName for the type target with replacement
// Replacement should expect the receiver (of type target) as the first argument
func PatchInstanceMethodEx(target reflect.Type, methodName string, alias, replacement interface{}) *PatchGuard {
	m, ok := target.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", methodName))
	}
	r := reflect.ValueOf(replacement)
	a := reflect.ValueOf(alias)
	patchValue(m.Func, r, &a)

	return &PatchGuard{m.Func, r, &a}
}

func patchValue(target reflect.Value, replacement reflect.Value, alias *reflect.Value) {
	lock.Lock()
	defer lock.Unlock()

	if target.Kind() != reflect.Func {
		panic("target has to be a Func")
	}

	if replacement.Kind() != reflect.Func {
		panic("replacement has to be a Func")
	}

	if target.Type() != replacement.Type() {
		panic(fmt.Sprintf("target and replacement have to have the same type %s != %s", target.Type(), replacement.Type()))
	}

	if alias != nil {
		if alias.Kind() != reflect.Func {
			panic("alias has to be a Func")
		}

		if target.Type() != alias.Type() {
			panic(fmt.Sprintf("target and alias have to have the same type %s != %s", target.Type(), alias.Type()))
		}
	}

	if patch, ok := patches[target.Pointer()]; ok {
		unpatch(target.Pointer(), patch)
	}

	var addr *uintptr
	var aliasPos uintptr
	var aliasBytes []byte
	if alias != nil {
		targetOffset, aliasOffset, aliasOrininal := replaceJBE(target.Pointer(), (*alias).Pointer())

		addr = new(uintptr)
		*addr = *(*uintptr)(getPtr(target)) + targetOffset
		aliasPos = (*alias).Pointer() + aliasOffset
		orininalBytes := replaceFunction(aliasPos, (uintptr)(unsafe.Pointer(addr)))

		aliasBytes = make([]byte, len(aliasOrininal)+len(orininalBytes))
		copy(aliasBytes, aliasOrininal)

		capcity := len(aliasBytes)
		len1 := len(aliasOrininal)
		for i := len1; i < capcity; i++ {
			aliasBytes[i] = orininalBytes[i-len1]
		}
	}

	orininalBytes := replaceFunction(target.Pointer(), (uintptr)(getPtr(replacement)))
	patches[target.Pointer()] = patch{orininalBytes, &replacement, (*alias).Pointer(), aliasBytes, addr}
}

// Unpatch removes any monkey patches on target
// returns whether target was patched in the first place
func Unpatch(target interface{}) bool {
	return unpatchValue(reflect.ValueOf(target))
}

// UnpatchInstanceMethod removes the patch on methodName of the target
// returns whether it was patched in the first place
func UnpatchInstanceMethod(target reflect.Type, methodName string) bool {
	m, ok := target.MethodByName(methodName)
	if !ok {
		panic(fmt.Sprintf("unknown method %s", methodName))
	}
	return unpatchValue(m.Func)
}

// UnpatchAll removes all applied monkeypatches
func UnpatchAll() {
	lock.Lock()
	defer lock.Unlock()
	for target, p := range patches {
		unpatch(target, p)
		delete(patches, target)
	}
}

// Unpatch removes a monkeypatch from the specified function
// returns whether the function was patched in the first place
func unpatchValue(target reflect.Value) bool {
	lock.Lock()
	defer lock.Unlock()
	patch, ok := patches[target.Pointer()]
	if !ok {
		return false
	}
	unpatch(target.Pointer(), patch)
	delete(patches, target.Pointer())
	return true
}

func unpatch(target uintptr, p patch) {
	copyToLocation(target, p.originalBytes)
	if p.addr != nil {
		copyToLocation(p.aliasPatchedPos, p.aliasOriginalBytes)
	}
}
