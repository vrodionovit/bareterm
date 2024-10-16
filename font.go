package main

import (
	// Стандартные библиотеки Go
	"fmt"
	"image"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	// Сторонние библиотеки
	"github.com/go-gl/gl/v3.3-core/gl"    // OpenGL библиотека
	"github.com/golang/freetype/truetype" // Для работы с TrueType шрифтами
	"golang.org/x/image/font"             // Интерфейсы для работы со шрифтами
	"golang.org/x/image/math/fixed"       // Для работы с фиксированной точкой
)

// Font представляет собой структуру для хранения информации о шрифте
type Font struct {
	face     font.Face       // Интерфейс для отрисовки глифов
	textures map[rune]uint32 // Кэш текстур для каждого символа
	size     int             // Размер шрифта
	path     string          // Путь к файлу шрифта
}

// NewFont создает новый экземпляр Font
func NewFont(fontName string, size int) (*Font, error) {
	// Поиск TTF файла шрифта
	fontPath, err := findTTFFont(fontName)
	if err != nil {
		return nil, fmt.Errorf("failed to find font: %v", err)
	}

	// Чтение файла шрифта
	fontBytes, err := ioutil.ReadFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %v", err)
	}

	// Парсинг TTF данных
	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	// Создание face для заданного размера шрифта
	face := truetype.NewFace(f, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	font := &Font{
		face:     face,
		textures: make(map[rune]uint32),
		size:     size,
		path:     fontPath,
	}

	// Прогрев кэша для часто используемых символов
	commonChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789.,!?-+/():;%&*"
	font.WarmupCache(commonChars)

	return font, nil
}

// findTTFFont ищет TTF файл шрифта в стандартных директориях
func findTTFFont(fontName string) (string, error) {
	// Список стандартных директорий для поиска шрифтов
	fontDirs := getFontDirs()
	var fontPath string

	// Поиск шрифта в каждой директории
	for _, dir := range fontDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Проверка расширения файла и имени
			if filepath.Ext(path) == ".ttf" && strings.Contains(strings.ToLower(info.Name()), strings.ToLower(fontName)) {
				fontPath = path
				return filepath.SkipAll // Прекращаем поиск, если нашли подходящий файл
			}
			return nil
		})
		if err == nil && fontPath != "" {
			return fontPath, nil
		}
	}
	return "", fmt.Errorf("font %s not found", fontName)
}

// GetCharTexture возвращает текстуру для заданного символа
func (f *Font) GetCharTexture(char rune) uint32 {
	// Проверка наличия текстуры в кэше
	if texture, ok := f.textures[char]; ok {
		return texture
	}

	// Получение границ глифа
	bounds, advance, ok := f.face.GlyphBounds(char)
	if !ok {
		fmt.Printf("Warning: Glyph not found for character %c (code %d)\n", char, char)
		return 0
	}

	// Вычисление размеров изображения для глифа
	width := int(advance.Round())
	height := bounds.Max.Y.Round() - bounds.Min.Y.Round()

	if width <= 0 || height <= 0 {
		fmt.Printf("Warning: Invalid dimensions for character %c: %dx%d\n", char, width, height)
		return 0
	}

	// Создание изображения для глифа
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: f.face,
		Dot:  fixed.Point26_6{X: -bounds.Min.X, Y: -bounds.Min.Y},
	}
	d.DrawString(string(char))

	if len(img.Pix) == 0 {
		fmt.Printf("Warning: Empty image for character %c\n", char)
		return 0
	}

	// Создание OpenGL текстуры из изображения
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	// Сохранение текстуры в кэше
	f.textures[char] = texture
	return texture
}

// WarmupCache предварительно создает текстуры для заданного набора символов
func (f *Font) WarmupCache(chars string) {
	for _, char := range chars {
		f.GetCharTexture(char)
	}
}

// Destroy освобождает ресурсы, связанные с шрифтом
func (f *Font) Destroy() {
	for _, texture := range f.textures {
		gl.DeleteTextures(1, &texture)
	}
}

func getFontDirs() []string {
	var fontDirs []string

	if runtime.GOOS == "windows" {
		// Пути для Windows
		windir := os.Getenv("WINDIR")
		fontDirs = []string{
			filepath.Join(windir, "Fonts"),
			filepath.Join(os.Getenv("LOCALAPPDATA"), "Microsoft", "Windows", "Fonts"),
			filepath.Join(os.Getenv("APPDATA"), "Local", "Microsoft", "Windows", "Fonts"),
		}
	} else {
		// Пути для Unix-подобных систем
		home := os.Getenv("HOME")
		fontDirs = []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(home, ".local/share/fonts"),
			filepath.Join(home, ".fonts"),
		}
	}

	return fontDirs
}
