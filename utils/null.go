package utils

import "io"

// Null is a reader/writer that returns EOF to all reads and writes
var Null = null{}

type null struct{}

func (null) Read([]byte) (int, error)  { return 0, io.EOF }
func (null) Write([]byte) (int, error) { return 0, io.EOF }
