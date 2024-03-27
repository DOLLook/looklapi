// Package memstore contains a store which is just
// a map of key-value entries with immutability capabilities.
//
// Developers can use that storage to their own apps if they like its behavior.
// It's fast and in the same time you get read-only access (safety) when you need it.
package memstore

import (
	"errors"
	"fmt"
	"reflect"
)

type (
	// entry is the entry of the context storage Store - .Values()
	entry struct {
		Key       string
		ValueRaw  interface{}
		immutable bool // if true then it can't change by its caller.
	}

	// Store is a map of key-value entries with immutability capabilities.
	Store map[string]entry
)

// Value returns the value of the entry,
// respects the immutable.
func (e entry) Value() interface{} {
	if e.immutable {
		// take its value, no pointer even if set with a reference.
		vv := reflect.Indirect(reflect.ValueOf(e.ValueRaw))

		// return copy of that slice
		if vv.Type().Kind() == reflect.Slice {
			newSlice := reflect.MakeSlice(vv.Type(), vv.Len(), vv.Cap())
			reflect.Copy(newSlice, vv)
			return newSlice.Interface()
		}
		// return a copy of that map
		if vv.Type().Kind() == reflect.Map {
			newMap := reflect.MakeMap(vv.Type())
			for _, k := range vv.MapKeys() {
				newMap.SetMapIndex(k, vv.MapIndex(k))
			}
			return newMap.Interface()
		}
		// if was *value it will return value{}.
		return vv.Interface()
	}
	return e.ValueRaw
}

// Save same as `Set`
// However, if "immutable" is true then saves it as immutable (same as `SetImmutable`).
//
// Returns the entry and true if it was just inserted, meaning that
// it will return the entry and a false boolean if the entry exists and it has been updated.
func (r *Store) save(key string, value interface{}, immutable bool) bool {
	if e, has := (*r)[key]; has {
		// replace if we can, else just return
		if e.immutable {
			return false
		} else {
			e.ValueRaw = value
			e.immutable = immutable
			return true
		}
	} else {
		// add
		kv := entry{
			Key:       key,
			ValueRaw:  value,
			immutable: immutable,
		}
		(*r)[key] = kv

		return true
	}
}

// Set saves a value to the key-value storage.
// Returns true if it was just inserted or updated, meaning that
// it will false if the entry exists and it is immutable.
//
// See `SetImmutable` and `Get`.
func (r *Store) Set(key string, value interface{}) bool {
	return r.save(key, value, false)
}

// SetImmutable saves a value to the key-value storage.
// Unlike `Set`, the value set to store cannot be changed by the caller later on (when .Get OR .Set)
func (r *Store) SetImmutable(key string, value interface{}) bool {
	return r.save(key, value, true)
}

// GetDefault returns the entry's value based on its key.
// If not found returns "def".
func GetDefault[T any](store *Store, key string, def T) T {
	if store == nil {
		return def
	}

	if e, has := (*store)[key]; has {
		if v, ok := e.Value().(T); ok {
			return v
		} else {
			return def
		}
	} else {
		return def
	}
}

var NotFoundErr = errors.New("not found store value")

// GetDefault returns the entry's value based on its key.
// If not found returns "def".
func Get[T any](store *Store, key string) (T, error) {
	var val T
	if store == nil {
		return val, errors.New("invalid store nil")
	}

	if e, has := (*store)[key]; has {
		if v, ok := e.Value().(T); ok {
			return v, nil
		} else {
			return val, errors.New(fmt.Sprintf("store value with type %s not match the type %s", reflect.TypeOf(e.ValueRaw).String(), reflect.TypeOf(val).String()))
		}
	} else {
		return val, NotFoundErr
	}
}

// Exists is a small helper which reports whether a key exists.
// It's not recommended to be used outside of templates.
// Use Get or GetEntry instead which will give you back the entry value too,
// so you don't have to loop again the key-value storage to get its value.
func (r *Store) Exists(key string) bool {
	_, has := (*r)[key]
	return has
}

// Remove deletes an entry linked to that "key",
// returns true if an entry is actually removed.
func (r *Store) Remove(key string) {
	delete(*r, key)
}

// Reset clears all the request entries.
func (r *Store) Reset() {
	clear(*r)
}

// Len returns the full length of the entries.
func (r *Store) Len() int {
	args := *r
	return len(args)
}
