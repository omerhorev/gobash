package command

import "io"

// Opens the file in path. if path is "-", use stdin wrapped in io.NopCloser
func fileOrStdin(path string, env *Env) (io.ReadCloser, error) {
	if path == "-" {
		return io.NopCloser(env.Stdin()), nil
	} else {
		return env.Open(path)
	}
}
