import (
	// ... существующие импорты ...
	"os"
	"path/filepath"
)

func findTTFFont(fontName string) (string, error) {
	fontDirs := []string{
		"/usr/share/fonts",
		"/usr/local/share/fonts",
		filepath.Join(os.Getenv("HOME"), ".local/share/fonts"),
		filepath.Join(os.Getenv("HOME"), ".fonts"),
	}

	for _, dir := range fontDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if filepath.Ext(path) == ".ttf" && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(fontName)) {
				fontPath = path
				return filepath.SkipAll
			}
			return nil
		})
		if err == nil && fontPath != "" {
			return fontPath, nil
		}
	}
	return "", fmt.Errorf("font %s not found", fontName)
}

func NewFont(fontName string, size int) (*Font, error) {
	fontPath, err := findTTFFont(fontName)
	if err != nil {
		return nil, fmt.Errorf("failed to find font: %v", err)
	}

	fontBytes, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return &Font{
		face:     face,
		textures: make(map[rune]uint32),
		size:     size,
		path:     fontPath,
	}, nil
}