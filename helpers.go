package napoleon

import "os"

func (n *Napoleon) CreateDirIfNotExist(path string) error {
	const mode = 0755

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, mode)

		if err != nil {
			return err
		}
	}

	return nil
}

func (n *Napoleon) CreateFileIfNotExist(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)

		if err != nil {
			return err
		}

		defer func(file *os.File) {
			_ = file.Close()
		}(file)
	}
	return nil
}
