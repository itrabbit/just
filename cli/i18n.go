package cli

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/itrabbit/just"
)

var (
	rxTransText      = regexp.MustCompile(`\.(Tr|Trans)\(\"([^\")]+)\"`)
	rxTransAddFunc   = regexp.MustCompile(`\.AddTranslationMap\(\"(.+)\"`)
	rxTransMapValues = regexp.MustCompile(`\"(.+)\":?(\t+|\s+|)\"(.+)\"`)
)

func getTranslationMap(data []byte) just.TranslationMap {
	var res just.TranslationMap = nil
	if matches := rxTransMapValues.FindAllSubmatch(data, -1); matches != nil && len(matches) > 0 {
		for _, match := range matches {
			if len(match) > 3 {
				if res == nil {
					res = make(just.TranslationMap)
				}
				res[string(match[1])] = string(match[3])
			}
		}
	}
	return res
}

func getOutFileMap(filePath string) map[string]just.TranslationMap {
	res := make(map[string]just.TranslationMap)
	if info, err := os.Stat(filePath); (os.IsExist(err) || err == nil) && !info.IsDir() {
		if data, _ := ioutil.ReadFile(filePath); data != nil && len(data) > 0 {
			if indexes := rxTransAddFunc.FindAllSubmatchIndex(data, -1); indexes != nil && len(indexes) > 0 {
				lang, start, end := "", 0, 0
				for _, index := range indexes {
					start = end
					if len(index) == 4 {
						end = index[0]
						if start > 0 && len(lang) > 0 {
							if m := getTranslationMap(data[start:end]); m != nil && len(m) > 0 {
								if _, ok := res[lang]; ok {
									for k, v := range m {
										res[lang][k] = v
									}
								} else {
									res[lang] = m
								}
							}
						}
						lang = strings.ToLower(strings.TrimSpace(string(data[index[2]:index[3]])))
					}
				}
				end = len(data) - 1
				if m := getTranslationMap(data[start:end]); m != nil && len(m) > 0 {
					if _, ok := res[lang]; ok {
						for k, v := range m {
							res[lang][k] = v
						}
					} else {
						res[lang] = m
					}
				}
			}
		}
	}
	return res
}

func init() {
	RegCmdHandler("i18n:build", func() error {
		var (
			lang        = flag.String("lang", "en,ru", "Languages")
			dirPath     = flag.String("dir", "", "DIR for search translation text")
			outFilePath = flag.String("out", "", "Output go file")
		)
		flag.CommandLine.Parse(os.Args[2:])
		if dirPath == nil || outFilePath == nil || len(*dirPath) < 1 || len(*outFilePath) < 2 {
			return fmt.Errorf("Too short command params")
		}
		if _, err := os.Stat(*dirPath); os.IsNotExist(err) {
			return err
		}
		fmt.Println("[INFO] Build i18n v0.0.1")
		fmt.Println("[INFO] Languages:", *lang)
		fmt.Println("[INFO] Start find text in *.go for translation in dir:")
		fmt.Println("\t", *dirPath)

		// Составляем список файлов
		files := make([]string, 0, 0)
		if err := filepath.Walk(*dirPath, func(path string, f os.FileInfo, err error) error {
			if !f.IsDir() && filepath.Ext(f.Name()) == ".go" {
				files = append(files, path)
			}
			return nil
		}); err != nil {
			return err
		}
		// Массив строк
		mapStrings := make(map[string]uint)

		// Парсим файлы
		if len(files) < 1 {
			return fmt.Errorf("GoLang files not found!")
		}
		fmt.Println("[INFO] Parsing files:")
		for _, filePath := range files {
			data, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Println("[ERROR]", err)
				continue
			}
			if len(data) > 0 {
				fmt.Print("\t ", filePath, ".")
				if matches := rxTransText.FindAllSubmatch(data, -1); matches != nil && len(matches) > 0 {
					for _, match := range matches {
						if len(match) > 2 {
							str := string(match[2])
							if _, ok := mapStrings[str]; !ok {
								mapStrings[str] = 1
							} else {
								mapStrings[str]++
							}
							fmt.Print(".")
						}
					}
				}
				fmt.Println(".")
			}
		}
		if len(mapStrings) < 1 {
			return fmt.Errorf("Translate text not found!")
		}
		// Проводим анализ выходного файла файла
		langMap := getOutFileMap(*outFilePath)

		// Вводим новые данные
		for _, l := range strings.Split(*lang, ",") {
			l = strings.ToLower(strings.TrimSpace(l))
			if len(l) > 0 {
				if _, ok := langMap[l]; !ok {
					langMap[l] = make(just.TranslationMap)
				}
				for str := range mapStrings {
					if _, ok := langMap[l][str]; !ok {
						langMap[l][str] = str
					}
				}
			}
		}
		// Генерируем ответ и сохраняем
		buf := &bytes.Buffer{}
		buf.WriteString("// The file is generated using the CLI JUST.\r\n" +
			"// Change only translation strings!\r\n" +
			"// Everything else can be removed when re-generating!\r\n" +
			"// - - - - - \r\n")

		buf.WriteString("// Last generated time: " + time.Now().Format(time.RFC1123) + "\r\n\r\n")

		buf.WriteString("package main\r\n")
		buf.WriteString("\r\n")
		buf.WriteString("import \"github.com/itrabbit/just\"\r\n")
		buf.WriteString("\r\n")

		buf.WriteString("func loadTranslations(t just.ITranslator) {\r\n")
		buf.WriteString("\tif t != nil {\r\n")
		for l, m := range langMap {
			if len(m) > 0 {
				buf.WriteString("\t\tt.AddTranslationMap(\"" + l + "\", just.TranslationMap{\r\n")
				for k, v := range m {
					buf.WriteString("\t\t\t\"" + k + "\": \"" + v + "\",\r\n")
				}
				buf.WriteString("\t\t})\r\n")
			}
		}
		buf.WriteString("\t}\r\n")
		buf.WriteString("}\r\n")
		defer buf.Reset()

		return ioutil.WriteFile(*outFilePath, buf.Bytes(), 0766)
	})
}
