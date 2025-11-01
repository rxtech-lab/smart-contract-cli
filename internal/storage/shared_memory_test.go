package storage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

type SharedMemoryTestSuite struct {
	suite.Suite
	storage SharedMemory
}

func (s *SharedMemoryTestSuite) SetupTest() {
	s.storage = NewSharedMemory()
}

func (s *SharedMemoryTestSuite) TearDownTest() {
	// Clean up after each test
	err := s.storage.Clear()
	s.Require().NoError(err)
}

// TestNewSharedMemory tests the constructor.
func (s *SharedMemoryTestSuite) TestNewSharedMemory() {
	storage := NewSharedMemory()
	s.NotNil(storage)

	keys, err := storage.List()
	s.NoError(err)
	s.Empty(keys)
}

// TestSetAndGet tests basic set and get operations.
func (s *SharedMemoryTestSuite) TestSetAndGet() {
	err := s.storage.Set("key1", "value1")
	s.NoError(err)

	value, err := s.storage.Get("key1")
	s.NoError(err)
	s.Equal("value1", value)
}

// TestGetNonExistentKey tests getting a key that doesn't exist.
func (s *SharedMemoryTestSuite) TestGetNonExistentKey() {
	value, err := s.storage.Get("nonexistent")
	s.NoError(err)
	s.Nil(value)
}

// TestSetMultipleKeys tests setting multiple keys.
func (s *SharedMemoryTestSuite) TestSetMultipleKeys() {
	err := s.storage.Set("key1", "value1")
	s.NoError(err)

	err = s.storage.Set("key2", 42)
	s.NoError(err)

	err = s.storage.Set("key3", true)
	s.NoError(err)

	value1, err := s.storage.Get("key1")
	s.NoError(err)
	s.Equal("value1", value1)

	value2, err := s.storage.Get("key2")
	s.NoError(err)
	s.Equal(42, value2)

	value3, err := s.storage.Get("key3")
	s.NoError(err)
	s.Equal(true, value3)
}

// TestSetOverwriteExistingKey tests overwriting an existing key.
func (s *SharedMemoryTestSuite) TestSetOverwriteExistingKey() {
	err := s.storage.Set("key1", "original")
	s.NoError(err)

	err = s.storage.Set("key1", "updated")
	s.NoError(err)

	value, err := s.storage.Get("key1")
	s.NoError(err)
	s.Equal("updated", value)
}

// TestDelete tests deleting a key.
func (s *SharedMemoryTestSuite) TestDelete() {
	err := s.storage.Set("key1", "value1")
	s.NoError(err)

	err = s.storage.Delete("key1")
	s.NoError(err)

	value, err := s.storage.Get("key1")
	s.NoError(err)
	s.Nil(value)
}

// TestDeleteNonExistentKey tests deleting a key that doesn't exist.
func (s *SharedMemoryTestSuite) TestDeleteNonExistentKey() {
	err := s.storage.Delete("nonexistent")
	s.NoError(err)
}

// TestList tests listing all keys.
func (s *SharedMemoryTestSuite) TestList() {
	err := s.storage.Set("key1", "value1")
	s.NoError(err)

	err = s.storage.Set("key2", "value2")
	s.NoError(err)

	err = s.storage.Set("key3", "value3")
	s.NoError(err)

	keys, err := s.storage.List()
	s.NoError(err)
	s.Len(keys, 3)
	s.Contains(keys, "key1")
	s.Contains(keys, "key2")
	s.Contains(keys, "key3")
}

// TestListEmpty tests listing when storage is empty.
func (s *SharedMemoryTestSuite) TestListEmpty() {
	keys, err := s.storage.List()
	s.NoError(err)
	s.Empty(keys)
}

// TestClear tests clearing all data.
func (s *SharedMemoryTestSuite) TestClear() {
	err := s.storage.Set("key1", "value1")
	s.NoError(err)

	err = s.storage.Set("key2", "value2")
	s.NoError(err)

	err = s.storage.Clear()
	s.NoError(err)

	keys, err := s.storage.List()
	s.NoError(err)
	s.Empty(keys)

	value, err := s.storage.Get("key1")
	s.NoError(err)
	s.Nil(value)
}

// TestClearEmpty tests clearing when already empty.
func (s *SharedMemoryTestSuite) TestClearEmpty() {
	err := s.storage.Clear()
	s.NoError(err)

	keys, err := s.storage.List()
	s.NoError(err)
	s.Empty(keys)
}

