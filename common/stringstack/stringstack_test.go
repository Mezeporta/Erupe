package stringstack

import (
	"testing"
)

func TestNew(t *testing.T) {
	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if len(s.stack) != 0 {
		t.Errorf("New() stack length = %d, want 0", len(s.stack))
	}
}

func TestStringStack_Set(t *testing.T) {
	s := New()
	s.Set("first")

	if len(s.stack) != 1 {
		t.Errorf("Set() stack length = %d, want 1", len(s.stack))
	}
	if s.stack[0] != "first" {
		t.Errorf("stack[0] = %q, want %q", s.stack[0], "first")
	}
}

func TestStringStack_Set_Replaces(t *testing.T) {
	s := New()
	s.Push("item1")
	s.Push("item2")
	s.Push("item3")

	// Set should replace the entire stack
	s.Set("new_item")

	if len(s.stack) != 1 {
		t.Errorf("Set() stack length = %d, want 1", len(s.stack))
	}
	if s.stack[0] != "new_item" {
		t.Errorf("stack[0] = %q, want %q", s.stack[0], "new_item")
	}
}

func TestStringStack_Push(t *testing.T) {
	s := New()
	s.Push("first")
	s.Push("second")
	s.Push("third")

	if len(s.stack) != 3 {
		t.Errorf("Push() stack length = %d, want 3", len(s.stack))
	}
	if s.stack[0] != "first" {
		t.Errorf("stack[0] = %q, want %q", s.stack[0], "first")
	}
	if s.stack[1] != "second" {
		t.Errorf("stack[1] = %q, want %q", s.stack[1], "second")
	}
	if s.stack[2] != "third" {
		t.Errorf("stack[2] = %q, want %q", s.stack[2], "third")
	}
}

func TestStringStack_Pop(t *testing.T) {
	s := New()
	s.Push("first")
	s.Push("second")
	s.Push("third")

	// Pop should return LIFO (last in, first out)
	val, err := s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v, want nil", err)
	}
	if val != "third" {
		t.Errorf("Pop() = %q, want %q", val, "third")
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v, want nil", err)
	}
	if val != "second" {
		t.Errorf("Pop() = %q, want %q", val, "second")
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v, want nil", err)
	}
	if val != "first" {
		t.Errorf("Pop() = %q, want %q", val, "first")
	}

	if len(s.stack) != 0 {
		t.Errorf("stack length = %d, want 0 after popping all items", len(s.stack))
	}
}

func TestStringStack_Pop_Empty(t *testing.T) {
	s := New()

	val, err := s.Pop()
	if err == nil {
		t.Error("Pop() on empty stack should return error")
	}
	if val != "" {
		t.Errorf("Pop() on empty stack returned %q, want empty string", val)
	}

	expectedError := "no items on stack"
	if err.Error() != expectedError {
		t.Errorf("Pop() error = %q, want %q", err.Error(), expectedError)
	}
}

func TestStringStack_LIFO_Behavior(t *testing.T) {
	s := New()
	items := []string{"A", "B", "C", "D", "E"}

	for _, item := range items {
		s.Push(item)
	}

	// Pop should return in reverse order (LIFO)
	for i := len(items) - 1; i >= 0; i-- {
		val, err := s.Pop()
		if err != nil {
			t.Fatalf("Pop() error = %v", err)
		}
		if val != items[i] {
			t.Errorf("Pop() = %q, want %q", val, items[i])
		}
	}
}

func TestStringStack_PushAfterPop(t *testing.T) {
	s := New()
	s.Push("first")
	s.Push("second")

	val, _ := s.Pop()
	if val != "second" {
		t.Errorf("Pop() = %q, want %q", val, "second")
	}

	s.Push("third")

	val, _ = s.Pop()
	if val != "third" {
		t.Errorf("Pop() = %q, want %q", val, "third")
	}

	val, _ = s.Pop()
	if val != "first" {
		t.Errorf("Pop() = %q, want %q", val, "first")
	}
}

