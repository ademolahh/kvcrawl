package server

import (
	"net/rpc"
	"slices"
	"sync"
	"testing"

	"github.com/ademolahh/kvcrawl/kv/shared"
)

func set(t *testing.T, kv *KV, key, value string) {
	t.Helper()
	var ok bool
	if err := kv.Set(&shared.SetArgs{Key: key, Value: value}, &ok); err != nil {
		t.Fatalf("Set(%q, %q): %v", key, value, err)
	}
	if !ok {
		t.Fatalf("Set(%q, %q) reported failure", key, value)
	}
}

func get(t *testing.T, kv *KV, key string) shared.GetReply {
	t.Helper()
	var reply shared.GetReply
	if err := kv.Get(shared.KeyArg{Key: key}, &reply); err != nil {
		t.Fatalf("Get(%q): %v", key, err)
	}
	return reply
}

func del(t *testing.T, kv *KV, key string) bool {
	t.Helper()
	var ok bool
	if err := kv.Delete(shared.KeyArg{Key: key}, &ok); err != nil {
		t.Fatalf("Delete(%q): %v", key, err)
	}
	return ok
}

func list(t *testing.T, kv *KV) []string {
	t.Helper()
	var reply shared.ListReply
	if err := kv.List(shared.EmptyArgs{}, &reply); err != nil {
		t.Fatalf("List: %v", err)
	}
	slices.Sort(reply.Values)
	return reply.Values
}

func TestSetThenGet(t *testing.T) {
	kv := NewKV()
	set(t, kv, "name", "ada")

	reply := get(t, kv, "name")
	if !reply.Ok {
		t.Errorf("Ok = false, want true for a key that was just set")
	}
	if reply.Value != "ada" {
		t.Errorf("Value = %q, want %q", reply.Value, "ada")
	}
}

func TestGetMissingKey(t *testing.T) {
	kv := NewKV()

	reply := get(t, kv, "absent")
	if reply.Ok {
		t.Errorf("Ok = true, want false for a key that was never set")
	}
	if reply.Value != "" {
		t.Errorf("Value = %q, want empty string", reply.Value)
	}
}

func TestGetEmptyValueIsPresent(t *testing.T) {
	kv := NewKV()
	set(t, kv, "blank", "")

	reply := get(t, kv, "blank")
	if !reply.Ok {
		t.Errorf("Ok = false, want true for a key holding an empty value")
	}
}

func TestSetOverwrites(t *testing.T) {
	kv := NewKV()
	set(t, kv, "name", "ada")
	set(t, kv, "name", "grace")

	if reply := get(t, kv, "name"); reply.Value != "grace" {
		t.Errorf("Value = %q, want %q after overwrite", reply.Value, "grace")
	}
	if got := list(t, kv); len(got) != 1 {
		t.Errorf("List = %v, want a single entry after overwriting the same key", got)
	}
}

func TestDeleteExistingKey(t *testing.T) {
	kv := NewKV()
	set(t, kv, "name", "ada")

	if !del(t, kv, "name") {
		t.Errorf("Delete reported false, want true for an existing key")
	}
	if reply := get(t, kv, "name"); reply.Ok {
		t.Errorf("Ok = true after delete, want false")
	}
}

func TestDeleteMissingKey(t *testing.T) {
	kv := NewKV()

	if del(t, kv, "absent") {
		t.Errorf("Delete reported true, want false for a key that was never set")
	}
}

func TestDeleteLeavesOtherKeys(t *testing.T) {
	kv := NewKV()
	set(t, kv, "a", "1")
	set(t, kv, "b", "2")

	del(t, kv, "a")

	if got, want := list(t, kv), []string{"2"}; !slices.Equal(got, want) {
		t.Errorf("List = %v, want %v", got, want)
	}
}

func TestListEmpty(t *testing.T) {
	kv := NewKV()

	if got := list(t, kv); len(got) != 0 {
		t.Errorf("List = %v, want empty", got)
	}
}

func TestListReturnsAllValues(t *testing.T) {
	kv := NewKV()
	set(t, kv, "a", "one")
	set(t, kv, "b", "two")
	set(t, kv, "c", "three")

	got := list(t, kv)
	want := []string{"one", "three", "two"} // sorted by the list helper
	if !slices.Equal(got, want) {
		t.Errorf("List = %v, want %v", got, want)
	}
}

func TestMethodsAreRPCRegisterable(t *testing.T) {
	if err := rpc.NewServer().Register(NewKV()); err != nil {
		t.Fatalf("Register: %v", err)
	}
}

func TestConcurrentAccess(t *testing.T) {
	kv := NewKV()
	const workers = 50

	var wg sync.WaitGroup
	for i := range workers {
		wg.Go(func() {
			key := string(rune('a' + i%26))

			var ok bool
			if err := kv.Set(&shared.SetArgs{Key: key, Value: "v"}, &ok); err != nil {
				t.Errorf("Set: %v", err)
			}

			var reply shared.GetReply
			if err := kv.Get(shared.KeyArg{Key: key}, &reply); err != nil {
				t.Errorf("Get: %v", err)
			}

			var listReply shared.ListReply
			if err := kv.List(shared.EmptyArgs{}, &listReply); err != nil {
				t.Errorf("List: %v", err)
			}

			var deleted bool
			if err := kv.Delete(shared.KeyArg{Key: key}, &deleted); err != nil {
				t.Errorf("Delete: %v", err)
			}
		})
	}
	wg.Wait()
}
