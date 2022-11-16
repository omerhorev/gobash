package gobash

import "os"

func defaultCdFunc(p string) (string, error) {
	if err := os.Chdir(p); err != nil {
		return "", err
	} else {
		if dir, err := os.Getwd(); err != nil {
			return p, err
		} else {
			return dir, nil
		}
	}
}