func TestStringStack_EmptyStrings(t *testing.T) {
	s := New()
	s.Push("")
	s.Push("text")
	s.Push("")

	val, err := s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v", err)
	}
	if val != "" {
		t.Errorf("Pop() = %q, want empty string", val)
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v", err)
	}
	if val != "text" {
		t.Errorf("Pop() = %q, want %q", val, "text")
	}

	val, err = s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v", err)
	}
	if val != "" {
		t.Errorf("Pop() = %q, want empty string", val)
	}
}

func TestStringStack_LongStrings(t *testing.T) {
	s := New()
	longString := ""
	for i := 0; i < 1000; i++ {
		longString += "A"
	}

	s.Push(longString)
	val, err := s.Pop()

	if err != nil {
		t.Errorf("Pop() error = %v", err)
	}
	if val != longString {
		t.Error("Pop() returned different string than pushed")
	}
	if len(val) != 1000 {
		t.Errorf("Pop() string length = %d, want 1000", len(val))
	}
}

func TestStringStack_ManyItems(t *testing.T) {
	s := New()
	count := 1000

	// Push many items
	for i := 0; i < count; i++ {
		s.Push("item")
	}

	if len(s.stack) != count {
		t.Errorf("stack length = %d, want %d", len(s.stack), count)
	}

	// Pop all items
	for i := 0; i < count; i++ {
		_, err := s.Pop()
		if err != nil {
			t.Errorf("Pop()[%d] error = %v", i, err)
		}
	}

	// Should be empty now
	if len(s.stack) != 0 {
		t.Errorf("stack length = %d, want 0 after popping all", len(s.stack))
	}

	// Next pop should error
	_, err := s.Pop()
	if err == nil {
		t.Error("Pop() on empty stack should return error")
	}
}

func TestStringStack_SetAfterOperations(t *testing.T) {
	s := New()
	s.Push("a")
	s.Push("b")
	s.Push("c")
	s.Pop()
	s.Push("d")

	// Set should clear everything
	s.Set("reset")

	if len(s.stack) != 1 {
		t.Errorf("stack length = %d, want 1 after Set", len(s.stack))
	}

	val, err := s.Pop()
	if err != nil {
		t.Errorf("Pop() error = %v", err)
	}
	if val != "reset" {
		t.Errorf("Pop() = %q, want %q", val, "reset")
	}
}

func TestStringStack_SpecialCharacters(t *testing.T) {
	s := New()
	specialStrings := []string{
		"Hello\nWorld",
		"Tab\tSeparated",
		"Quote\"Test",
		"Backslash\\Test",
		"Unicode: テスト",
		"Emoji: 😀",
		"",
		" ",
		"  spaces  ",
	}

	for _, str := range specialStrings {
		s.Push(str)
	}

	// Pop in reverse order
	for i := len(specialStrings) - 1; i >= 0; i-- {
		val, err := s.Pop()
		if err != nil {
			t.Errorf("Pop() error = %v", err)
		}
		if val != specialStrings[i] {
			t.Errorf("Pop() = %q, want %q", val, specialStrings[i])
		}
	}
}

func BenchmarkStringStack_Push(b *testing.B) {
	s := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push("test string")
	}
}

func BenchmarkStringStack_Pop(b *testing.B) {
	s := New()
	// Pre-populate
	for i := 0; i < 10000; i++ {
		s.Push("test string")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if len(s.stack) == 0 {
			// Repopulate
			for j := 0; j < 10000; j++ {
				s.Push("test string")
			}
		}
		_, _ = s.Pop()
	}
}

func BenchmarkStringStack_PushPop(b *testing.B) {
	s := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push("test")
		_, _ = s.Pop()
	}
}

func BenchmarkStringStack_Set(b *testing.B) {
	s := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Set("test string")
	}
}
