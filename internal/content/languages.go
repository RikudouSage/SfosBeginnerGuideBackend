package content

import "io/fs"

type LanguageProvider interface {
	List() ([]string, error)
}

type FSLocalizer struct {
	root fs.FS
	path string
}

func NewFSLocalizer(root fs.FS, path string) *FSLocalizer {
	return &FSLocalizer{root: root, path: path}
}

func (receiver *FSLocalizer) List() ([]string, error) {
	entries, err := fs.ReadDir(receiver.root, receiver.path)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		result = append(result, entry.Name())
	}

	return result, nil
}