// TestDifferentValueTypes tests storing different types of values.
func (s *SharedMemoryTestSuite) TestDifferentValueTypes() {
	// String
	err := s.storage.Set("string", "hello")
	s.NoError(err)

	// Integer
	err = s.storage.Set("int", 42)
	s.NoError(err)

	// Boolean
	err = s.storage.Set("bool", true)
	s.NoError(err)

	// Float
	err = s.storage.Set("float", 3.14)
	s.NoError(err)

	// Slice
	err = s.storage.Set("slice", []string{"a", "b", "c"})
	s.NoError(err)

	// Map
	err = s.storage.Set("map", map[string]int{"one": 1, "two": 2})
	s.NoError(err)

	// Struct
	type TestStruct struct {
		Name string
		Age  int
	}
	err = s.storage.Set("struct", TestStruct{Name: "John", Age: 30})
	s.NoError(err)

	// Verify all values
	stringVal, _ := s.storage.Get("string")
	s.Equal("hello", stringVal)

	intVal, _ := s.storage.Get("int")
	s.Equal(42, intVal)

	boolVal, _ := s.storage.Get("bool")
	s.Equal(true, boolVal)

	floatVal, _ := s.storage.Get("float")
	s.Equal(3.14, floatVal)

	sliceVal, _ := s.storage.Get("slice")
	s.Equal([]string{"a", "b", "c"}, sliceVal)

	mapVal, _ := s.storage.Get("map")
	s.Equal(map[string]int{"one": 1, "two": 2}, mapVal)

	structVal, _ := s.storage.Get("struct")
	s.Equal(TestStruct{Name: "John", Age: 30}, structVal)
}

// TestConcurrentAccess tests concurrent read/write operations.
func (s *SharedMemoryTestSuite) TestConcurrentAccess() {
	const numGoroutines = 100
	const numOperations = 10

	var waitGroup sync.WaitGroup

	// Concurrent writes
	for index := 0; index < numGoroutines; index++ {
		waitGroup.Add(1)
		go func(id int) {
			defer waitGroup.Done()
			for j := 0; j < numOperations; j++ {
				key := "key_" + string(rune(id))
				value := id*numOperations + j
				err := s.storage.Set(key, value)
				s.NoError(err)
			}
		}(index)
	}

	waitGroup.Wait()

	// Verify some data was written
	keys, err := s.storage.List()
	s.NoError(err)
	s.NotEmpty(keys)
}

// TestConcurrentReadWrite tests concurrent reads and writes.
func (s *SharedMemoryTestSuite) TestConcurrentReadWrite() {
	// Pre-populate with some data
	for i := 0; i < 10; i++ {
		err := s.storage.Set("key_"+string(rune(i)), i)
		s.NoError(err)
	}

	const numGoroutines = 50
	var waitGroup sync.WaitGroup

	// Concurrent reads
	for index := 0; index < numGoroutines; index++ {
		waitGroup.Add(1)
		go func(id int) {
			defer waitGroup.Done()
			key := "key_" + string(rune(id%10))
			_, err := s.storage.Get(key)
			s.NoError(err)
		}(index)
	}

	// Concurrent writes
	for index := 0; index < numGoroutines; index++ {
		waitGroup.Add(1)
		go func(id int) {
			defer waitGroup.Done()
			key := "new_key_" + string(rune(id))
			err := s.storage.Set(key, id)
			s.NoError(err)
		}(index)
	}

	waitGroup.Wait()

	// Verify storage is still functional
	keys, err := s.storage.List()
	s.NoError(err)
	s.NotEmpty(keys)
}

// TestConcurrentDelete tests concurrent delete operations.
func (s *SharedMemoryTestSuite) TestConcurrentDelete() {
	// Pre-populate with data
	for i := 0; i < 100; i++ {
		err := s.storage.Set("key_"+string(rune(i)), i)
		s.NoError(err)
	}

	var waitGroup sync.WaitGroup

	// Concurrent deletes
	for index := 0; index < 100; index++ {
		waitGroup.Add(1)
		go func(id int) {
			defer waitGroup.Done()
			key := "key_" + string(rune(id))
			err := s.storage.Delete(key)
			s.NoError(err)
		}(index)
	}

	waitGroup.Wait()

	// Verify all keys are deleted
	keys, err := s.storage.List()
	s.NoError(err)
	s.Empty(keys)
}

// TestConcurrentList tests concurrent list operations.
func (s *SharedMemoryTestSuite) TestConcurrentList() {
	// Pre-populate with data
	for i := 0; i < 10; i++ {
		err := s.storage.Set("key_"+string(rune(i)), i)
		s.NoError(err)
	}

	var waitGroup sync.WaitGroup

	// Concurrent list operations
	for i := 0; i < 50; i++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			keys, err := s.storage.List()
			s.NoError(err)
			s.NotEmpty(keys)
		}()
	}

	waitGroup.Wait()
}

// TestNilValue tests storing nil values.
func (s *SharedMemoryTestSuite) TestNilValue() {
	err := s.storage.Set("nil_key", nil)
	s.NoError(err)

	value, err := s.storage.Get("nil_key")
	s.NoError(err)
	s.Nil(value)

	// Verify key exists in list
	keys, err := s.storage.List()
	s.NoError(err)
	s.Contains(keys, "nil_key")
}

// TestEmptyStringKey tests using empty string as key.
func (s *SharedMemoryTestSuite) TestEmptyStringKey() {
	err := s.storage.Set("", "value")
	s.NoError(err)

	value, err := s.storage.Get("")
	s.NoError(err)
	s.Equal("value", value)

	keys, err := s.storage.List()
	s.NoError(err)
	s.Contains(keys, "")
}

func TestSharedMemoryTestSuite(t *testing.T) {
	suite.Run(t, new(SharedMemoryTestSuite))
}
