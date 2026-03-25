package paths

import "os"

func stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}
